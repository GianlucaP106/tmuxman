package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	*tview.Application

	// root component
	root *tview.Pages

	// panel tables
	sessions *Table
	windows  *Table
	panes    *Table

	// preview view
	preview *tview.TextView
}

const MODAL_NAME = "modal"

func (ui *UI) confirm(title string, done func(bool)) {
	m := tview.NewModal()
	m.SetText(surroundSpace(title))
	m.AddButtons([]string{"Enter - confirm | Esc - cancel"})
	m.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		done(buttonIndex > -1)
		ui.closeModal()
	})
	m.SetBackgroundColor(tcell.ColorBlack)
	m.SetBorderColor(tcell.ColorGreen)
	m.SetBorder(false)

	ui.openModal(m)
}

func (ui *UI) editor(title string, done func(string)) {
	// build input field and set styles
	i := tview.NewInputField()
	i.SetFieldBackgroundColor(tcell.ColorWhite)
	i.SetFieldTextColor(tcell.ColorBlack)
	i.SetBackgroundColor(tcell.ColorNone)
	i.SetBorder(true)
	i.SetBorderColor(tcell.ColorGreen)
	i.SetTitle(surroundSpace(title))
	i.SetBorderPadding(1, 1, 1, 1)
	i.SetTitleColor(tcell.ColorLightSteelBlue)

	// set acceptance function
	i.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		return true
	})

	// call the passed function in the done func
	i.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			done(i.GetText())
		}
		ui.closeModal()
	})

	// center the input
	c := center(i, 40, 5)

	// open the editor
	ui.openModal(c)
}

func (ui *UI) openModal(v tview.Primitive) {
	// open modal by adding a page
	ui.root.AddPage(MODAL_NAME, v, true, true)
	ui.root.ShowPage(MODAL_NAME)
}

func (ui *UI) closeModal() {
	// close modal by deleting the page
	ui.root.RemovePage(MODAL_NAME)
}

type Table struct {
	*tview.Table
	colTitles []string
}

// Returns a new table with defaults and configs
func newTable(title string, colTitles []string, onSelectionChanged func(idx int)) *Table {
	// init table and set common attributes
	t := &Table{}
	t.Table = tview.NewTable()
	t.colTitles = colTitles
	t.SetTitle(surroundSpace(title))
	t.SetBorder(true)

	// allow selecting the rows
	t.SetSelectable(true, false)

	// fix the title row
	t.SetFixed(1, 0)

	// set the default colors
	t.SetBackgroundColor(tcell.ColorNone)

	// set the border padding to be 1 around
	t.SetBorderPadding(1, 1, 1, 1)

	// make table show the borders
	t.SetBorders(true)
	t.SetBordersColor(tcell.ColorBlack)

	// set the callbacks for focused and unfocused (sets styles...)
	onFocus := func() {
		// focused colors
		s := tcell.StyleDefault.
			Background(tcell.ColorLightCyan).
			Foreground(tcell.ColorBlack)
		t.SetSelectedStyle(s)
		t.SetBorderColor(tcell.ColorGreen)

		// make table selectable when view is focused
		t.SetSelectable(true, false)

		// set title color when focused
		t.SetTitleColor(tcell.ColorLightSteelBlue)
	}
	onUnfocus := func() {
		// unfocused colors
		s := tcell.StyleDefault.
			Background(tcell.ColorNone).
			Foreground(tcell.ColorWhite)
		t.SetSelectedStyle(s)
		t.SetBorderColor(tcell.ColorWhite)

		// make table unselectable when unfosued
		// this is because of styling
		t.SetSelectable(false, false)

		// set title color back to default
		t.SetTitleColor(tcell.ColorWhite)
	}
	t.SetFocusFunc(onFocus)
	t.SetBlurFunc(onUnfocus)

	// by default, we set the un focused styles
	onUnfocus()

	// set the col titles
	t.setColTitles()

	// the default selection is 1, since title is not selectable
	t.Select(1, 0)

	// set the selection changed function
	// this sets the default behavior and then calls the callback
	t.SetSelectionChangedFunc(func(row, column int) {
		// if the first row, reselect 1 (dont allow title selection)
		if row == 0 {
			row = 1
			t.Select(row, 0)
		}

		// `row - 1` because the idx of the obj will start at 0
		idx := row - 1

		// run the call back
		onSelectionChanged(idx)
	})

	return t
}

// internal method to set col titles
func (t *Table) setColTitles() {
	for idx, title := range t.colTitles {
		t.SetCell(0, idx, tview.NewTableCell(title).SetTextColor(tcell.ColorWheat))
	}
}

// Overriding this method to reset the col titles
func (t *Table) Clear() {
	t.Table.Clear()
	t.setColTitles()
}

func center(p tview.Primitive, width, height int) *tview.Flex {
	// build a flex wrap over the passed view to center it
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}
