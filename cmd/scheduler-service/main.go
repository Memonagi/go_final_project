package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Memonagi/go_final_project/internal/handler"
	"github.com/Memonagi/go_final_project/tests"
	"github.com/go-chi/chi/v5"
)

const (
	webDir = "./web"
)

func main() {
	// получение значения переменной окружения
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = strconv.Itoa(tests.Port)
	}

	// создание маршрутизатора и обработка запросов
	r := chi.NewRouter()
	r.Handle("/*", http.FileServer(http.Dir(webDir)))

	// создание экземпляра Handler
	h := &handler.Handler{}

	// вычисление следующей даты
	r.Get("/api/nextdate", h.GetNextDate)
	// добавление задачи в БД
	r.MethodFunc(http.MethodPost, "/api/service", h.AddTask)
	// получение списка задач
	r.MethodFunc(http.MethodGet, "/api/tasks", h.GetAllTasks)
	// получение задачи по ее идентификатору
	r.MethodFunc(http.MethodGet, "/api/service", h.GetTaskId)
	// редактирование задачи
	r.MethodFunc(http.MethodPut, "/api/service", h.UpdateTaskId)
	// выполнение задачи
	r.MethodFunc(http.MethodPost, "/api/service/done", h.TaskDone)
	// удаление задачи
	r.MethodFunc(http.MethodDelete, "/api/service", h.DeleteTask)

	// запуск сервера
	log.Printf("запуск веб-сервера на порту %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		fmt.Println(err)
	}
}
