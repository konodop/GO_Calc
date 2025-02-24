package agent

import (
	"strconv"
	"time"
)

type Task struct {
	ID             int           `json:"id"`
	Arg1           float64       `json:"arg1"`
	Arg2           float64       `json:"arg2"`
	Operation      rune          `json:"operation"`
	Operation_time time.Duration `json:"operation_time"`
}

type Data struct {
	ID     int    `json:"id"`
	Result string `json:"result"`
}

func Calc(task Task) Data {
	ans := ""
	lefted := task.Arg1
	righted := task.Arg2
	var q rune = ' '
	if q == '*' {
		ans = strconv.FormatFloat(lefted*righted, 'f', 6, 64)
	} else if q == '/' {
		ans = strconv.FormatFloat(lefted/righted, 'f', 6, 64)
	} else if q == '-' {
		ans = strconv.FormatFloat(lefted-righted, 'f', 6, 64)
	} else {
		ans = strconv.FormatFloat(lefted+righted, 'f', 6, 64)
	}
	f := Data{ID: task.ID, Result: ans}
	return f
}
