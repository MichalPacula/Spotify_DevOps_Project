package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"database/sql"
	"os"

	"github.com/bmizerany/pat"
	_ "github.com/go-sql-driver/mysql"
)

type DbRow struct {
	Id int
	Username string
	DisplayName string
	Followers string
	ProfileUrl string
}

var spotifyData *SpotifyData
var db *sql.DB
var err error

func main() {
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(index))
	mux.Post("/search", http.HandlerFunc(getData))
	mux.Get("/searchResult", http.HandlerFunc(searchResult))
	mux.Get("/data", http.HandlerFunc(dataPage))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	http.Handle("/", mux)

	db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(db:3306)/spotify_data_db", os.Getenv("MYSQL_ROOT_PASSWORD")))
	if err != nil {
		log.Fatal("Error connecting to db: ", err)
	}
	defer db.Close()

	log.Print("Listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
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

	http.Redirect(w, r, "/searchResult", http.StatusSeeOther)
}

func searchResult(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/search_result.html", spotifyData)
}

func dataPage(w http.ResponseWriter, r *http.Request) {
	dbResults, err := db.Query("SELECT * FROM spotify_data ORDER BY id DESC LIMIT 5")

	var dbResultsArray []DbRow
	if err != nil {
		log.Fatal("Error querying db for top 5 results: ", err)
	}
	for dbResults.Next() {
		var dbRow DbRow
		err := dbResults.Scan(&dbRow.Id, &dbRow.Username, &dbRow.DisplayName, &dbRow.Followers, &dbRow.ProfileUrl)
		if err != nil {
			log.Fatal("Error assigning values from dbResults to dbRow: ", err)
		}
		dbResultsArray = append(dbResultsArray, dbRow)
	}
	render(w, "templates/data_page.html", dbResultsArray)
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
