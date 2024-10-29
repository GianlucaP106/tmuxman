package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI wraps the tview application and holds a reference to the views.
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

const (
	modalName         = "modal"
	editorModalWidth  = 40
	editorModalHeight = 5
)

// Opens a cheatsheet
func (ui *UI) help(keys KeybdindingHolder) {
	t := newTable("Cheatsheet", []string{"Key", "Description"}, func(idx int) {})

	row := 1
	for _, binding := range keys {
		t.SetCell(row, 0, tview.NewTableCell(binding.key.display))
		t.SetCell(row, 1, tview.NewTableCell(binding.description))
		row++
	}

	c := center(t, 50, 20)
	ui.openModal(c)
}

// Opens a confirmation modal.
func (ui *UI) confirm(title string, done func(bool)) {
	// build view
	t := tview.NewTextView()
	t.SetTitle(surroundSpace(title))
	t.SetDynamicColors(true)

	// set the style
	t.SetBorder(true)
	t.SetBackgroundColor(tcell.ColorNone)
	t.SetBorderColor(tcell.ColorLightYellow)
	t.SetTitleColor(tcell.ColorBlue)

	// set the text
	t.SetText("\n[green]Enter[white] - Confim  |  [red]Escape[white] - Cancel")
	t.SetTextAlign(tview.AlignCenter)

	// set func to close on enter/esc
	t.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			// confirm
			done(true)
		case tcell.KeyEsc:
			// cancel
			done(false)
		}
		ui.closeModal()
	})

	// center with dimensions
	c := center(t, max(editorModalWidth, len(title)+6), editorModalHeight)
	ui.openModal(c)
}

// Opens a single line editor modal.
func (ui *UI) editor(title string, done func(string)) {
	// build input field
	i := tview.NewInputField()
	i.SetTitle(surroundSpace(title))

	// set styles
	i.SetFieldBackgroundColor(tcell.ColorDimGray)
	i.SetFieldTextColor(tcell.ColorWhite)
	i.SetBackgroundColor(tcell.ColorNone)
	i.SetBorder(true)
	i.SetBorderColor(tcell.ColorGreen)
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
	c := center(i, editorModalWidth, editorModalHeight)

	// open the editor
	ui.openModal(c)
}

// Opens a generic model around the passed primitive.
func (ui *UI) openModal(v tview.Primitive) {
	// open modal by adding a page
	ui.root.AddPage(modalName, v, true, true)
	ui.root.ShowPage(modalName)
}

// Closes the openned modal.
func (ui *UI) closeModal() {
	// close modal by deleting the page
	ui.root.RemovePage(modalName)
}

// Wrapper over tview table to hold titles and other utils.
type Table struct {
	*tview.Table
	colTitles []string
}

// Returns a new table with defaults and configs.
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
		t.SetBorderColor(tcell.ColorLightYellow)

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

// Internal method to set col titles.
func (t *Table) setColTitles() {
	for idx, title := range t.colTitles {
		t.SetCell(0, idx, tview.NewTableCell(title).SetTextColor(tcell.ColorWheat))
	}
}

// Gets the selected index.
func (t *Table) getSelected() (idx int) {
	row, _ := t.GetSelection()
	if row == 0 {
		return 0
	}
	return row - 1
}

// Overriding this method to reset the col titles.
func (t *Table) Clear() {
	t.Table.Clear()
	t.setColTitles()
}

// Type containing a key binding.
// Helpful for defining clean actions and for deriving cheatsheet.
type Keybinding struct {
	handler     func()
	key         *Key
	description string
}

// Key type for both handling event and displaying cheatsheet.
type Key struct {
	display string
	rune    rune
	key     tcell.Key
}

// KeybdindingHolder holds many keybindings.
type KeybdindingHolder []*Keybinding

// Handles a event by finidng the binding that mataches this event.
func (k KeybdindingHolder) handle(event *tcell.EventKey) {
	for _, binding := range k {
		if event.Key() == tcell.KeyRune {
			// if key is rune then we check the run
			if event.Rune() == binding.key.rune {
				binding.handler()
				return
			}
		} else if event.Key() == binding.key.key {
			// if key is not rune then we check the key
			binding.handler()
			return
		}
	}
}

// Centers a tview primitive.
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
