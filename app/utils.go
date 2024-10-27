package app

import (
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
	return "Mon, 02 Jan 15:04:05"
}

func trimStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	trimmed := s[len(s)-n:]
	return "..." + trimmed
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
