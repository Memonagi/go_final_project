package service

import (
	"errors"
	"github.com/Memonagi/go_final_project/internal/constants"
	"github.com/Memonagi/go_final_project/internal/database"
	"github.com/Memonagi/go_final_project/internal/date"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
	db      database.DB
}

// CheckRepeat проверяет корректность указанного правила повторения
func (t *Task) CheckRepeat() error {

	if t.Repeat == "" {
		return nil
	}
	switch string(t.Repeat[0]) {
	case "y":
		return nil
	case "d":
		s := strings.Split(t.Repeat, " ")
		if len(s) != 2 {
			return errors.New("правило повторения указано в неправильном формате")
		} else {
			days, err := strconv.Atoi(s[1])
			if err != nil || days < 1 || days > 400 {
				return errors.New("указано неверное количество дней")
			}
		}
	case "w":
		s := strings.Split(t.Repeat, " ")
		if len(s) != 2 {
			return errors.New("правило повторения указано в неправильном формате")
		} else {
			weekDays := strings.Split(s[1], ",")
			for _, e := range weekDays {
				wDay, err := strconv.Atoi(e)
				if err != nil || wDay < 1 || wDay > 7 {
					return errors.New("указано неверное количество дней")
				}
			}
		}
	default:
		return errors.New("правило повторения указано в неправильном формате")
	}
	return nil
}

// CheckTitle проверяет наличие заголовка
func (t *Task) CheckTitle() (string, error) {

	if len(t.Title) == 0 {
		return "", errors.New("заголовок задачи не может быть пустым")
	}
	return t.Title, nil
}

// CheckDate проверяет корректность указанной даты
func (t *Task) CheckDate() (string, error) {

	now := time.Now()
	if t.Date == "" || t.Date == "today" {
		return now.Format(constants.DateFormat), nil
	} else {
		outDate, err := time.Parse(constants.DateFormat, t.Date)
		if err != nil {
			return "", errors.New("неправильный формат даты")
		}
		if outDate.Before(now) {
			return now.Format(constants.DateFormat), nil
		} else {
			return outDate.Format(constants.DateFormat), nil
		}
	}
}

// AddTask добавляет новую задачу в БД
func (t *Task) AddTask() (int64, error) {

	titleOfTask, err := t.CheckTitle()
	if err != nil {
		return 0, err
	}
	t.Title = titleOfTask

	now := time.Now()
	if t.Repeat == "" {
		dateOfTask, err := t.CheckDate()
		if err != nil {
			return 0, err
		}
		t.Date = dateOfTask
	} else {
		err = t.CheckRepeat()
		if err != nil {
			return 0, err
		}
		dateOfTask, err := t.CheckDate()
		if err != nil {
			return 0, err
		}
		if dateOfTask == now.Format(constants.DateFormat) {
			t.Date = dateOfTask
		} else {
			nextDate, err := date.NextDate(now, dateOfTask, t.Repeat)
			if err != nil {
				return 0, err
			}
			t.Date = nextDate
		}
	}
	task := constants.Task{
		ID:      t.ID,
		Date:    t.Date,
		Title:   t.Title,
		Comment: t.Comment,
		Repeat:  t.Repeat,
	}
	return t.db.AddTask(task)
}

// GetAllTasks получает список ближайших задач
func (t *Task) GetAllTasks() ([]constants.Task, error) {
	return t.db.GetAllTasks()
}

// GetTaskId получает задачу по ее ID
func (t *Task) GetTaskId(id string) (constants.Task, error) {

	if id == "" {
		return constants.Task{}, nil
	}

	var task constants.Task

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return constants.Task{}, err
	}

	return t.db.GetTaskId(int64(idInt), task)
}

// UpdateTask редактирует задачу
func (t *Task) UpdateTask() error {

	titleOfTask, err := t.CheckTitle()
	if err != nil {
		return err
	}
	t.Title = titleOfTask

	now := time.Now()
	if t.Date == "" {
		t.Date = now.Format(constants.DateFormat)
	} else {
		dateOfTask, err := time.Parse(constants.DateFormat, t.Date)
		if err != nil {
			return err
		}
		if dateOfTask.Before(now) {
			if t.Repeat == "" {
				t.Date = now.Format(constants.DateFormat)
			} else {
				nextDate, err := date.NextDate(now, t.Date, t.Repeat)
				if err != nil {
					return err
				}
				t.Date = nextDate
			}
		}
	}

	if err := t.CheckRepeat(); err != nil {
		return err
	}
	task := constants.Task{
		ID:      t.ID,
		Date:    t.Date,
		Title:   t.Title,
		Comment: t.Comment,
		Repeat:  t.Repeat,
	}
	return t.db.UpdateTask(task)
}

// DoneTask делает задачу выполненной
func (t *Task) DoneTask(id string) error {

	if id == "" {
		return errors.New("не указан ID")
	}

	var task constants.Task
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	_, err = t.db.GetTaskId(int64(idInt), task)
	if err != nil {
		return err
	}

	switch t.Repeat {
	case "":
		return t.db.DeleteTaskId(int64(idInt))
	default:
		now := time.Now()
		nextDate, err := date.NextDate(now, t.Date, t.Repeat)
		if err != nil {
			return err
		}
		if err = t.db.DoneTask(nextDate, int64(idInt)); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTask удаляет задачу
func (t *Task) DeleteTask(id string) error {
	if id == "" {
		return errors.New("не указан ID")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return t.db.DeleteTaskId(int64(idInt))
}
