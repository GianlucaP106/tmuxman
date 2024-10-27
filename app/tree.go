package app

import (
	"fmt"
	"strconv"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tree struct {
	*tview.TreeView
}

type TreeNode struct {
	value any
	typ   TreeNodeType
}

type TreeNodeType uint

const (
	Session TreeNodeType = iota
	Window
	Pane
)

func (a *App) initTree() {
	// instantiate tree view
	t := &Tree{}
	t.TreeView = tview.NewTreeView()

	// style
	t.SetBackgroundColor(tcell.ColorNone)
	t.SetTitle(surroundSpace("Tree"))
	t.SetGraphicsColor(tcell.ColorLightYellow)
	t.SetBorder(true)

	// build tree
	t.build()

	// set the root as the current node
	root := t.GetRoot()
	t.SetCurrentNode(root)
	root.CollapseAll().Expand()

	// keybindings
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
				key:     tcell.KeyRune,
				rune:    'w',
				display: "w",
			},
			description: "Toggle collapse/expand",
			handler: func() {
				// get cur node and invert its expanded
				cur := t.GetCurrentNode()
				cur.SetExpanded(!cur.IsExpanded())
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'D',
				display: "D",
			},
			description: "Kill item",
			handler: func() {
				// TODO: errror

				// get current node
				cur := t.GetCurrentNode()
				node := unwrapNode(cur)

				a.ui.confirm("Are you sure you want to kill this "+node.name()+" ?", func(b bool) {
					if !b {
						return
					}

					// handle all cases to kill
					switch node.typ {
					case Session:
						node.session().Kill()
					case Window:
						node.window().Kill()
					case Pane:
						node.pane().Kill()
					}

					t.sync()
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'a',
				display: "a",
			},
			description: "Create a new session",
			handler: func() {
				tmux, _ := gotmux.DefaultTmux()
				a.ui.editor("New session name", "", func(s string) {
					session, _ := tmux.NewSession(&gotmux.SessionOptions{
						Name: s,
					})

					a.ui.Suspend(func() {
						session.Attach()
					})

					t.sync()
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'r',
				display: "r",
			},
			description: "Rename this item (sessions and windows only)",
			handler: func() {
				cur := t.GetCurrentNode()
				node := unwrapNode(cur)
				if node.typ == Pane {
					return
				}

				var existingName string
				switch node.typ {
				case Session:
					existingName = node.session().Name
				case Window:
					existingName = node.window().Name
				}

				a.ui.editor("New "+node.name()+" name", existingName, func(s string) {
					switch node.typ {
					case Session:
						node.session().Rename(s)
					case Window:
						node.window().Rename(s)
					}

					t.sync()
				})
			},
		},
		{
			key: &Key{
				key:     tcell.KeyRune,
				rune:    'R',
				display: "R",
			},
			description: "Refresh",
			handler: func() {
				t.build()
				t.GetRoot().CollapseAll().Expand()
			},
		},
		{
			key: &Key{
				display: "Enter",
			},
			description: "Attach to item",
		},
	})

	// register the keybindings
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		kh.handle(event)
		return event
	})

	// set the Enter selected keybinding (enter)
	t.SetSelectedFunc(func(node *tview.TreeNode) {
		n := unwrapNode(node)
		switch n.typ {
		case Session:
			a.ui.Suspend(func() {
				n.session().Attach()
			})
		case Window:
			linkedSessions, _ := n.window().ListLinkedSessions()
			if len(linkedSessions) > 0 {
				a.ui.Suspend(func() {
					linkedSessions[0].Attach()
				})
			}
		case Pane:

		}
	})

	// set the changed function to display preview
	t.SetChangedFunc(func(node *tview.TreeNode) {
		// unwrape node
		n := unwrapNode(node)

		// cases for the node
		switch n.typ {
		case Session:
			windows, _ := n.session().ListWindows()
			if len(windows) == 0 {
				return
			}
			window := windows[0]
			panes, _ := window.ListPanes()
			if len(panes) == 0 {
				return
			}

			pane := panes[0]
			a.preview.update(pane)
		case Window:
			panes, _ := n.window().ListPanes()
			if len(panes) == 0 {
				return
			}
			pane := panes[0]
			a.preview.update(pane)
		case Pane:
			a.preview.update(n.pane())
		}
	})

	a.tree = t
}

// Builds tree from tmux data.
func (t *Tree) build() {
	// set the root node
	root := tview.NewTreeNode("sessions")
	t.SetRoot(root)
	root.SetSelectable(false)

	// get tmux client
	tmux, _ := gotmux.DefaultTmux()

	// get all sessions
	sessions, _ := tmux.ListSessions()

	for _, s := range sessions {
		// build tree sessionNode with session
		sessionNode := t.buildNode(s)

		// add the node to root
		root.AddChild(sessionNode)
	}
}

// Syncs tmux data to the tree and removes no longer existing items.
func (t *Tree) sync() {
	// get all sessions
	tmux, _ := gotmux.DefaultTmux()
	sessions, _ := tmux.ListSessions()

	// build a session map for quick access
	sessionMap := make(map[string]*gotmux.Session)
	for _, session := range sessions {
		sessionMap[session.Name] = session
	}

	// go over the sessions and verify if they exist
	root := t.GetRoot()
	for _, sessionNode := range root.GetChildren() {
		// if the session doesnt exist remove it and continue (dont loop into windows)
		internalSessionNode := unwrapNode(sessionNode)
		session := internalSessionNode.session()
		session = sessionMap[session.Name]
		if session == nil {
			root.RemoveChild(sessionNode)
			continue
		}

		// update session reference and text
		internalSessionNode.value = session
		sessionNode.SetText(internalSessionNode.title())

		// delete the session from the map to indicate that it is processed
		delete(sessionMap, session.Name)

		// build window map for this session
		windows, _ := session.ListWindows()
		windowMap := make(map[string]*gotmux.Window)
		for _, w := range windows {
			windowMap[w.Id] = w
		}

		// since this session still exists loop over children of if
		for _, windowNode := range sessionNode.GetChildren() {
			// if the window doesnt exist remove from this tree and continue
			internalWindowNode := unwrapNode(windowNode)
			window := internalWindowNode.window()
			window = windowMap[window.Id]
			if window == nil {
				sessionNode.RemoveChild(windowNode)
				continue
			}

			// update window ref and text
			internalWindowNode.value = window
			windowNode.SetText(internalWindowNode.title())

			// delete the window from the map to mark it as processed
			delete(windowMap, window.Id)

			// get panes for this window
			panes, _ := window.ListPanes()

			// build pane map
			paneMap := make(map[string]*gotmux.Pane)
			for _, p := range panes {
				paneMap[p.Id] = p
			}

			for _, paneNode := range windowNode.GetChildren() {
				// if this pane doesnot exist, remove it
				internalPaneNode := unwrapNode(paneNode)
				pane := internalPaneNode.pane()
				pane = paneMap[pane.Id]
				if pane == nil {
					windowNode.RemoveChild(paneNode)
					continue
				}

				// update pane ref and text
				internalPaneNode.value = pane
				paneNode.SetText(internalPaneNode.title())

				// delete pane from map to mark as processed
				delete(paneMap, pane.Id)
			}

			// add any pane remaining
			for _, p := range paneMap {
				windowNode.AddChild(newTreeNode(p, Pane))
			}
		}

		// add any remaining windows from the window map
		for _, w := range windowMap {
			sessionNode.AddChild(newTreeNode(w, Window))
		}
	}

	// post prcessing all existing sessions in the tree, add any that are missing
	for _, s := range sessionMap {
		root.AddChild(t.buildNode(s))
	}
}

func (t *Tree) buildNode(session *gotmux.Session) *tview.TreeNode {
	// build root (session)
	root := newTreeNode(session, Session)

	// build out children (windows)
	windows, _ := session.ListWindows()
	for _, w := range windows {
		windowNode := newTreeNode(w, Window)
		root.AddChild(windowNode)
		panes, _ := w.ListPanes()

		// build out panes
		for _, p := range panes {
			paneNode := newTreeNode(p, Pane)
			windowNode.AddChild(paneNode)
		}
	}

	return root
}

func unwrapNode(node *tview.TreeNode) *TreeNode {
	return node.GetReference().(*TreeNode)
}

func newTreeNode(v any, nodeType TreeNodeType) *tview.TreeNode {
	tn := &TreeNode{}
	tn.typ = nodeType
	tn.value = v
	wrapped := tview.NewTreeNode(tn.title())
	wrapped.SetReference(tn)
	wrapped.SetSelectable(true)
	switch nodeType {
	case Session:
		wrapped.SetColor(tcell.ColorBlue)
	case Window:
		wrapped.SetColor(tcell.ColorBlue)
	case Pane:
		wrapped.SetColor(tcell.ColorLightGrey)
	}
	return wrapped
}

// Gets the name of the tree node.
func (t *TreeNode) name() string {
	switch t.typ {
	case Session:
		return "session"
	case Window:
		return "window"
	case Pane:
		return "pane"
	}

	return ""
}

// Returns the title for the tree node.
func (t *TreeNode) title() string {
	var title string
	switch t.typ {
	case Session:
		s := t.session()
		time := unixTime(s.Activity).Format(timeFormat())
		title = fmt.Sprintf("(%s) - %s", time, s.Name)
	case Window:
		w := t.window()
		active := ""
		if w.Active {
			active = "(active)"
		}
		idx := strconv.Itoa(w.Index)
		title = fmt.Sprintf("%s - %s %s", idx, w.Name, active)
	case Pane:
		p := t.pane()
		active := ""
		if p.Active {
			active = "(active)"
		}
		title = fmt.Sprintf("%s %s", p.CurrentCommand, active)
	}

	return trimStr(" "+title, 60)
}

func (t *TreeNode) session() *gotmux.Session {
	session := t.value.(*gotmux.Session)
	return session
}

func (t *TreeNode) window() *gotmux.Window {
	window := t.value.(*gotmux.Window)
	return window
}

func (t *TreeNode) pane() *gotmux.Pane {
	pane := t.value.(*gotmux.Pane)
	return pane
}
