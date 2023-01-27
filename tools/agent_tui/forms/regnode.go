package forms

import (
	"fmt"
	"net"

	"github.com/gdamore/tcell/v2"
	tuiNet "github.com/openshift/agent-installer-utils/tools/agent_tui/net"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	CONNECTIVITYCHECK string = "Check connectivity"
	NETCONFIGURE      string = "Configure networking"
	DONE              string = "Done"
	RENDEZVOUSLABEL   string = "Rendezvous IP Address"
)

func RegNodeModalForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	statusView := tview.NewTextView()

	statusView.SetBackgroundColor(newt.ColorGray).
		SetBorder(true).
		SetBorderColor(tcell.ColorBlack).
		SetTitle("Status").
		SetTitleColor(tcell.ColorBlack)
	statusView.SetTextColor(tcell.ColorBlack).
		SetDynamicColors(true)

	goodConnectivity := false
	ipField := tview.NewInputField().
		SetFieldWidth(40).
		SetLabel(RENDEZVOUSLABEL)

	ipField.SetDoneFunc(func(key tcell.Key) {
		if net.ParseIP(ipField.GetText()) == nil {
			statusView.SetText("[red::b]The specified Rendezvous IP is not a valid IP Address")
		} else {
			statusView.Clear()
		}
	})

	regNodeConfigForm := tview.NewForm().
		AddFormItem(ipField).
		AddButton(CONNECTIVITYCHECK, func() {
			statusView.Clear()
			fmt.Fprintf(statusView, "Running connectivity check. Please wait...\n")
			go func() {
				addr := ipField.GetText()
				output, err := tuiNet.ValidateConnectivity(addr)
				if err != nil {
					goodConnectivity = false
					fmt.Fprintf(statusView, "[red::b]Failed to connect to %s (%s)[black]\n%s", addr, err, tview.Escape(string(output)))
				} else {
					goodConnectivity = true
					statusView.SetText("[green::b]Connectivity check successful")
				}
				app.Draw()
			}()
		}). // TODO: Make the connectivity check screen
		AddButton(NETCONFIGURE, tuiNet.NMTUIRunner(app, pages, nil)).
		AddButton(DONE, func() {
			if goodConnectivity {
				app.Stop()
			} else {
				statusView.SetText("[red::b]Can't continue installation without a successful connectivity check")
			}
		})
	regNodeConfigForm.
		SetLabelColor(tcell.ColorBlack).
		SetBorder(true).
		SetTitle("Non installation orchestrating node configuration").
		SetTitleColor(tcell.ColorBlack).
		SetBackgroundColor(newt.ColorGray).
		SetBorderColor(tcell.ColorBlack)

	width := 80
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(regNodeConfigForm, 0, 1, true).
			AddItem(statusView, 0, 2, false).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}
