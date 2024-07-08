package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Memonagi/go_final_project/internal/date"
	"github.com/Memonagi/go_final_project/internal/models"
	"github.com/Memonagi/go_final_project/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *service.Service
	server  http.Server
	port    int
}

// New создает маршрутизатор и обрабатывает запросы
func New(port int, service *service.Service) *Handler {
	r := chi.NewRouter()
	r.Handle("/*", http.FileServer(http.Dir(models.WebDir)))

	h := Handler{
		service: service,
		server: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
		port: port,
	}

	r.Route("/api", func(r chi.Router) {
		r.Get("/nextdate", h.getNextDate)
		r.Get("/tasks", h.getAllTasks)
		r.Route("/task", func(r chi.Router) {
			r.Post("/", h.addTask)
			r.Get("/", h.getTaskId)
			r.Put("/", h.updateTaskId)
			r.Post("/done", h.taskDone)
			r.Delete("/", h.deleteTask)
		})
	})

	return &h
}

// Run запускает сервер
func (h *Handler) Run(ctx context.Context) error {
	logrus.Infof("запуск веб-сервера на порту %d", h.port)

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	go func() {
		<-ctx.Done()
		logrus.Info("закрытие сервера")

		ctxGf, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := h.server.Shutdown(ctxGf); err != nil {
			logrus.Warn("ошибка плавного закрытия сервера")
			return
		}
	}()

	if err := h.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("ошибка запуска сервера")
	}
	return nil
}

// errorResponse возвращает ошибку в формате {"error":"текст ошибки"}
func errorResponse(w http.ResponseWriter, errorText string, err error) {
	errorResponse := models.Response{
		Error: fmt.Errorf("%s: %w", errorText, err).Error(),
	}
	response, _ := json.Marshal(errorResponse)
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, fmt.Errorf("error: %w", err).Error(), http.StatusInternalServerError)
	}
}

// getNextDate GET-обработчик для получения следующей даты
func (h *Handler) getNextDate(w http.ResponseWriter, r *http.Request) {
	nowReq := r.FormValue("now")
	dateReq := r.FormValue("date")
	repeatReq := r.FormValue("repeat")

	now, err := time.Parse(models.DateFormat, nowReq)
	if err != nil {
		http.Error(w, "неправильный формат даты", http.StatusInternalServerError)
		return
	}

	nextDate, err := date.NextDate(now, dateReq, repeatReq)
	if err != nil {
		http.Error(w, "ошибка вычисления следующей даты", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))
	if err != nil {
		logrus.Warn("ошибка записи следующей даты")
		return
	}
}

// addTask POST-обработчик для добавления новой задачи
func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errorResponse(w, "ошибка десериализации JSON", err)
		return
	}

	taskId, err := h.service.AddTask(r.Context(), task)
	if err != nil {
		errorResponse(w, "не удалось добавить новую задачу", err)
		return
	}

	response := models.Response{Id: taskId}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		logrus.Warn("ошибка сериализации JSON")
		return
	}
}

// getAllTasks GET-обработчик для получения списка ближайших задач
func (h *Handler) getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetAllTasks(r.Context())
	if err != nil {
		errorResponse(w, "не удалось получить список ближайших задач", err)
		return
	}

	response := models.Response{Tasks: tasks}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		logrus.Warn("ошибка сериализации JSON")
		return
	}
}

// getTaskId GET-обработчик для получения задачи по ее id
func (h *Handler) getTaskId(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	taskStruct, err := h.service.GetTaskId(r.Context(), id)
	if err != nil {
		errorResponse(w, "не удалось найти задачу по ее ID", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(taskStruct); err != nil {
		logrus.Warn("ошибка сериализации JSON")
		return
	}
}

// updateTaskId PUT-обработчик для редактирования задачи
func (h *Handler) updateTaskId(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errorResponse(w, "ошибка десериализации JSON", err)
		return
	}

	updateTask, err := h.service.UpdateTask(r.Context(), task)
	if err != nil {
		errorResponse(w, "не удалось отредактировать задачу", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updateTask); err != nil {
		logrus.Warn("ошибка сериализации JSON")
		return
	}
}

// taskDone POST-обработчик для выполнения задачи
func (h *Handler) taskDone(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if err := h.service.TaskDone(r.Context(), id); err != nil {
		errorResponse(w, "не удалось отметить задачу выполненной", err)
		return
	}

	response := struct{}{}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Warn("ошибка сериализации JSON")
		return
	}
}

// deleteTask DELETE-обработчик для удаления задачи
func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		errorResponse(w, "не удалось удалить задачу", err)
		return
	}

	response := struct{}{}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Warn("ошибка сериализации JSON")
		return
	}
}
