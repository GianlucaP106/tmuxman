package app

import (
	"log"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	// tmux client
	tmux *gotmux.Tmux

	// app state
	state *AppState

	// ui instance, tview app, and views
	ui *UI
}

type AppState struct {
	// current list of sessions/windows/panes
	sessions []*gotmux.Session
	windows  []*gotmux.Window
	panes    []*gotmux.Pane
}

const STRING_LENGTH_LIMIT = 30

func Start() {
	// instantiate app
	app := newApp()

	// init ui and build widget tree
	app.initUI()

	// sync all data to the views
	app.syncAll()

	// run main loop
	if err := app.ui.Run(); err != nil {
		log.Panic(err)
	}
}

func newApp() *App {
	// init tmux client
	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		log.Panic(err)
	}

	// instantiate app, state and tmux api
	app := &App{}
	app.state = &AppState{}
	app.tmux = tmux
	return app
}

func (a *App) initUI() {
	// instantiate ui and tview app
	a.ui = &UI{}
	a.ui.Application = tview.NewApplication()

	// build view container and individual views
	a.initSessionsView()
	a.initWindows()
	a.initPanesView()
	a.initPreviewView()

	// build and assemble panel with views
	panel := tview.NewFlex()
	panel.SetDirection(tview.FlexRow)
	panel.AddItem(a.ui.sessions, 0, 1, true)
	panel.AddItem(a.ui.windows, 0, 1, false)
	panel.AddItem(a.ui.panes, 0, 1, false)

	// build and set panelFlex view
	panelFlex := tview.NewFlex()
	panelFlex.SetTitle("Root")
	panelFlex.AddItem(panel, 0, 1, true)
	panelFlex.AddItem(a.ui.preview, 0, 2, false)

	// build page view as root to enable modals and other widgets
	root := tview.NewPages()
	root.AddPage("main", panelFlex, true, true)
	root.SetBackgroundColor(tcell.ColorNone)

	// set root
	a.ui.root = root
	a.ui.SetRoot(root, true)
}

func (a *App) syncAll() {
	// sync sessions
	a.syncSessions()

	// if there are sessions, sync the windows for the first one
	if len(a.state.sessions) > 0 {
		session := a.state.sessions[0]
		a.syncWindowsView(session)
	}

	// do the same for panes
	if len(a.state.windows) > 0 {
		window := a.state.windows[0]
		a.syncPanes(window)
	}

	// do the same for preview
	if len(a.state.panes) > 0 {
		pane := a.state.panes[0]
		a.syncPreview(pane)
	}
}

func (a *App) syncWindowsDown(session *gotmux.Session) {
	// sync windows first
	a.syncWindowsView(session)

	// if there are some windows then sync panes
	if len(a.state.windows) > 0 {
		window := a.state.windows[0]
		a.syncPanes(window)
	}

	// same for preview
	if len(a.state.panes) > 0 {
		pane := a.state.panes[0]
		a.syncPreview(pane)
	}
}

func (a *App) syncPanesDown(window *gotmux.Window) {
	// sync panes first
	a.syncPanes(window)

	// sync preview if there are panes
	if len(a.state.panes) > 0 {
		pane := a.state.panes[0]
		a.syncPreview(pane)
	}
}
