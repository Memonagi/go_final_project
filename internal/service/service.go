package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Memonagi/go_final_project/internal/database"
	"github.com/Memonagi/go_final_project/internal/date"
	"github.com/Memonagi/go_final_project/internal/models"
)

type Service struct {
	db *database.DB
}

func New(db *database.DB) *Service {
	return &Service{
		db: db,
	}
}

// CheckRepeat проверяет корректность указанного правила повторения
func (s *Service) CheckRepeat(task models.Task) error {
	if task.Repeat == "" {
		return nil
	}
	switch string(task.Repeat[0]) {
	case "y":
		return nil
	case "d":
		s := strings.Split(task.Repeat, " ")
		if len(s) != 2 {
			return errors.New("правило повторения указано в неправильном формате")
		} else {
			days, err := strconv.Atoi(s[1])
			if err != nil || days < 1 || days > 400 {
				return errors.New("указано неверное количество дней")
			}
		}
	case "w":
		s := strings.Split(task.Repeat, " ")
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
func (s *Service) CheckTitle(task models.Task) (string, error) {
	if len(task.Title) == 0 {
		return "", errors.New("заголовок задачи не может быть пустым")
	}
	return task.Title, nil
}

// CheckDate проверяет корректность указанной даты
func (s *Service) CheckDate(task models.Task) (string, error) {
	now := time.Now()
	if task.Date == "" || task.Date == "today" {
		return now.Format(models.DateFormat), nil
	} else {
		outDate, err := time.Parse(models.DateFormat, task.Date)
		if err != nil {
			return "", errors.New("неправильный формат даты")
		}
		if outDate.Before(now) {
			return now.Format(models.DateFormat), nil
		} else {
			return outDate.Format(models.DateFormat), nil
		}
	}
}

// AddTask добавляет новую задачу в БД
func (s *Service) AddTask(ctx context.Context, task models.Task) (string, error) {
	titleOfTask, err := s.CheckTitle(task)
	if err != nil {
		return "", err
	}
	task.Title = titleOfTask

	now := time.Now()
	if task.Repeat == "" {
		dateOfTask, err := s.CheckDate(task)
		if err != nil {
			return "", err
		}
		task.Date = dateOfTask
	} else {
		err = s.CheckRepeat(task)
		if err != nil {
			return "", err
		}
		dateOfTask, err := s.CheckDate(task)
		if err != nil {
			return "", err
		}
		if dateOfTask == now.Format(models.DateFormat) {
			task.Date = dateOfTask
		} else {
			nextDate, err := date.NextDate(now, dateOfTask, task.Repeat)
			if err != nil {
				return "", err
			}
			task.Date = nextDate
		}
	}
	return s.db.AddTask(ctx, task)
}

// GetAllTasks получает список ближайших задач
func (s *Service) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	return s.db.GetAllTasks(ctx)
}

// GetTaskId получает задачу по ее ID
func (s *Service) GetTaskId(ctx context.Context, id string) (models.Task, error) {
	if id == "" {
		return models.Task{}, errors.New("не указан идентификатор")
	}

	var task models.Task

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.Task{}, err
	}

	return s.db.GetTaskId(ctx, int64(idInt), task)
}

// UpdateTask редактирует задачу
func (s *Service) UpdateTask(ctx context.Context, task models.Task) (models.Task, error) {
	if task.ID == "" {
		return models.Task{}, errors.New("не указан ID")
	}

	titleOfTask, err := s.CheckTitle(task)
	if err != nil {
		return models.Task{}, err
	}
	task.Title = titleOfTask

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(models.DateFormat)
	} else {
		dateOfTask, err := time.Parse(models.DateFormat, task.Date)
		if err != nil {
			return models.Task{}, err
		}
		if dateOfTask.Before(now) {
			if task.Repeat == "" {
				task.Date = now.Format(models.DateFormat)
			} else {
				nextDate, err := date.NextDate(now, task.Date, task.Repeat)
				if err != nil {
					return models.Task{}, err
				}
				task.Date = nextDate
			}
		}
	}

	if err := s.CheckRepeat(task); err != nil {
		return models.Task{}, err
	}
	return s.db.UpdateTask(ctx, task)
}

// TaskDone делает задачу выполненной
func (s *Service) TaskDone(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("не указан ID")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	var task models.Task

	task, err = s.db.GetTaskId(ctx, int64(idInt), task)
	if err != nil {
		return err
	}

	switch task.Repeat {
	case "":
		return s.db.DeleteTaskId(ctx, int64(idInt))
	default:
		now := time.Now()
		nextDate, err := date.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
		if err = s.db.TaskDone(ctx, nextDate, int64(idInt)); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTask удаляет задачу
func (s *Service) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("не указан ID")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return s.db.DeleteTaskId(ctx, int64(idInt))
}
