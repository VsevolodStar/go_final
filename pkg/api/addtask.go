package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"final-project/pkg/db"
)

// writeJSON сериализует данные в JSON и отправляет ответ
func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования ответа: %v", err)
	}
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
	if task == nil {
		return fmt.Errorf("задача не может быть nil")
	}

	now := time.Now()
	nowStr := now.Format(db.DateFormat)

	if task.Date == "" {
		task.Date = nowStr
	}

	t, err := time.Parse(db.DateFormat, task.Date)
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

	if nowStr > t.Format(db.DateFormat) {
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
		sendError(w, "ошибка десериализации JSON: "+err.Error(), http.StatusBadRequest)
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
		log.Printf("Ошибка добавления задачи в БД: %v", err)
		sendError(w, "Ошибка при добавлении задачи в БД: "+err.Error(), http.StatusInternalServerError)
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
		sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

// updateTaskHandler обновляет данные существующей Задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, "Ошибка десериализации JSON: "+err.Error(), http.StatusBadRequest)
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
		sendError(w, err.Error(), http.StatusBadRequest)
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
		sendError(w, err.Error(), http.StatusBadRequest)
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
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.Repeat == "" {
		// Если разовая Задача, то удаляем
		if err := db.DeleteTask(id); err != nil {
			log.Printf("Ошибка при удалении задачи: %v", err)
			sendError(w, "Ошибка при удалении задачи: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Если периодическая Задача, то вычисляем новую дату
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			log.Printf("Ошибка при вычислении следующей даты: %v", err)
			sendError(w, "Ошибка при вычислении следующей даты: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateDate(nextDate, id); err != nil {
			log.Printf("Ошибка при обновлении даты задачи: %v", err)
			sendError(w, "Ошибка при обновлении даты задачи: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}
