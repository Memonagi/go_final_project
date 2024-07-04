package database

import (
	"database/sql"
	"errors"
	"github.com/Memonagi/go_final_project/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
)

type DB struct {
	db *sql.DB
}

// New подключает к БД
func New(dbFile string) (*DB, error) {

	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	createTable := `CREATE TABLE IF NOT EXISTS scheduler (
        id       INTEGER PRIMARY KEY AUTOINCREMENT,
        date     CHAR(8)      NOT NULL,
        title    VARCHAR(128) NOT NULL,
        comment  TEXT,
        repeat   VARCHAR(128)  NOT NULL
    );`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// CloseDatabase закрывает БД
func (db *DB) CloseDatabase() error {
	return db.db.Close()
}

// AddTask добавляет задачу в БД
func (db *DB) AddTask(task models.Task) (string, error) {

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(id)), nil
}

// GetAllTasks получает все задачи из БД
func (db *DB) GetAllTasks() ([]models.Task, error) {

	rows, err := db.db.Query("SELECT * FROM scheduler ORDER BY date LIMIT ?", models.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {

		var taskStruct models.Task

		if err = rows.Scan(&taskStruct.ID, &taskStruct.Date, &taskStruct.Title, &taskStruct.Comment, &taskStruct.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, taskStruct)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskId получает задачу из БД по ее ID
func (db *DB) GetTaskId(id int64, task models.Task) (models.Task, error) {

	query := "SELECT * FROM scheduler WHERE id = ?"

	if err := db.db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

// UpdateTask редактирует задачу в БД
func (db *DB) UpdateTask(task models.Task) (models.Task, error) {

	row, err := db.db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return models.Task{}, err
	}
	checkRow, err := row.RowsAffected()
	if err != nil || checkRow == 0 {
		return models.Task{}, errors.New("ошибка обновления задачи в базе данных")
	}
	return task, nil
}

// TaskDone выполняет задачу в БД
func (db *DB) TaskDone(nextDate string, id int64) error {

	_, err := db.db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTaskId удаляет задачу из БД
func (db *DB) DeleteTaskId(id int64) error {

	row, err := db.db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}

	checkRow, err := row.RowsAffected()
	if err != nil || checkRow == 0 {
		return err
	}
	return nil
}
