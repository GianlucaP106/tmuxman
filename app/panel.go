package app

import (
	"fmt"
	"log"
	"strconv"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Panel struct {
	*tview.Flex

	// panel tables
	sessions *Table[gotmux.Session]
	windows  *Table[gotmux.Window]
	panes    *Table[gotmux.Pane]
}

// Inits the panel (sessions, windows and panes table).
func (a *App) initPanel() {
	p := &Panel{}

	// inist the views in the panel
	p.initSessionsView(a)
	p.initWindows(a)
	p.initPanesView(a)

	// sync the data and update the preview
	pane := p.sync()
	a.preview.update(pane)

	// build and assemble panel with views
	p.Flex = tview.NewFlex()
	p.SetDirection(tview.FlexRow)
	p.AddItem(p.sessions, 0, 1, true)
	p.AddItem(p.windows, 0, 1, false)
	p.AddItem(p.panes, 0, 1, false)
	a.panel = p
}

func (p *Panel) initSessionsView(a *App) {
	// create the table, pass title, cols, and function that runs on table navigation
	colTitles := []string{"Name", "Last Attached", "Created"}
	t := newTable("Sessions", colTitles, func(s *gotmux.Session) {
		// sync windows down
		pane := p.syncWindowsDown(s)

		// update the preview
		a.preview.update(pane)
	})

	// when session is selected
	t.SetSelectedFunc(func(row, _ int) {
		// suspend the ui and attach the session
		s := p.sessions.getSelected()
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
						session := t.getSelected()
						session.Kill()
						pane := p.sync()
						a.preview.update(pane)
					}
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'r',
				display: "r",
			},
			description: "Rename session",
			handler: func() {
				session := t.getSelected()
				a.ui.editor("New session name", session.Name, func(s string) {
					session.Rename(s)
					pane := p.sync()
					a.preview.update(pane)
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'a',
				display: "a",
			},
			description: "Create new session",
			handler: func() {
				tmux, _ := gotmux.DefaultTmux()
				a.ui.editor("New session name", "", func(s string) {
					session, _ := tmux.NewSession(&gotmux.SessionOptions{
						Name: s,
					})

					a.ui.Suspend(func() {
						session.Attach()
					})

					pane := p.sync()
					a.preview.update(pane)
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				display: "space",
				rune:    ' ',
			},
			description: "Focus windows",
			handler: func() {
				a.ui.SetFocus(p.windows)
			},
		},
		{
			key: &Key{
				display: "Left/Right Arrow",
			},
			description: "Cycle views",
		},
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		if event.Key() == tcell.KeyCtrlJ || (isKeyDown(event) && row == t.GetRowCount()-1) {
			a.ui.SetFocus(p.windows)
			return event
		}

		kh.handle(event)
		return event
	})

	p.sessions = t
}

func (p *Panel) initWindows(a *App) {
	// create table
	colTitles := []string{"ID", "Index", "Name", "Activity", "Active", "# Clients", "Size", "Cell Size"}
	t := newTable("Windows", colTitles, func(w *gotmux.Window) {
		pane := p.syncPanesDown(w)
		a.preview.update(pane)
	})

	t.SetSelectedFunc(func(row, column int) {
		s := p.sessions.getSelected()
		a.ui.Suspend(func() {
			s.Attach()
		})
	})

	// defing keybindings
	var kh KeybdindingHolder
	kh = KeybdindingHolder([]*Keybinding{
		{
			key: &Key{
				display: "Enter",
			},
			description: "Attach to session",
			handler: func() {
				// get selected session and attach
				session := p.sessions.getSelected()
				session.Attach()
			},
		},
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
				key:     tcell.KeyRune,
				rune:    'r',
				display: "r",
			},
			description: "Rename window",
			handler: func() {
				cur := p.windows.getSelected()
				a.ui.editor("New window name", cur.Name, func(s string) {
					cur.Rename(s)
				})
				pane := p.sync()
				a.preview.update(pane)
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'D',
				display: "D",
			},
			description: "Kill window",
			handler: func() {
				cur := p.windows.getSelected()
				a.ui.confirm("Are you sure you want to kill this window?", func(b bool) {
					if !b {
						return
					}

					cur.Kill()
					pane := p.sync()
					a.preview.update(pane)
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyEsc,
				display: "esc",
			},
			description: "Go back",
			handler: func() {
				a.ui.SetFocus(p.sessions)
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				display: "space",
				rune:    ' ',
			},
			description: "Focus panes",
			handler: func() {
				a.ui.SetFocus(p.panes)
			},
		},
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		k := event.Key()
		switch {
		// vim-like and if selection is last/first in list
		case k == tcell.KeyCtrlK || (isKeyUp(event) && row == 1):
			a.ui.SetFocus(p.sessions)

		case k == tcell.KeyCtrlJ || (isKeyDown(event) && row == t.GetRowCount()-1):
			a.ui.SetFocus(p.panes)
		}

		kh.handle(event)
		return event
	})

	p.windows = t
}

func (p *Panel) initPanesView(a *App) {
	// create table
	colTitles := []string{"Command", "PID", "Path", "Title", "Active"}
	t := newTable("Panes", colTitles, func(p *gotmux.Pane) {
		a.preview.update(p)
	})

	var kh KeybdindingHolder
	kh = KeybdindingHolder([]*Keybinding{
		{
			key: &Key{
				display: "Enter",
			},
			description: "Attach to session",
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				display: "D",
				rune:    'D',
			},
			description: "Kill pane",
			handler: func() {
				a.ui.confirm("Are you sure you want to kill this pane?", func(b bool) {
					if !b {
						return
					}

					t.getSelected().Kill()
					p.sync()
				})
			},
		},
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
				key:     tcell.KeyEsc,
				display: "esc",
			},
			description: "Go back",
			handler: func() {
				a.ui.SetFocus(p.windows)
			},
		},
	})

	t.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.ui.Suspend(func() {
				p.sessions.getSelected().Attach()
			})
		}
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		k := event.Key()
		switch {
		// vim-like and if selection is last in list
		case k == tcell.KeyCtrlK || (isKeyUp(event) && row == 1):
			a.ui.SetFocus(p.windows)
		}

		kh.handle(event)
		return event
	})

	p.panes = t
}

// Syncs the entire panel with tmux.
// Returns the selected pane
func (p *Panel) sync() *gotmux.Pane {
	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		log.Panic(err)
	}

	// sync sessions
	sessions, _ := tmux.ListSessions()
	p.syncSessions(sessions)

	// if there is a selected session, sync the windows for it
	session := p.sessions.getSelected()
	if session != nil {
		return p.syncWindowsDown(session)
	}

	return nil
}

// Syncs the panel from windows down.
func (p *Panel) syncWindowsDown(session *gotmux.Session) *gotmux.Pane {
	windows, _ := session.ListWindows()
	p.syncWindows(windows)
	window := p.windows.getSelected()
	if window != nil {
		return p.syncPanesDown(window)
	}
	return nil
}

// Syncs the panel from the panes down (and preview).
func (p *Panel) syncPanesDown(window *gotmux.Window) *gotmux.Pane {
	panes, _ := window.ListPanes()
	p.syncPanes(panes)
	pane := p.panes.getSelected()
	return pane
}

// Syncs sessions to the table
func (p *Panel) syncSessions(sessions []*gotmux.Session) {
	// set rows on the table
	p.sessions.setRows(sessions, func(s *gotmux.Session) []*tview.TableCell {
		created := unixTime(s.Created).Format(timeFormat())
		lastAttached := unixTime(s.LastAttached).Format(timeFormat())
		name := trimStrBack(s.Name, stringLengthLimit)
		return []*tview.TableCell{
			tview.NewTableCell(name),
			tview.NewTableCell(lastAttached),
			tview.NewTableCell(created),
		}
	})
}

// Syncs the windows to the table.
func (p *Panel) syncWindows(windows []*gotmux.Window) {
	// set rows for windows
	p.windows.setRows(windows, func(w *gotmux.Window) []*tview.TableCell {
		active := "No"
		if w.Active {
			active = "Yes"
		}

		activity := unixTime(w.Activity).Format(timeFormat())
		clients := strconv.Itoa(w.ActiveClients)
		idx := strconv.Itoa(w.Index)
		dimensions := fmt.Sprintf("%d x %d", w.Width, w.Height)
		cellDimensions := fmt.Sprintf("%d x %d", w.CellWidth, w.CellHeight)
		return []*tview.TableCell{
			tview.NewTableCell(w.Id),
			tview.NewTableCell(idx),
			tview.NewTableCell(w.Name),
			tview.NewTableCell(activity),
			tview.NewTableCell(active),
			tview.NewTableCell(clients),
			tview.NewTableCell(dimensions),
			tview.NewTableCell(cellDimensions),
		}
	})
}

// Syncs panes to the table.
func (p *Panel) syncPanes(panes []*gotmux.Pane) {
	// set rows from panes
	p.panes.setRows(panes, func(pane *gotmux.Pane) []*tview.TableCell {
		pid := strconv.Itoa(int(pane.Pid))

		active := "No"
		if pane.Active {
			active = "Yes"
		}

		path := trimStrBack(pane.CurrentPath, stringLengthLimit)
		return []*tview.TableCell{
			tview.NewTableCell(pane.CurrentCommand),
			tview.NewTableCell(pid),
			tview.NewTableCell(path),
			tview.NewTableCell(pane.Title),
			tview.NewTableCell(active),
		}
	})
}
