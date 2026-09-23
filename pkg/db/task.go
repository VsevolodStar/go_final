package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Поля структуры Задачи
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в БД и возвращает идентификатор новой записи
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Tasks возвращает список задач, опционально отфильтрованный по строке поиска и с ограничением количества
func Tasks(search string, limit int) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		t, errParse := time.Parse("02.01.2006", search)
		if errParse == nil {
			dateStr := t.Format("20060102")
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = :date ORDER BY date LIMIT :limit`
			rows, err = DB.Query(query, sql.Named("date", dateStr), sql.Named("limit", limit))
		} else {
			searchStr := "%" + search + "%"
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit`
			rows, err = DB.Query(query, sql.Named("search", searchStr), sql.Named("limit", limit))
		}
	} else {
		query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit`
		rows, err = DB.Query(query, sql.Named("limit", limit))
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		var t Task
		var id int64
		var comment, repeat sql.NullString
		err := rows.Scan(&id, &t.Date, &t.Title, &comment, &repeat)
		if err != nil {
			return nil, err
		}
		t.ID = fmt.Sprint(id)
		if comment.Valid {
			t.Comment = comment.String
		}
		if repeat.Valid {
			t.Repeat = repeat.String
		}
		tasks = append(tasks, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTask возвращает задачу по её id
func GetTask(id string) (*Task, error) {
	var t Task
	var dbID int64
	var comment, repeat sql.NullString
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	err := DB.QueryRow(query, id).Scan(&dbID, &t.Date, &t.Title, &comment, &repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, err
	}
	t.ID = fmt.Sprint(dbID)
	if comment.Valid {
		t.Comment = comment.String
	}
	if repeat.Valid {
		t.Repeat = repeat.String
	}
	return &t, nil
}

// UpdateTask обновляет атрибуты Задачи
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`задача не найдена`)
	}
	return nil
}

// DeleteTask удаляет Задачу из БД
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

// UpdateDate обновляет дату выполнения Задачи
func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, next, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}
