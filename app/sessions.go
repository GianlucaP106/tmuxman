package app

import (
	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) initSessionsView() {
	// create the table, pass title, cols, and function that runs on table navigation
	colTitles := []string{"Name", "Last Attached", "Created"}
	t := newTable("Sessions", colTitles, func(idx int) {
		// bounds check on the idx
		if idx >= 0 && idx < len(a.state.sessions) {
			// get selected session and sync windows, panes, preview
			session := a.state.sessions[idx]
			a.syncWindowsDown(session)
		}
	})

	// when session is selected
	t.SetSelectedFunc(func(row, _ int) {
		// suspend the ui and attach the session
		s := a.state.sessions[row-1]
		a.ui.Suspend(func() {
			s.Attach()
		})
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		k := event.Key()
		r := event.Rune()
		switch {
		// vim-like and if selection is last in list
		case k == tcell.KeyCtrlJ || (isKeyDown(event) && row == t.GetRowCount()-1):
			a.ui.SetFocus(a.ui.windows)
		case r == '?':
			a.ui.confirm("hello", func(b bool) {
				if b {
					t.SetTitle("hello world this is set")
				}
			})
			// a.ui.editor("New session name", func(s string) {
			// 	t.SetTitle(s)
			// })
		}

		return event
	})

	a.ui.sessions = t
}

func (a *App) syncSessions() {
	// get all sessions
	a.state.sessions, _ = a.tmux.ListSessions()

	// function to set a row from a session, according the cols
	setRow := func(row int, s *gotmux.Session) {
		t := a.ui.sessions
		created := unixTime(s.Created).Format(timeFormat())
		lastAttached := unixTime(s.LastAttached).Format(timeFormat())
		name := trimStr(s.Name, STRING_LENGTH_LIMIT)
		t.SetCell(row, 0, tview.NewTableCell(name))
		t.SetCell(row, 1, tview.NewTableCell(lastAttached))
		t.SetCell(row, 2, tview.NewTableCell(created))
	}

	// clear the table
	a.ui.sessions.Clear()

	// build table from sessions
	for row, session := range a.state.sessions {
		setRow(row+1, session)
	}
}
