CREATE DATABASE spotify_data_db;
USE spotify_data_db;
CREATE TABLE spotify_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255),
    displayname VARCHAR(255),
    followers VARCHAR(255),
    profileurl VARCHAR(255)
);
