package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"final-project/pkg/db"
)

// writeJSON сериализует данные в JSON и отправляет ответ
func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendError отправляет JSON с полем error
func sendError(w http.ResponseWriter, errMsg string, statusCode int) {
	writeJSON(w, map[string]string{"error": errMsg}, statusCode)
}

// taskHandler распределяет запросы по нужным методам
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// checkDate проверяет и форматирует дату и правило повторения Задачи
func checkDate(task *db.Task) error {
	now := time.Now()
	nowStr := now.Format(DateFormat)

	if task.Date == "" {
		task.Date = nowStr
	}

	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты: %v", err)
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("неверный формат правила повторения: %v", err)
		}
	}

	if nowStr > t.Format(DateFormat) {
		if task.Repeat == "" {
			task.Date = nowStr
		} else {
			task.Date = next
		}
	}

	return nil
}

// addTaskHandler обрабатывает добавление новой задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		sendError(w, "ошибка при добавлении задачи в БД", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"id": fmt.Sprint(id)}, http.StatusOK)
}

// getTaskHandler возвращает Задачу по переданному id
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		sendError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

// updateTaskHandler обновляет данные существующей Задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		sendError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		sendError(w, "Задача не найдена", http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

// deleteTaskHandler удаляет Задачу по переданному id
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		sendError(w, "Задача не найдена", http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

// doneTaskHandler отмечает Задачу как выполненную
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		sendError(w, "Задача не найдена", http.StatusBadRequest)
		return
	}

	if task.Repeat == "" {
		// Если разовая Задача, то удаляем
		if err := db.DeleteTask(id); err != nil {
			sendError(w, "Ошибка при удалении задачи", http.StatusInternalServerError)
			return
		}
	} else {
		// Если периодическая Задача, то вычисляем новую дату
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendError(w, "Ошибка при вычислении следующей даты", http.StatusInternalServerError)
			return
		}

		if err := db.UpdateDate(nextDate, id); err != nil {
			sendError(w, "Ошибка при обновлении даты задачи", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}
