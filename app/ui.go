package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI wraps the tview application and holds a reference to the views.
type UI struct {
	*tview.Application

	// refresh task handler
	refresher *Refresher

	// root component (allowing for modals and other functionalities)
	root *tview.Pages
}

const (
	modalName         = "modal"
	editorModalWidth  = 40
	editorModalHeight = 5
	stringLengthLimit = 30
)

func newUI() *UI {
	ui := &UI{}
	ui.refresher = newRefresher(1)
	ui.refresher.start()
	ui.Application = tview.NewApplication()
	return ui
}

// Opens a cheatsheet
func (ui *UI) help(keys KeybdindingHolder) {
	t := newTable("Cheatsheet", []string{"Key", "------------ Description ------------"}, func(v *Keybinding) {})
	t.setRows(keys, func(k *Keybinding) []*tview.TableCell {
		return []*tview.TableCell{
			tview.NewTableCell(k.key.display),
			tview.NewTableCell(k.description),
		}
	})

	t.SetDoneFunc(func(key tcell.Key) {
		ui.closeModal()
	})

	c := center(t, 50, 24)
	ui.openModal(c)
	ui.SetFocus(t)
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
func (ui *UI) editor(title string, defaultVal string, done func(string)) {
	// build input field
	i := tview.NewInputField()
	i.SetTitle(surroundSpace(title))

	// set styles
	i.SetFieldBackgroundColor(tcell.ColorDimGray)
	i.SetFieldTextColor(tcell.ColorWhite)
	i.SetBackgroundColor(tcell.ColorNone)
	i.SetBorder(true)
	i.SetBorderColor(tcell.ColorLightYellow)
	i.SetBorderPadding(1, 1, 1, 1)
	i.SetTitleColor(tcell.ColorLightSteelBlue)
	i.SetText(defaultVal)

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

// Queues a refresh task.
func (ui *UI) queue(task RefreshTask) {
	ui.refresher.refresh <- task
}

// Wrapper over tview table to hold titles and other utils.
type Table[T any] struct {
	*tview.Table
	colTitles []string
	values    []*T
}

// Returns a new table with defaults and configs.
func newTable[T any](title string, colTitles []string, onSelectionChanged func(v *T)) *Table[T] {
	// init table and set common attributes
	t := &Table[T]{}
	t.Table = tview.NewTable()
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
	t.setColTitles(colTitles)

	// the default selection is 1, since title is not selectable
	t.Select(1, 0)

	// set the selection changed function
	// this sets the default behavior and then calls the callback
	t.SetSelectionChangedFunc(func(row, column int) {
		// if the first row, reselect 1 (dont allow title selection)
		if row == 0 {
			t.Select(1, 0)
			return
		}

		// `row - 1` because the idx of the obj will start at 0
		idx := row - 1

		// run the call back
		onSelectionChanged(t.values[idx])
	})

	return t
}

// Internal method to set col titles.
func (t *Table[T]) setColTitles(titles []string) {
	t.colTitles = titles
	for idx, title := range t.colTitles {
		t.SetCell(0, idx, tview.NewTableCell(title).SetTextColor(tcell.ColorWheat))
	}
}

// Gets the selected index.
func (t *Table[T]) getSelected() *T {
	row, _ := t.GetSelection()
	if row == 0 {
		return nil
	}

	idx := row - 1
	if idx >= len(t.values) {
		return nil
	}

	return t.values[row-1]
}

// Sets all the rows in the table based on the values passed
func (t *Table[T]) setRows(values []*T, col func(*T) []*tview.TableCell) {
	t.Clear()
	for row, val := range values {
		cols := col(val)
		for col, cell := range cols {
			t.SetCell(row+1, col, cell)
		}
	}
	t.values = values
}

// Overriding this method to reset the col titles.
func (t *Table[T]) Clear() {
	t.Table.Clear()
	t.setColTitles(t.colTitles)
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

// Refresher is to handle refresh tasks that are sent to the channel.
type Refresher struct {
	refresh chan RefreshTask
	n       int
}

// Task to be ran in async.
type RefreshTask func()

// Creates a refresher
func newRefresher(n int) *Refresher {
	r := &Refresher{}
	r.n = n

	// set a buffer of 1000 as this will never get hit
	// if it it gets hit... deadlock. Alternative is variable length buffer queue.
	r.refresh = make(chan RefreshTask, 1000)
	return r
}

func (r *Refresher) start() {
	for i := 0; i < r.n; i++ {
		go func() {
			for task := range r.refresh {
				task()
			}
		}()
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

func setupTabs(tabs *tview.Pages, pages []string) {
	curPage := 0
	show := func(p int) {
		tabs.SwitchToPage(pages[p])
	}
	tabs.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		switch k {
		case tcell.KeyRight:
			curPage++
			if curPage >= len(pages) {
				curPage = 0
			}
			show(curPage)
		case tcell.KeyLeft:
			curPage--
			if curPage < 0 {
				curPage = len(pages) - 1
			}
			show(curPage)
		}
		return event
	})
}
