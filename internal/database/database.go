package database

import (
	"database/sql"
	"github.com/Memonagi/go_final_project/internal/constants"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type DB struct {
	Db *sql.DB
}

// CheckDatabase проверяет существование БД
func (db *DB) CheckDatabase() error {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return err
		}
		file.Close()
	}

	db.Db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}

	createTable := `CREATE TABLE IF NOT EXISTS scheduler (
        id       INTEGER PRIMARY KEY AUTOINCREMENT,
        date     CHAR(8)      NOT NULL,
        title    VARCHAR(128) NOT NULL,
        comment  TEXT,
        repeat   VARCHAR(128)  NOT NULL
    );`

	_, err = db.Db.Exec(createTable)
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase подключает к БД
func (db *DB) NewDatabase() (*DB, error) {

	if err := db.CheckDatabase(); err != nil {
		return nil, err
	}

	var err error
	db.Db, err = sql.Open("sqlite3", "scheduler.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CloseDatabase закрывает БД
func (db *DB) CloseDatabase() error {
	return db.Db.Close()
}

// AddTask добавляет задачу в БД
func (db *DB) AddTask(task constants.Task) (int64, error) {

	db, err := db.NewDatabase()
	if err != nil {
		return 0, err
	}
	defer db.CloseDatabase()

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetAllTasks получает все задачи из БД
func (db *DB) GetAllTasks() ([]constants.Task, error) {

	db, err := db.NewDatabase()
	if err != nil {
		return nil, err
	}
	defer db.CloseDatabase()

	rows, err := db.Db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?", constants.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []constants.Task

	for rows.Next() {

		var taskStruct constants.Task

		if err = rows.Scan(&taskStruct.ID, &taskStruct.Date, &taskStruct.Title, &taskStruct.Comment, &taskStruct.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, taskStruct)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []constants.Task{}
	}
	return tasks, nil
}

// GetTaskId получает задачу из БД по ее ID
func (db *DB) GetTaskId(id int64, task constants.Task) (constants.Task, error) {
	db, err := db.NewDatabase()
	if err != nil {
		return constants.Task{}, err
	}
	defer db.CloseDatabase()

	query := "SELECT * FROM scheduler WHERE id = ?"

	if err := db.Db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		return constants.Task{}, err
	}
	return task, nil
}

// UpdateTask редактирует задачу в БД
func (db *DB) UpdateTask(task constants.Task) error {
	db, err := db.NewDatabase()
	if err != nil {
		return err
	}
	defer db.CloseDatabase()

	row, err := db.Db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	checkRow, err := row.RowsAffected()
	if err != nil || checkRow == 0 {
		return err
	}
	return nil
}

// DoneTask выполняет задачу в БД
func (db *DB) DoneTask(nextDate string, id int64) error {
	db, err := db.NewDatabase()
	if err != nil {
		return err
	}
	defer db.CloseDatabase()

	_, err = db.Db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTaskId удаляет задачу из БД
func (db *DB) DeleteTaskId(id int64) error {
	db, err := db.NewDatabase()
	if err != nil {
		return err
	}
	defer db.CloseDatabase()

	row, err := db.Db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}

	checkRow, err := row.RowsAffected()
	if err != nil || checkRow == 0 {
		return err
	}
	return nil
}
