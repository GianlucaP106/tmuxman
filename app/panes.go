package app

import (
	"strconv"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) initPanesView() {
	// create table
	colTitles := []string{"Command", "PID", "Path", "Title", "Active"}
	t := newTable("Panes", colTitles, func(idx int) {
	})

	// set key bindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.GetSelection()
		k := event.Key()
		switch {
		// vim-like and if selection is last in list
		case k == tcell.KeyCtrlK || (isKeyUp(event) && row == 1):
			a.ui.SetFocus(a.ui.windows)
		}

		return event
	})

	a.ui.panes = t
}

func (a *App) syncPanes(window *gotmux.Window) {
	// get all panes for selected window
	a.state.panes, _ = window.ListPanes()

	// function to set row for given pane
	setRow := func(row int, p *gotmux.Pane) {
		t := a.ui.panes
		pid := strconv.Itoa(int(p.Pid))

		active := "No"
		if p.Active {
			active = "Yes"
		}

		path := trimStr(p.CurrentPath, STRING_LENGTH_LIMIT)
		t.SetCell(row, 0, tview.NewTableCell(p.CurrentCommand))
		t.SetCell(row, 1, tview.NewTableCell(pid))
		t.SetCell(row, 2, tview.NewTableCell(path))
		t.SetCell(row, 3, tview.NewTableCell(p.Title))
		t.SetCell(row, 4, tview.NewTableCell(active))
	}

	// clear the table
	a.ui.panes.Clear()

	// build table from panes
	for row, pane := range a.state.panes {
		setRow(row+1, pane)
	}
}
