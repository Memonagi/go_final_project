package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Memonagi/go_final_project/internal/database"
	"github.com/Memonagi/go_final_project/internal/date"
	"github.com/Memonagi/go_final_project/internal/models"
)

const dateFormat = "20060102"

type Service struct {
	db *database.DB
}

var (
	errRule  = errors.New("правило повторения указано в неправильном формате")
	errDays  = errors.New("указано неверное количество дней")
	errTitle = errors.New("заголовок задачи не может быть пустым")
	errDate  = errors.New("неправильный формат даты")
	errID    = errors.New("не указан ID")
)

func New(db *database.DB) *Service {
	return &Service{
		db: db,
	}
}

// CheckRepeat проверяет корректность указанного правила повторения.
func (s *Service) checkRepeat(task models.Task) error {
	if task.Repeat == "" {
		return nil
	}

	switch string(task.Repeat[0]) {
	case "y":
		return nil
	case "d":
		if err := s.checkDays(task); err != nil {
			return err
		}
	case "w":
		if err := s.checkWeek(task); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w", errRule)
	}

	return nil
}

// checkDays проверяет корректность указанного правила повторения дней.
func (s *Service) checkDays(task models.Task) error {
	slice := strings.Split(task.Repeat, " ")

	if len(slice) != 2 {
		return fmt.Errorf("%w", errRule)
	}

	days, err := strconv.Atoi(slice[1])

	if err != nil || days < 1 || days > 400 {
		return fmt.Errorf("%w", errDays)
	}

	return nil
}

// checkWeek проверяет корректность указанного правила повторения дней недели.
func (s *Service) checkWeek(task models.Task) error {
	slice := strings.Split(task.Repeat, " ")

	if len(slice) != 2 {
		return fmt.Errorf("%w", errRule)
	}

	weekDays := strings.Split(slice[1], ",")

	for _, e := range weekDays {
		wDay, err := strconv.Atoi(e)
		if err != nil || wDay < 1 || wDay > 7 {
			return fmt.Errorf("%w", errDays)
		}
	}

	return nil
}

// CheckTitle проверяет наличие заголовка.
func (s *Service) checkTitle(task models.Task) (string, error) {
	if len(task.Title) == 0 {
		return "", fmt.Errorf("%w", errTitle)
	}

	return task.Title, nil
}

// CheckDate проверяет корректность указанной даты.
func (s *Service) checkDate(task models.Task) (string, error) {
	now := time.Now()

	if task.Date == "" || task.Date == "today" {
		return now.Format(dateFormat), nil
	}

	outDate, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return "", fmt.Errorf("%w", errDate)
	}

	if outDate.Before(now) {
		return now.Format(dateFormat), nil
	}

	return outDate.Format(dateFormat), nil
}

// AddTask добавляет новую задачу в БД.
func (s *Service) AddTask(ctx context.Context, task models.Task) (string, error) {
	titleOfTask, err := s.checkTitle(task)
	if err != nil {
		return "", err
	}

	task.Title = titleOfTask

	now := time.Now()

	dateOfTask, err := s.addTaskHelper(task, now)
	if err != nil {
		return "", err
	}

	task.Date = dateOfTask

	taskID, err := s.db.AddTask(ctx, task)
	if err != nil {
		return "", fmt.Errorf("ошибка добавления задачи: %w", err)
	}

	return taskID, nil
}

func (s *Service) addTaskHelper(task models.Task, now time.Time) (string, error) {
	if task.Repeat == "" {
		dateOfTask, err := s.checkDate(task)
		if err != nil {
			return "", err
		}

		return dateOfTask, nil
	}

	err := s.checkRepeat(task)
	if err != nil {
		return "", err
	}

	dateOfTask, err := s.checkDate(task)
	if err != nil {
		return "", err
	}

	if dateOfTask == now.Format(dateFormat) {
		task.Date = dateOfTask
	} else {
		nextDate, err := date.NextDate(now, dateOfTask, task.Repeat)
		if err != nil {
			return "", fmt.Errorf("ошибка вычисления следующей даты: %w", err)
		}
		task.Date = nextDate
	}

	return task.Date, nil
}

// GetAllTasks получает список ближайших задач.
func (s *Service) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	tasks, err := s.db.GetAllTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка задач: %w", err)
	}

	return tasks, nil
}

// GetTaskID получает задачу по ее ID.
func (s *Service) GetTaskID(ctx context.Context, id string) (models.Task, error) {
	if id == "" {
		return models.Task{}, fmt.Errorf("%w", errID)
	}

	var task models.Task

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.Task{}, fmt.Errorf("ошибка конвертации ID: %w", err)
	}

	taskID, err := s.db.GetTaskID(ctx, int64(idInt), task)
	if err != nil {
		return models.Task{}, fmt.Errorf("ошибка получения задачи из списка: %w", err)
	}

	return taskID, nil
}

// UpdateTask редактирует задачу.
func (s *Service) UpdateTask(ctx context.Context, task models.Task) (models.Task, error) {
	if task.ID == "" {
		return models.Task{}, fmt.Errorf("%w", errID)
	}

	titleOfTask, err := s.checkTitle(task)
	if err != nil {
		return models.Task{}, err
	}

	task.Title = titleOfTask

	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	dateOfTask, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return models.Task{}, fmt.Errorf("%w", errDate)
	}

	if dateOfTask.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(dateFormat)
		}

		nextDate, err := date.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return models.Task{}, fmt.Errorf("ошибка вычисления следующей даты: %w", err)
		}

		task.Date = nextDate
	}

	if err := s.checkRepeat(task); err != nil {
		return models.Task{}, fmt.Errorf("%w", errRule)
	}

	updatedTask, err := s.db.UpdateTask(ctx, task)
	if err != nil {
		return models.Task{}, fmt.Errorf("ошибка обновления задачи: %w", err)
	}

	return updatedTask, nil
}

// TaskDone делает задачу выполненной.
func (s *Service) TaskDone(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w", errID)
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("ошибка конвертации ID: %w", err)
	}

	var task models.Task

	task, err = s.db.GetTaskID(ctx, int64(idInt), task)
	if err != nil {
		return fmt.Errorf("ошибка получения задачи из списка: %w", err)
	}

	switch task.Repeat {
	case "":
		if err = s.db.DeleteTaskID(ctx, int64(idInt)); err != nil {
			return fmt.Errorf("ошибка удаления задачи: %w", err)
		}
	default:
		now := time.Now()

		nextDate, err := date.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("ошибка вычисления следующей даты: %w", err)
		}

		if err = s.db.TaskDone(ctx, nextDate, int64(idInt)); err != nil {
			return fmt.Errorf("ошибка выполнения задачи: %w", err)
		}
	}

	return nil
}

// DeleteTask удаляет задачу.
func (s *Service) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w", errID)
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("ошибка конвертации ID: %w", err)
	}

	if err := s.db.DeleteTaskID(ctx, int64(idInt)); err != nil {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	return nil
}
