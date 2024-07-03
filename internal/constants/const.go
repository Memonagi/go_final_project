package constants

// Task структура задач
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Response структура отображения ответа
type Response struct {
	Id    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
	Tasks []Task `json:"tasks,omitempty"`
}

const (
	// DateFormat формат даты
	DateFormat = "20060102"
	// Limit лимит для отображения задач
	Limit = 50
)

// WeekMap мапа индексов дней недели
var WeekMap = map[int]int{
	1: 1,
	2: 2,
	3: 3,
	4: 4,
	5: 5,
	6: 6,
	0: 7,
}
