package main

import (
	"github.com/Memonagi/go_final_project/internal/database"
	"github.com/Memonagi/go_final_project/internal/handler"
	"github.com/Memonagi/go_final_project/internal/service"
	"os"
	"strconv"

	"github.com/Memonagi/go_final_project/tests"
)

func main() {

	port, _ := strconv.Atoi(os.Getenv("TODO_PORT"))
	if port == 0 {
		port = tests.Port
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	db, err := database.New(dbFile)
	if err != nil {
		panic(err)
	}
	defer db.CloseDatabase()

	service := service.New(db)

	server := handler.New(port, service)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
