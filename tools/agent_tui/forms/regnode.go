package forms

import (
	"fmt"

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
	statusView.SetScrollable(true).SetWrap(true)

	goodConnectivity := false

	regNodeConfigForm := tview.NewForm().
		AddTextView(RENDEZVOUSLABEL, tuiNet.GetRendezvousHostIP(), 40, 1, true, false).
		AddButton(CONNECTIVITYCHECK, func() {
			statusView.Clear()
			fmt.Fprintln(statusView, "Running connectivity checks. Please wait...")
			go func() {
				_, err := tuiNet.CheckRegistryConnectivity()
				if err != nil {
					goodConnectivity = false
					fmt.Fprintf(statusView, "[red::b]Cannot reach release image at %s (%s)[black]\n", tuiNet.GetReleaseImageURL(), err)
				} else {
					goodConnectivity = true
					fmt.Fprintf(statusView, "Successfully reached release image at %s \n", tuiNet.GetReleaseImageURL())
				}

				_, err = tuiNet.CheckRendezvousHostConnectivity()
				if err != nil {
					goodConnectivity = false
					fmt.Fprintf(statusView, "[red::b]Failed to ping rendezvous host at %s (%s)[black]\n", tuiNet.GetRendezvousHostIP(), err)
				} else {
					goodConnectivity = true
					fmt.Fprintf(statusView, "Successfully pinged rendezvous host at %s \n", tuiNet.GetRendezvousHostIP())
				}

				if !goodConnectivity {
					fmt.Fprint(statusView, "[red::b]Connectivity checks failed [black] \n")
				} else {
					fmt.Fprint(statusView, "[green::b]Connectivity checks successful [black] \n")
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
		SetTitle("Agent-based Installer Network Connectivity Check").
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
