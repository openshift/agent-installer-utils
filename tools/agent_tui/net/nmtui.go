package net

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/nmstate/nmstate/rust/src/go/nmstate/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui"
	"github.com/rivo/tview"
)

func NMTUIRunner(app *tview.Application, pages *tview.Pages, treeView *tview.TreeView) func() {
	return func() {
		app.Suspend(func() {
			cmd := exec.Command("nmtui")
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				agent_tui.PanicDialog(app, err)
			}
		})
		nm := nmstate.New()
		state, err := nm.RetrieveNetState()
		if err != nil {
			agent_tui.PanicDialog(app, err)
		}

		var netState NetState
		if err := json.Unmarshal([]byte(state), &netState); err != nil {
			agent_tui.PanicDialog(app, err)
		}

		//netStatePage, err := modalNetStateJSONPage(&netState, pages)
		if treeView == nil {
			netStatePage, err := ModalTreeView(netState, pages)
			if err != nil {
				agent_tui.PanicDialog(app, err)
			}
			pages.AddPage("netstate", netStatePage, true, true)
		} else {
			updatedTreeView, err := TreeView(netState, pages)
			if err != nil {
				agent_tui.PanicDialog(app, err)
			}
			treeView.SetRoot(updatedTreeView.GetRoot())
		}
	}
}
