package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"time"

	"github.com/Memonagi/go_final_project/internal/date"
	"github.com/Memonagi/go_final_project/internal/models"
	"github.com/Memonagi/go_final_project/internal/service"
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

	// создание экземпляра Handler
	h := Handler{
		service: service,
		server: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
		port: port,
	}

	// вычисление следующей даты
	r.Get("/api/nextdate", h.getNextDate)
	// добавление задачи в БД
	r.MethodFunc(http.MethodPost, "/api/service", h.addTask)
	// получение списка задач
	r.MethodFunc(http.MethodGet, "/api/tasks", h.getAllTasks)
	// получение задачи по ее идентификатору
	r.MethodFunc(http.MethodGet, "/api/service", h.getTaskId)
	// редактирование задачи
	r.MethodFunc(http.MethodPut, "/api/service", h.updateTaskId)
	// выполнение задачи
	r.MethodFunc(http.MethodPost, "/api/service/done", h.taskDone)
	// удаление задачи
	r.MethodFunc(http.MethodDelete, "/api/service", h.deleteTask)

	return &h
}

// Run запускает сервер
func (h *Handler) Run() error {

	log.Printf("запуск веб-сервера на порту %d", h.port)
	if err := h.server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
	return nil
}

// errorResponse возвращает ошибку в формате {"error":"текст ошибки"}
func errorResponse(w http.ResponseWriter, errorText string, err error) {
	errorResponse := models.Response{
		Error: fmt.Errorf("%s: %w", errorText, err).Error()}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nextDate, err := date.NextDate(now, dateReq, repeatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))

	if err != nil {
		http.Error(w, fmt.Errorf("writing tasks data error: %w", err).Error(), http.StatusBadRequest)
	}
}

// addTask POST-обработчик для добавления новой задачи
func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		errorResponse(w, "ошибка десериализации JSON", err)
		return
	}

	taskId, err := h.service.AddTask(task)
	if err != nil {
		errorResponse(w, err.Error(), err)
		return
	}

	response := models.Response{Id: taskId}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errorResponse(w, "ошибка сериализации JSON", err)
		return
	}
}

// getAllTasks GET-обработчик для получения списка ближайших задач
func (h *Handler) getAllTasks(w http.ResponseWriter, r *http.Request) {

	tasks, err := h.service.GetAllTasks()
	if err != nil {
		errorResponse(w, err.Error(), err)
		return
	}

	response := models.Response{Tasks: tasks}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errorResponse(w, "ошибка сериализации JSON", err)
		return
	}
}

// getTaskId GET-обработчик для получения задачи по ее id
func (h *Handler) getTaskId(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	taskStruct, err := h.service.GetTaskId(id)
	if err != nil {
		errorResponse(w, err.Error(), err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(taskStruct)
	if err != nil {
		errorResponse(w, err.Error(), err)
		return
	}
}

// updateTaskId PUT-обработчик для редактирования задачи
func (h *Handler) updateTaskId(w http.ResponseWriter, r *http.Request) {

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		errorResponse(w, "ошибка десериализации JSON", err)
		return
	}

	updateTask, err := h.service.UpdateTask(task)
	if err != nil {
		errorResponse(w, err.Error(), err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updateTask); err != nil {
		errorResponse(w, "ошибка сериализации JSON", err)
		return
	}
}

// taskDone POST-обработчик для выполнения задачи
func (h *Handler) taskDone(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	if err := h.service.TaskDone(id); err != nil {
		errorResponse(w, err.Error(), err)
		return
	}

	response := struct{}{}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errorResponse(w, "ошибка сериализации JSON", err)
		return
	}
}

// deleteTask DELETE-обработчик для удаления задачи
func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	if err := h.service.DeleteTask(id); err != nil {
		errorResponse(w, err.Error(), err)
		return
	}

	response := struct{}{}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errorResponse(w, "ошибка сериализации JSON", err)
		return
	}
}
