package api

import (
	"net/http"

	"final-project/pkg/db"
)

const limit = 20

type tasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// TasksHandler обрабатывает запрос на получение списка Задач с возможностью поиска и ограничением количества
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")
	tasks, err := db.Tasks(search, limit) // Ограничили запрос в limit задач
	if err != nil {
		sendError(w, "ошибка при получении задач: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tasks == nil {
		tasks = []*db.Task{}
	}

	writeJSON(w, tasksResp{Tasks: tasks}, http.StatusOK)
}
