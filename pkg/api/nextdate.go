package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Определяет стандартный формат даты для приложения
const DateFormat = "20060102"

// NextDate вычисляет следующую дату задачи в зависимости от правила повторения
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("empty repeat rule")
	}

	startDate, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("invalid repeat rule")
	}

	rule := parts[0]
	nowStr := now.Format(DateFormat)

	switch rule {
	case "y":
		if len(parts) > 1 {
			return "", errors.New("invalid repeat rule for y")
		}
		date := startDate
		for {
			date = date.AddDate(1, 0, 0)
			if date.Format(DateFormat) > nowStr {
				return date.Format(DateFormat), nil
			}
		}

	case "d":
		if len(parts) != 2 {
			return "", errors.New("invalid repeat rule for d")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", errors.New("invalid days interval")
		}
		date := startDate
		for {
			date = date.AddDate(0, 0, days)
			if date.Format(DateFormat) > nowStr {
				return date.Format(DateFormat), nil
			}
		}

	case "w":
		if len(parts) != 2 {
			return "", errors.New("invalid repeat rule for w")
		}
		daysStr := strings.Split(parts[1], ",")
		var allowedDays []time.Weekday
		for _, ds := range daysStr {
			d, err := strconv.Atoi(ds)
			if err != nil || d < 1 || d > 7 {
				return "", errors.New("invalid weekday")
			}
			var wd time.Weekday
			if d == 7 {
				wd = time.Sunday
			} else {
				wd = time.Weekday(d)
			}
			allowedDays = append(allowedDays, wd)
		}

		date := startDate
		for {
			date = date.AddDate(0, 0, 1)
			if matchesWeekday(date, allowedDays) {
				if date.Format(DateFormat) > nowStr {
					return date.Format(DateFormat), nil
				}
			}
		}

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("invalid repeat rule for m")
		}
		daysStr := strings.Split(parts[1], ",")
		var allowedDays []int
		for _, ds := range daysStr {
			d, err := strconv.Atoi(ds)
			if err != nil || (d < 1 && d != -1 && d != -2) || d > 31 {
				return "", errors.New("invalid day of month")
			}
			allowedDays = append(allowedDays, d)
		}

		var allowedMonths []int
		if len(parts) == 3 {
			monthsStr := strings.Split(parts[2], ",")
			for _, ms := range monthsStr {
				m, err := strconv.Atoi(ms)
				if err != nil || m < 1 || m > 12 {
					return "", errors.New("invalid month")
				}
				allowedMonths = append(allowedMonths, m)
			}
		}

		date := startDate
		for {
			date = date.AddDate(0, 0, 1)
			if matchesMonthRule(date, allowedDays, allowedMonths) {
				if date.Format(DateFormat) > nowStr {
					return date.Format(DateFormat), nil
				}
			}
		}

	default:
		return "", errors.New("unknown repeat rule")
	}
}

// matchesWeekday проверяет, входит ли день недели даты в список разрешённых дней
func matchesWeekday(d time.Time, allowed []time.Weekday) bool {
	wd := d.Weekday()
	for _, a := range allowed {
		if wd == a {
			return true
		}
	}
	return false
}

// matchesMonthRule проверяет соответствие даты правилам по дням месяца и месяцам
func matchesMonthRule(d time.Time, allowedDays []int, allowedMonths []int) bool {
	if len(allowedMonths) > 0 {
		monthOk := false
		for _, m := range allowedMonths {
			if int(d.Month()) == m {
				monthOk = true
				break
			}
		}
		if !monthOk {
			return false
		}
	}

	day := d.Day()
	lastDay := time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, d.Location()).Day()

	for _, ad := range allowedDays {
		if ad > 0 {
			if day == ad {
				return true
			}
		} else if ad == -1 {
			if day == lastDay {
				return true
			}
		} else if ad == -2 {
			if day == lastDay-1 {
				return true
			}
		}
	}
	return false
}

// NextDateHandler обрабатывает запросы /api/nextdate.
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error
	if nowStr != "" {
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		now = time.Now()
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, nextDate)
}
