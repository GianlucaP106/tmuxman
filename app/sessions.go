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
		s := a.getSession()
		a.syncWindowsDown(s)
	})

	// when session is selected
	t.SetSelectedFunc(func(row, _ int) {
		// suspend the ui and attach the session
		s := a.getSession()
		a.ui.Suspend(func() {
			s.Attach()
		})
	})

	// defing keybindings
	var kh KeybdindingHolder
	kh = KeybdindingHolder([]*Keybinding{
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    '?',
				display: "?",
			},
			description: "Toggle cheatsheet",
			handler: func() {
				a.ui.help(kh)
			},
		},
		{
			key: &Key{
				display: "Enter",
			},
			description: "Attach to session",
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'D',
				display: "D",
			},
			description: "Kill session",
			handler: func() {
				a.ui.confirm("Are you sure you want to kill this session", func(b bool) {
					if b {
						idx := t.getSelected()
						session := a.state.sessions[idx]
						session.Kill()
						a.syncAll()
					}
				})
			},
		},
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		if event.Key() == tcell.KeyCtrlJ || (isKeyDown(event) && row == t.GetRowCount()-1) {
			a.ui.SetFocus(a.ui.windows)
			return event
		}

		kh.handle(event)
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

func (a *App) getSession() *gotmux.Session {
	idx := a.ui.sessions.getSelected()
	// bounds check
	if idx >= 0 && idx < len(a.state.sessions) {
		// get selected session
		return a.state.sessions[idx]
	}

	return nil
}
