package main

import (
	"log"
	"os"

	"final-project/pkg/db"
	"final-project/pkg/server"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Ошибка при инициализации БД: %v", err)
	}
	defer db.DB.Close()

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = server.DefaultPort
	}

	log.Printf("Запуск веб-сервера на порту %s...\n", port)

	if err := server.Start(port); err != nil {
		log.Printf("Ошибка при работе сервера: %v", err)
		return
	}
}
