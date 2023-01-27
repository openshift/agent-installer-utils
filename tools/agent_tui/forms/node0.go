package forms

import (
	"encoding/json"

	"github.com/gdamore/tcell/v2"
	"github.com/nmstate/nmstate/rust/src/go/nmstate/v2"
	tuiNet "github.com/openshift/agent-installer-utils/tools/agent_tui/net"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	CONTINUE  string = "Continue Installation"
	CONFIGURE string = "Configure Networking"
)

func Node0Form(app *tview.Application, pages *tview.Pages) tview.Primitive {
	nm := nmstate.New()
	jsonNetState, err := nm.RetrieveNetState()
	if err != nil {
		panic(err)
	}

	var netState tuiNet.NetState
	if err = json.Unmarshal([]byte(jsonNetState), &netState); err != nil {
		panic(err)
	}

	ifaceTreeView, err := tuiNet.TreeView(netState, pages)
	if err != nil {
		panicDialog := tview.NewModal().
			SetText(err.Error()).
			AddButtons([]string{"Quit"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.Stop()
			})
		app.SetRoot(panicDialog, false)
		panic(err)
	}

	form := tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		AddButton(CONTINUE, func() {
			app.Stop()
		}).
		AddButton(CONFIGURE, tuiNet.NMTUIRunner(app, pages, ifaceTreeView))

	form.SetBackgroundColor(newt.ColorGray)

	node0Form := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(ifaceTreeView, 0, 4, true).
			AddItem(form, 3, 0, false).
			AddItem(nil, 0, 2, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	// Modify input capture of tree view to pass control to the buttons
	ifaceTreeView.SetInputCapture(func(event *tcell.EventKey) (eventKey *tcell.EventKey) {
		switch event.Key() {
		case tcell.KeyTab:
			form.SetFocus(0)
			app.SetFocus(form)
			eventKey = nil
		case tcell.KeyBacktab:
			form.SetFocus(form.GetButtonCount() - 1)
			app.SetFocus(form)
			eventKey = nil
		default:
			eventKey = event
		}
		return
	})

	// Modify form to go back to treeview
	form.SetInputCapture(func(event *tcell.EventKey) (eventKey *tcell.EventKey) {
		switch event.Key() {
		case tcell.KeyUp:
			app.SetFocus(ifaceTreeView)
			eventKey = nil
		case tcell.KeyTAB:
			if _, index := form.GetFocusedItemIndex(); index == (form.GetButtonCount() - 1) {
				app.SetFocus(ifaceTreeView)
				eventKey = nil
			} else {
				eventKey = event
			}
		case tcell.KeyBacktab:
			if _, index := form.GetFocusedItemIndex(); index == 0 {
				app.SetFocus(ifaceTreeView)
				eventKey = nil
			} else {
				eventKey = event
			}
		default:
			eventKey = event
		}
		return
	})

	return node0Form
}
