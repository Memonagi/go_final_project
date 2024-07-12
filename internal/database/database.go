package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/Memonagi/go_final_project/internal/models"
	// Импорт для работы с БД.
	_ "github.com/mattn/go-sqlite3"
)

const limit = 50

type DB struct {
	db *sql.DB
}

// NewDB подключает к БД.
func NewDB(ctx context.Context, dbFile string) (*DB, error) {
	_, err := os.Stat(dbFile)

	if os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания БД: %w", err)
		}

		if err = file.Close(); err != nil {
			return nil, fmt.Errorf("ошибка закрытия БД: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла БД: %w", err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS scheduler (
        id       INTEGER PRIMARY KEY AUTOINCREMENT,
        date     CHAR(8)      NOT NULL,
        title    VARCHAR(128) NOT NULL,
        comment  TEXT,
        repeat   VARCHAR(128)  NOT NULL
    );`

	_, err = db.ExecContext(ctx, createTable)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы: %w", err)
	}

	return &DB{db}, nil
}

// CloseDatabase закрывает БД.
func (db *DB) CloseDatabase() error {
	if err := db.db.Close(); err != nil {
		return fmt.Errorf("ошибка отключения БД: %w", err)
	}

	return nil
}

// AddTask добавляет задачу в БД.
func (db *DB) AddTask(ctx context.Context, task models.Task) (string, error) {
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"

	res, err := db.db.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", fmt.Errorf("ошибка добавления задачи в БД: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("ошибка получения ID добавленной задачи: %w", err)
	}

	return strconv.Itoa(int(id)), nil
}

// GetAllTasks получает все задачи из БД.
func (db *DB) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	rows, err := db.db.QueryContext(ctx, "SELECT * FROM scheduler ORDER BY date LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска задач в БД: %w", err)
	}

	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var taskStruct models.Task

		err = rows.Scan(&taskStruct.ID, &taskStruct.Date, &taskStruct.Title, &taskStruct.Comment, &taskStruct.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения списка задач из БД: %w", err)
		}

		tasks = append(tasks, taskStruct)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка получения списка задач из БД: %w", err)
	}

	if len(tasks) == 0 {
		return []models.Task{}, nil
	}

	return tasks, nil
}

// GetTaskID получает задачу из БД по ее ID.
func (db *DB) GetTaskID(ctx context.Context, id int64, task models.Task) (models.Task, error) {
	query := "SELECT * FROM scheduler WHERE id = ?"

	err := db.db.QueryRowContext(ctx, query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return models.Task{}, fmt.Errorf("ошибка получения задачи из БД: %w", err)
	}

	return task, nil
}

// UpdateTask редактирует задачу в БД.
func (db *DB) UpdateTask(ctx context.Context, task models.Task) (models.Task, error) {
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"

	row, err := db.db.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return models.Task{}, fmt.Errorf("ошибка обновления задачи в базе данных: %w", err)
	}

	checkRow, err := row.RowsAffected()

	if err != nil || checkRow == 0 {
		return models.Task{}, fmt.Errorf("ошибка обновления задачи в базе данных: %w", err)
	}

	return task, nil
}

// TaskDone выполняет задачу в БД.
func (db *DB) TaskDone(ctx context.Context, nextDate string, id int64) error {
	_, err := db.db.ExecContext(ctx, "UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
	if err != nil {
		return fmt.Errorf("ошибка выполнения задачи: %w", err)
	}

	return nil
}

// DeleteTaskID удаляет задачу из БД.
func (db *DB) DeleteTaskID(ctx context.Context, id int64) error {
	row, err := db.db.ExecContext(ctx, "DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	checkRow, err := row.RowsAffected()

	if err != nil || checkRow == 0 {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	return nil
}
