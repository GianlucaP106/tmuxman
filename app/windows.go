package app

import (
	"fmt"
	"strconv"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) initWindows() {
	// create table
	colTitles := []string{"ID", "Index", "Name", "Activity", "Active", "# Clients", "Size", "Cell Size"}
	t := newTable("Windows", colTitles, func(idx int) {
		w := a.getWindow()
		a.syncPanesDown(w)
	})

	t.SetSelectedFunc(func(row, column int) {
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
				display: "Enter",
			},
			description: "Attach to session",
			handler: func() {
				// get selected session and attach
				session := a.getSession()
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
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		k := event.Key()
		switch {
		// vim-like and if selection is last/first in list
		case k == tcell.KeyCtrlK || (isKeyUp(event) && row == 1):
			a.ui.SetFocus(a.ui.sessions)

		case k == tcell.KeyCtrlJ || (isKeyDown(event) && row == t.GetRowCount()-1):
			a.ui.SetFocus(a.ui.panes)
		}

		kh.handle(event)
		return event
	})

	a.ui.windows = t
}

func (a *App) syncWindowsView(session *gotmux.Session) {
	// get all windows for the selected session
	a.state.windows, _ = session.ListWindows()

	// define helper function to set row for given window
	setRow := func(row int, w *gotmux.Window) {
		t := a.ui.windows

		active := "No"
		if w.Active {
			active = "Yes"
		}

		activity := unixTime(w.Activity).Format(timeFormat())
		clients := strconv.Itoa(w.ActiveClients)
		idx := strconv.Itoa(w.Index)
		dimensions := fmt.Sprintf("%d x %d", w.Width, w.Height)
		cellDimensions := fmt.Sprintf("%d x %d", w.CellWidth, w.CellHeight)

		t.SetCell(row, 0, tview.NewTableCell(w.Id))
		t.SetCell(row, 1, tview.NewTableCell(idx))
		t.SetCell(row, 2, tview.NewTableCell(w.Name))
		t.SetCell(row, 3, tview.NewTableCell(activity))
		t.SetCell(row, 4, tview.NewTableCell(active))
		t.SetCell(row, 5, tview.NewTableCell(clients))
		t.SetCell(row, 6, tview.NewTableCell(dimensions))
		t.SetCell(row, 7, tview.NewTableCell(cellDimensions))
	}

	// clear table
	a.ui.windows.Clear()

	// build table from windows
	for row, w := range a.state.windows {
		setRow(row+1, w)
	}
}

func (a *App) getWindow() *gotmux.Window {
	idx := a.ui.windows.getSelected()
	// do a bounds check on the index
	if idx >= 0 && idx < len(a.state.windows) {
		// get selected window
		return a.state.windows[idx]
	}

	return nil
}
