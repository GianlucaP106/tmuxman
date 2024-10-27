package app

import (
	"fmt"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) initPreviewView() {
	p := tview.NewTextView()
	p.SetBorder(true)
	p.SetTitle("Preview")
	p.SetDynamicColors(true)
	p.SetBackgroundColor(tcell.ColorNone)
	p.SetWrap(false)
	a.ui.preview = p
}

func (a *App) syncPreview(pane *gotmux.Pane) {
	// clear preview before
	a.ui.preview.Clear()

	// capture the contents of the current pane
	content, _ := pane.Capture()

	// write to the target view with a ansii writer
	ansiiWriter := tview.ANSIWriter(a.ui.preview)
	fmt.Fprintln(ansiiWriter, content)
}
