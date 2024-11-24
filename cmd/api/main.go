package main 

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tursodatabase/libsql-client-go/libsql"
	"database/sql"
)

type application struct {
	db *sql.DB
	logger *log.Logger
}

func(){
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	if := err godotenv.Load(); err != nil {
		logger.Fatal("error loading .env file")
	}

dbURL := os.Getenv("TURSO_DATABASE_URL")
dbToken := os.Getenv("TURSO_AUTH_TOKEN")

if dbUrl == "" || dbToken == "" {
	logger.Fatal("database url and auth token are not set in .env")
}

connString := dbUrl + "?authToken" + dbToken

//create database connection
db, err := sql.Open ("libsql", connString)
if err != nil{
	logger.Fatal("error opening database:::", err)
}
defer db.Close()

//connection pool settings
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)

//test database connection 
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.PingContext(ctx); err != nil {
	logger.Fatal("error connecting to the database")
}

logger.Printf("connected to the database")

//create application 
app := &application{
	db: db,
	logger: logger,
}

logger.Printf("starting server...")


}