package forms

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/net"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	CONNECTIVITYCHECK string = "Check connectivity"
	NETCONFIGURE      string = "Configure networking"
	DONE              string = "Done"
	RENDEZVOUSLABEL   string = "Rendezvous IP Address"
)

func RegNodeModalForm(app *tview.Application, pages *tview.Pages, validations *net.Validations) tview.Primitive {
	statusView := tview.NewTextView()

	statusView.SetBackgroundColor(newt.ColorGray).
		SetBorder(true).
		SetBorderColor(tcell.ColorBlack).
		SetTitle("Status").
		SetTitleColor(tcell.ColorBlack)
	statusView.SetTextColor(tcell.ColorBlack).
		SetDynamicColors(true)
	statusView.SetScrollable(true).SetWrap(true)

	regNodeConfigForm := tview.NewForm().
		AddTextView(RENDEZVOUSLABEL, validations.RendezvousHostIP, 40, 1, true, false).
		AddButton(CONNECTIVITYCHECK, func() {
			statusView.Clear()
			go func() {
				fmt.Fprintln(statusView, "Running connectivity checks. Please wait...")
				validations.CheckConnectivity()
				updateStatusView(statusView, validations)
				app.Draw()
			}()
		}). // TODO: Make the connectivity check screen
		AddButton(NETCONFIGURE, net.NMTUIRunner(app, pages, nil)).
		AddButton(DONE, func() {
			if !validations.HasConnectivityIssue() {
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

	// Prefill the status view if the initial validation checks performed when
	// the application started up indicated there is an issue.
	if validations.HasConnectivityIssue() {
		updateStatusView(statusView, validations)
	}

	// Change navigation. By default Tab moves through form buttons. Now,
	// * Left and Right keys moves through buttons in the form
	// * Tab and Back Tab moves to status view
	regNodeConfigForm.SetInputCapture(func(event *tcell.EventKey) (eventKey *tcell.EventKey) {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(statusView)
			eventKey = nil
		case tcell.KeyBacktab:
			app.SetFocus(statusView)
			eventKey = nil
		case tcell.KeyRight:
			if _, index := regNodeConfigForm.GetFocusedItemIndex(); index == (regNodeConfigForm.GetButtonCount() - 1) {
				eventKey = nil
			} else {
				app.SetFocus(regNodeConfigForm.GetButton(index + 1))
				eventKey = event
			}
		case tcell.KeyLeft:
			if _, index := regNodeConfigForm.GetFocusedItemIndex(); index == 0 {
				eventKey = nil
			} else {
				app.SetFocus(regNodeConfigForm.GetButton((index) - 1))
				eventKey = event
			}
		default:
			eventKey = event
		}
		return
	})

	// Register tab keys to switch to connectivity check form
	statusView.SetInputCapture(func(event *tcell.EventKey) (eventKey *tcell.EventKey) {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(regNodeConfigForm)
			eventKey = nil
		case tcell.KeyBacktab:
			app.SetFocus(regNodeConfigForm)
			eventKey = nil
		default:
			eventKey = event
		}
		return
	})

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

func updateStatusView(statusView *tview.TextView, validations *net.Validations) {
	goodConnectivity := true

	// fmt.Fprintln(statusView, "Running connectivity checks. Please wait...")

	if validations.ReleaseImagePullError != "" {
		goodConnectivity = false
		fmt.Fprintf(statusView, "[red]Cannot reach release image at %s\n", validations.ReleaseImageURL)
		fmt.Fprintf(statusView, "%s[black]\n", validations.ReleaseImagePullError)

		if validations.ReleaseImageDomainNameResolutionError != "" {
			fmt.Fprintf(statusView, "[red]nslookup release image host at %s failed: \n", validations.ReleaseImageDomainName)
			fmt.Fprintf(statusView, "%s[black]\n", validations.ReleaseImageDomainNameResolutionError)
		} else {
			fmt.Fprintf(statusView, "nslookup release image host at %s successful\n", validations.ReleaseImageDomainName)
		}

		if validations.ReleaseImageHostPingError != "" {
			fmt.Fprintf(statusView, "[red]ping release image host at %s failed: \n", validations.ReleaseImageDomainName)
			fmt.Fprintf(statusView, "%s[black]\n", validations.ReleaseImageHostPingError)
		} else {
			fmt.Fprintf(statusView, "ping release image host at %s successful\n", validations.ReleaseImageDomainName)
		}
	} else {
		fmt.Fprintf(statusView, "Successfully reached release image at %s \n", validations.ReleaseImageURL)
	}

	if validations.RendezvousHostPingError != "" {
		goodConnectivity = false
		fmt.Fprintf(statusView, "[red]Failed to ping rendezvous host at %s (%s)[black]\n", validations.RendezvousHostIP, validations.RendezvousHostPingError)
	} else {
		fmt.Fprintf(statusView, "Successfully pinged rendezvous host at %s \n", validations.RendezvousHostIP)
	}

	if goodConnectivity {
		fmt.Fprintf(statusView, "[green]Connectivity checks successful[black]\n")
	} else {
		fmt.Fprint(statusView, "[red]Connectivity checks failed[black]\n")
	}
}
