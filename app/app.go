package app

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	// panel, preview and tree views
	preview *Preview
	panel   *Panel
	tree    *Tree

	// ui instance, tview app
	ui *UI
}

func Start() {
	// instantiate app
	app := newApp()

	// init ui and build widget tree
	app.initUI()

	// run main loop
	if err := app.ui.Run(); err != nil {
		log.Panic(err)
	}
}

func newApp() *App {
	// instantiate app, state and tmux api
	app := &App{}
	return app
}

func (a *App) initUI() {
	// instantiate ui and tview app
	a.ui = newUI()

	// init the root views
	a.initPreview()
	a.initPanel()
	a.initTree()

	// build the tabs that can toggle between table and tree view
	tabs := tview.NewPages()
	pages := []string{"tree", "panel"}
	tabs.AddPage(pages[0], a.tree, true, false)
	tabs.AddPage(pages[1], a.panel, true, false)
	tabs.ShowPage(pages[0])
	tabs.SetBackgroundColor(tcell.ColorNone)
	setupTabs(tabs, pages)

	// build and set rootFlex view
	rootFlex := tview.NewFlex()
	rootFlex.SetTitle("Root")
	rootFlex.AddItem(tabs, 0, 1, true)
	rootFlex.AddItem(a.preview, 0, 2, false)

	// build page view as root to enable modals and other widgets
	root := tview.NewPages()
	root.AddPage("main", rootFlex, true, true)
	root.SetBackgroundColor(tcell.ColorNone)

	// set root
	a.ui.root = root
	a.ui.SetRoot(root, true)
}
