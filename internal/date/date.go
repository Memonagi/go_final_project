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
	errRule = errors.New("неверный формат правила")
	errDays = errors.New("указано неверное количество дней")
	errDate = errors.New("неверный формат даты")
)

// NextDate функция для определения следующей даты в соответствии с правилом.
func NextDate(now time.Time, dateString string, repeat string) (string, error) {
	var updateDate string

	if repeat == "" {
		return "", fmt.Errorf("%w", errRule)
	}

	date, err := time.Parse(models.DateFormat, dateString)
	if err != nil {
		return "", fmt.Errorf("%w", errDate)
	}

	repeatSlice := strings.Split(repeat, " ")
	rule := repeatSlice[0]

	switch rule {
	case "d":
		updateDate, err = dateByDay(now, date, repeatSlice)
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}
	case "y":
		updateDate, err = dateByYear(now, date, repeatSlice)
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}
	case "w":
		updateDate, err = dateByWeek(now, date, repeatSlice)
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}
	default:
		return "", fmt.Errorf("%w", errRule)
	}

	return updateDate, nil
}

// dateByDay правило повторения дней.
func dateByDay(now time.Time, date time.Time, repeatSlice []string) (string, error) {
	var updateDate time.Time

	if len(repeatSlice) != 2 {
		return "", fmt.Errorf("%w", errRule)
	}

	days, err := strconv.Atoi(repeatSlice[1])

	if err != nil || days < 1 || days > 400 {
		return "", fmt.Errorf("%w", errDays)
	}

	for {
		updateDate = date.AddDate(0, 0, days)

		if date.After(now) && !date.Equal(now) {
			break
		}
	}

	return updateDate.Format(models.DateFormat), nil
}

// dateByYear правило повторения лет.
func dateByYear(now time.Time, date time.Time, repeatSlice []string) (string, error) {
	var updateDate time.Time

	if len(repeatSlice) != 1 {
		return "", fmt.Errorf("%w", errRule)
	}

	for {
		updateDate = date.AddDate(1, 0, 0)

		if date.After(now) && !date.Equal(now) {
			break
		}
	}

	return updateDate.Format(models.DateFormat), nil
}

// dateByWeek правило повторения дней недели.
func dateByWeek(now time.Time, date time.Time, repeatSlice []string) (string, error) {
	var updateDate time.Time

	week, err := parseWeek(repeatSlice)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	for i, day := range week {
		if day == 7 {
			week[i] = 0
		}
	}

	updateDate = date.AddDate(0, 0, 1)

	for _, e := range week {
		if updateDate.Weekday() == time.Weekday(e) && updateDate.After(now) && !updateDate.Equal(now) {
			return updateDate.Format(models.DateFormat), nil
		}

		updateDate = date.AddDate(0, 0, 1)
	}

	return updateDate.Format(models.DateFormat), nil
}

func parseWeek(repeatSlice []string) ([]int, error) {
	if len(repeatSlice) != 2 {
		return nil, fmt.Errorf("%w", errRule)
	}

	wSlice := strings.Split(repeatSlice[1], ",")
	week := make([]int, 0, len(wSlice))

	for _, e := range wSlice {
		wDay, err := strconv.Atoi(e)
		if err != nil || wDay < 1 || wDay > 7 {
			return nil, fmt.Errorf("%w", errDays)
		}

		week = append(week, wDay)
	}

	return week, nil
}
