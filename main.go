package main

import (
	"78/database"
	"78/handlers"
	"78/logger"
	"fmt"
	"log"
	"net/http"
)

func main() {

	//*********************
	err := database.Connect()
	if err != nil {
		log.Fatal("Connection failed", err)
		logger.Log.Error("could not connect to database") //log to logger
	}
	//*********************

	defer database.Close()
	//*********************

	const dsn = "postgres://postgres:@localhost:5432/postgres?sslmode=disable"

	if err := database.RunMigrations(dsn); err != nil {
		log.Fatal(err)
		logger.Log.Error("could not migrate") //log to logger

	}

	//*********************

	mux := http.NewServeMux() //creates a "router - maps urls to handler function

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateTask(w, r)
		case http.MethodGet:
			handlers.GetAllTasks(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}

	})

	mux.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTaskByID(w, r)

		case http.MethodDelete:
			handlers.DeleteTask(w, r)

		case http.MethodPut:
			handlers.UpdateTask(w, r)

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		}
	})
	//*********************
	if err := logger.Init(); err != nil {
		log.Fatal("failed to initialize log", err)
	}
	defer logger.Close()

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
