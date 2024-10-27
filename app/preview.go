package app

import (
	"fmt"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Preview struct {
	*tview.TextView
}

func (a *App) initPreview() {
	p := &Preview{}
	p.TextView = tview.NewTextView()
	p.SetBorder(true)
	p.SetTitle(surroundSpace("Preview"))
	p.SetDynamicColors(true)
	p.SetBackgroundColor(tcell.ColorNone)
	p.SetWrap(false)
	a.preview = p
}

func (p *Preview) update(pane *gotmux.Pane) {
	if pane == nil {
		return
	}

	// clear preview before
	p.Clear()

	// capture the contents of the current pane
	content, _ := pane.Capture()

	// write to the target view with a ansii writer
	ansiiWriter := tview.ANSIWriter(p)
	fmt.Fprintln(ansiiWriter, content)
}
