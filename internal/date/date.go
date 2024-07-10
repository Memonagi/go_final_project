package date

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Memonagi/go_final_project/internal/models"
)

var (
	errDate = errors.New("неверный формат даты")
	errDays = errors.New("указано неверное количество дней")
	errRule = errors.New("неверный формат правила")
)

// NextDate функция для определения следующей даты в соответствии с правилом (решено без учета повторения по месяцам).
func NextDate(now time.Time, dateString string, repeat string) (string, error) {
	if repeat == "" {
		return "", errRule
	}

	date, err := time.Parse(models.DateFormat, dateString)
	if err != nil {
		return "", errDate
	}

	repeatSlice := strings.Split(repeat, " ")
	rule := repeatSlice[0]

	switch rule {
	case "d":
		return dayRule(now, date, repeatSlice)
	case "y":
		return yearRule(now, date, repeatSlice)
	case "w":
		return weekRule(now, date, repeatSlice)
	default:
		return "", fmt.Errorf("%w", errRule)
	}
}

// dayRule проверяет правило повторения дней.
func dayRule(now time.Time, date time.Time, repeatSlice []string) (string, error) {
	if len(repeatSlice) != 2 {
		return "", errRule
	}

	days, err := strconv.Atoi(repeatSlice[1])
	if err != nil || days < 1 || days > 400 {
		return "", errDays
	}

	for {
		date = date.AddDate(0, 0, days)
		if date.After(now) && !date.Equal(now) {
			break
		}
	}

	return date.Format(models.DateFormat), nil
}

// yearRule проверяет правило повторения лет.
func yearRule(now time.Time, date time.Time, repeatSlice []string) (string, error) {
	if len(repeatSlice) != 1 {
		return "", fmt.Errorf("%w", errRule)
	}

	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) && !date.Equal(now) {
			break
		}
	}

	return date.Format(models.DateFormat), nil
}

// weekRule проверяет правило повторения дней недели.
func weekRule(now time.Time, date time.Time, repeatSlice []string) (string, error) {
	if len(repeatSlice) != 2 {
		return "", fmt.Errorf("%w", errRule)
	}

	wSlice := strings.Split(repeatSlice[1], ",")
	week := make([]int, 0, len(wSlice))

	for _, e := range wSlice {
		wDay, err := strconv.Atoi(e)
		if err != nil || wDay < 1 || wDay > 7 {
			return "", fmt.Errorf("%w", errDays)
		}

		week = append(week, wDay)
	}

	for i, day := range week {
		if day == 7 {
			week[i] = 0
		}
	}

	var err error

	date, err = weekDay(now, date, week)
	if err != nil {
		return "", err
	}

	return date.Format(models.DateFormat), nil
}

func weekDay(now time.Time, date time.Time, week []int) (time.Time, error) {
	date = date.AddDate(0, 0, 1)
	weekMap := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
		6: 6,
		0: 7,
	}

	for _, ok := weekMap[int(date.Weekday())]; ok; _, ok = weekMap[int(date.Weekday())] {
		for _, e := range week {
			if date.Weekday() == time.Weekday(e) && date.After(now) && !date.Equal(now) {
				return date, nil
			}
		}

		date = date.AddDate(0, 0, 1)
	}

	return time.Time{}, errRule
}
