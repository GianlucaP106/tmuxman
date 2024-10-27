package app

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

func unixTime(t string) time.Time {
	activityInt, err := strconv.Atoi(t)
	if err != nil {
	}

	time := time.Unix(int64(activityInt), 0)
	return time
}

func timeFormat() string {
	return "02 Jan 15:04"
}

func trimStrBack(s string, n int) string {
	if len(s) <= n {
		return s
	}
	trimmed := s[len(s)-n:]
	return "..." + trimmed
}

func trimStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	trimmed := s[:n]
	return trimmed + "..."
}

func surroundSpace(s string) string {
	return " " + s + " "
}

func isKeyUp(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyUp || event.Rune() == 'k'
}

func isKeyDown(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyDown || event.Rune() == 'j'
}

var logger *log.Logger

func initLogger(path string) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panicln(err)
	}
	logger = log.New(f, "", 0)
}
