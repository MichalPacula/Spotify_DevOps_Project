package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bmizerany/pat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

type UserData struct {
	Id          int
	Username    string
	DisplayName string
	Followers   string
	ProfileUrl  string
}

var (
	spotifyData *SpotifyData
	db          *sql.DB
	redisdb     *redis.Client
	err         error
)

type dataPageStruct struct {
	DbData    []UserData
	RedisData []map[string]string
}

var ctx = context.Background()

func main() {
	SetupDatabase()

	SetupRedis()

	SetupHttpServer()
}

func SetupHttpServer() {
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(index))
	mux.Post("/search", http.HandlerFunc(getData))
	mux.Get("/searchResult", http.HandlerFunc(searchResult))
	mux.Get("/data", http.HandlerFunc(dataPage))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	http.Handle("/", mux)

	log.Print("Listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func SetupDatabase() {
	db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(db:3306)/spotify_data_db", os.Getenv("MYSQL_ROOT_PASSWORD")))
	if err != nil {
		log.Fatal("Error connecting to db: ", err)
	}
	defer db.Close()
}

func SetupRedis() {
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/index.html", nil)
}

func getData(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")

	spotifyData = GetSpotifyData(username)

	insert, err := db.Query(fmt.Sprintf("INSERT INTO spotify_data(username, displayname, followers, profileurl) VALUES('%s', '%s', '%s', '%s')", spotifyData.Username, spotifyData.DisplayName, spotifyData.Followers, spotifyData.ProfileUrl))
	if err != nil {
		log.Fatal("Error inserting spotify data to db: ", err)
	}
	defer insert.Close()

	redisUserId, err := redisdb.Incr(ctx, "user:id").Result()
	if err != nil {
		log.Fatal("Error generating next redis user id: ", err)
	}

	userKey := fmt.Sprintf("user:%d", int(redisUserId))
	err = redisdb.HSet(ctx, userKey, map[string]interface{}{
		"id":          int(redisUserId),
		"username":    spotifyData.Username,
		"displayname": spotifyData.DisplayName,
		"followers":   spotifyData.Followers,
		"profileurl":  spotifyData.ProfileUrl,
	}).Err()
	if err != nil {
		log.Fatal("Error inserting user data to redis: ", err)
	}

	err = redisdb.RPush(ctx, "user_ids", int(redisUserId)).Err()
	if err != nil {
		log.Fatal("Error inserting user id to redis: ", err)
	}

	fmt.Println("Inserted data to redis")

	http.Redirect(w, r, "/searchResult", http.StatusSeeOther)
}

func searchResult(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/search_result.html", spotifyData)
}

func dataPage(w http.ResponseWriter, r *http.Request) {
	dbResults, err := db.Query("SELECT * FROM spotify_data ORDER BY id DESC LIMIT 5")

	var dbResultsArray []UserData
	if err != nil {
		log.Fatal("Error querying db for top 5 results: ", err)
	}
	for dbResults.Next() {
		var dbRow UserData
		err := dbResults.Scan(&dbRow.Id, &dbRow.Username, &dbRow.DisplayName, &dbRow.Followers, &dbRow.ProfileUrl)
		if err != nil {
			log.Fatal("Error assigning values from dbResults to dbRow: ", err)
		}

		dbResultsArray = append(dbResultsArray, dbRow)
	}

	latestUserIDs, err := redisdb.LRange(ctx, "user_ids", -5, -1).Result()
	if err != nil {
		log.Fatal("Error retrieving user ids from redis: ", err)
	}

	var redisResultsArray []map[string]string
	for _, idStr := range latestUserIDs {
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Fatal("Error converting user id to int: ", err)
		}

		userKey := fmt.Sprintf("user:%d", userID)
		userData, err := redisdb.HGetAll(ctx, userKey).Result()
		if err != nil {
			log.Fatal("Error retrieving user data from redis: ", err)
		}

		redisResultsArray = append(redisResultsArray, userData)
	}

	fmt.Println(redisResultsArray)
	fmt.Println("Retrievied data from redis")

	dataPageData := dataPageStruct{
		DbData:    dbResultsArray,
		RedisData: redisResultsArray,
	}

	render(w, "templates/data_page.html", dataPageData)
}

func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Print(err)
		http.Error(w, "Sorry something went wrong", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}
}
