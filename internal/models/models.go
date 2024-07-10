package models

// Task структура задач.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Response структура отображения ответа.
type Response struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
	Tasks []Task `json:"tasks"`
}

const (
	DateFormat = "20060102"
	Limit      = 50
	WebDir     = "./web"
)
