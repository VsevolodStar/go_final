package api

import (
	"net/http"
)

// Init регистрирует все API-обработчики в переданном ServeMux
func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", NextDateHandler)
	mux.HandleFunc("/api/task", auth(taskHandler))
	mux.HandleFunc("/api/task/done", auth(doneTaskHandler))
	mux.HandleFunc("/api/tasks", auth(TasksHandler))
	mux.HandleFunc("/api/signin", signinHandler)
}
