package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Memonagi/go_final_project/internal/constants"
	"github.com/Memonagi/go_final_project/internal/date"
	"github.com/Memonagi/go_final_project/internal/service"
)

type Handler struct {
	task service.Task
}

// GetNextDate GET-обработчик для получения следующей даты
func (h *Handler) GetNextDate(w http.ResponseWriter, r *http.Request) {
	nowReq := r.FormValue("now")
	dateReq := r.FormValue("date")
	repeatReq := r.FormValue("repeat")

	now, err := time.Parse(constants.DateFormat, nowReq)
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

// AddTask POST-обработчик для добавления новой задачи
func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {

	err := json.NewDecoder(r.Body).Decode(&h.task)
	if err != nil {
		response := constants.Response{Error: "ошибка десериализации JSON"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	taskId, err := h.task.AddTask()
	if err != nil {
		response := constants.Response{Error: err.Error()}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := constants.Response{Id: taskId}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		response := constants.Response{Error: "ошибка сериализации JSON"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// GetAllTasks GET-обработчик для получения списка ближайших задач
func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {

	tasks, err := h.task.GetAllTasks()
	if err != nil {
		response := constants.Response{Error: err.Error()}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := constants.Response{Tasks: tasks}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		response := constants.Response{Error: "ошибка сериализации JSON"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// GetTaskId GET-обработчик для получения задачи по ее id
func (h *Handler) GetTaskId(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	taskStruct, err := h.task.GetTaskId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(taskStruct)
	if err != nil {
		response := constants.Response{Error: "ошибка сериализации JSON"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// UpdateTaskId PUT-обработчик для редактирования задачи
func (h *Handler) UpdateTaskId(w http.ResponseWriter, r *http.Request) {

	if err := h.task.UpdateTask(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	taskStruct := constants.Task{
		ID:      h.task.ID,
		Date:    h.task.Date,
		Title:   h.task.Title,
		Comment: h.task.Comment,
		Repeat:  h.task.Repeat,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(taskStruct); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TaskDone POST-обработчик для выполнения задачи
func (h *Handler) TaskDone(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	if err := h.task.DoneTask(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct{}{}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteTask DELETE-обработчик для удаления задачи
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if err := h.task.DeleteTask(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct{}{}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := constants.Response{Error: "ошибка сериализации JSON"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}
