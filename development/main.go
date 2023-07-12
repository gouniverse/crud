package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gouniverse/crud"
	"github.com/gouniverse/hb"
	"github.com/gouniverse/utils"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func crudHandler(w http.ResponseWriter, r *http.Request) {
	crudInstance, err := crud.NewCrud(crud.CrudConfig{
		Endpoint:           "/crud",
		HomeURL:            "/",
		EntityNameSingular: "User",
		EntityNamePlural:   "Users",
		ColumnNames:        []string{"First Name", "Last Name"},
		CreateFields: []crud.FormField{
			{
				Type:  "string",
				Label: "Name",
				Name:  "name",
			},
		},
		UpdateFields: []crud.FormField{
			{
				Type:  "string",
				Label: "First Name",
				Name:  "first_name",
			},
			{
				Type:  "string",
				Label: "Last Name",
				Name:  "last_name",
			},
		},
		FuncRows: func() ([]crud.Row, error) {
			return []crud.Row{
				{
					ID:   "ID1",
					Data: []string{"Jon", "Doe"},
				},
				{
					ID:   "ID2",
					Data: []string{"Sarah", "Smith"},
				},
				{
					ID:   "ID3",
					Data: []string{"Tom", "Sawyer"},
				},
			}, nil
		},
		FuncCreate: func(data map[string]string) (string, error) {
			// Logic for creating a new user
			return "ID4", nil
		},
		FuncUpdate: func(entityID string, data map[string]string) error {
			// Logic for updating an existing user
			return nil
		},
		FuncTrash: func(entityID string) error {
			// Logic for deleting an existing user
			return nil
		},
		FuncFetchUpdateData: func(entityID string) (map[string]string, error) {
			// Logic for fetching an existing user
			return map[string]string{
				"name": "Charles Dickens",
			}, nil
		},
	})

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	crudInstance.Handler(w, r)
}

func main() {
	log.Println("1. Initializing environment variables...")
	utils.EnvInitialize()

	log.Println("2. Initializing database...")
	var err error
	db, err = mainDb(utils.Env("DB_DRIVER"), utils.Env("DB_HOST"), utils.Env("DB_PORT"), utils.Env("DB_DATABASE"), utils.Env("DB_USERNAME"), utils.Env("DB_PASSWORD"))

	if err != nil {
		log.Panic("Database is NIL: " + err.Error())
		return
	}

	if db == nil {
		log.Panic("Database is NIL")
		return
	}

	log.Println("4. Starting server on http://" + utils.Env("SERVER_HOST") + ":" + utils.Env("SERVER_PORT") + " ...")
	log.Println("URL: http://" + utils.Env("APP_URL") + " ...")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		menu := hb.NewHTML("<a href='/crud'>Standalone CRUD</a>")
		w.Write([]byte(menu.ToHTML()))
	})
	mux.HandleFunc("/crud", crudHandler)

	srv := &http.Server{
		Handler: mux,
		Addr:    utils.Env("SERVER_HOST") + ":" + utils.Env("SERVER_PORT"),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func mainDb(driverName string, dbHost string, dbPort string, dbName string, dbUser string, dbPass string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if driverName == "sqlite" {
		dsn := dbName
		db, err = sql.Open("sqlite", dsn)
	}
	if driverName == "mysql" {
		dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
		db, err = sql.Open("mysql", dsn)
	}
	if driverName == "postgres" {
		dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort + " sslmode=disable TimeZone=Europe/London"
		db, err = sql.Open("postgres", dsn)
	}
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, errors.New("database for driver " + driverName + " could not be intialized")
	}

	return db, nil
}
