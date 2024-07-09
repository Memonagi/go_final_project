package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Memonagi/go_final_project/internal/database"
	"github.com/Memonagi/go_final_project/internal/handler"
	"github.com/Memonagi/go_final_project/internal/service"
	"github.com/sirupsen/logrus"
)

const defaultPort = 7540

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	port, _ := strconv.Atoi(os.Getenv("TODO_PORT"))
	if port == 0 {
		port = defaultPort
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	db, err := database.New(ctx, dbFile)
	if err != nil {
		logrus.Panicf("ошибка подключения к БД: %v", err)
	}
	defer func() {
		if err := db.CloseDatabase(); err != nil {
			logrus.Warnf("ошибка закрытия БД: %v", err)
		}
	}()

	svc := service.New(db)

	server := handler.New(port, svc)

	if err := server.Run(ctx); err != nil {
		logrus.Panicf("ошибка запуска сервера: %v", err)
	}
}
