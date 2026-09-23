package server

import (
	"log"
	"net/http"

	"final-project/pkg/api"
)

const (
	// Порт по умолчанию
	DefaultPort = "7540"
	// Директория с файлами фронтенда
	DefaultWebDir = "./web"
)

// InitRouter настраивает маршрутизацию, включая раздачу файлов фронта
func InitRouter() http.Handler {
	mux := http.NewServeMux()

	api.Init(mux)

	// Раздача файлов фронтенда
	fileServer := http.FileServer(http.Dir(DefaultWebDir))
	mux.Handle("/", fileServer)

	return mux
}

// Start запускает веб-сервер на указанном порту.
func Start(port string) error {
	handler := InitRouter()

	log.Printf("Сервер запущен и слушает порт %s (директория статики: %s)...\n", port, DefaultWebDir)
	return http.ListenAndServe(":"+port, handler)
}
