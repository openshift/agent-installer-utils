package agent_tui

import (
	"testing"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/ui"
)

func TestChecksPage(t *testing.T) {
	cases := []struct {
		name   string
		config checks.Config
		steps  func(app *AppTester)
	}{
		{
			name: "initial check failure with recover",
			steps: func(app *AppTester) {
				appConfig := checks.Config{
					ReleaseImageURL: "quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64",
					LogPath:         "/tmp/delete-me",
				}

				// initially the network isn't up yet, so the check will fail
				app.SetPullCheckError("Trying to pull quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64\nError: initializing source docker://quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64: can't talk to a V1 container registry\n")
				tester := app.Start(appConfig)
				tester.WaitForScreenContent("✖ quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64")

				// after a while, the network goes up, and the timeout dialog appears
				app.SetPullCheckOk()
				tester.WaitForScreenContent(
					"Agent-based installer",
					"connectivity checks passed",
					"This prompt will timeout")
			},
		},
		{
			name: "initial check failure with user interaction",
			steps: func(app *AppTester) {
				appConfig := checks.Config{
					ReleaseImageURL: "quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64",
					LogPath:         "/tmp/delete-me",
				}

				// initially the network isn't up yet, so the check will fail
				app.SetPullCheckError("Trying to pull quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64\nError: initializing source docker://quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64: can't talk to a V1 container registry\n")
				tester := app.Start(appConfig)
				tester.WaitForScreenContent("✖ quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64")

				// simulate the user touching the interface
				app.ScreenPressTab()

				// after a while, the network goes up, but since the user took control then the timeout dialog will not appear
				app.SetPullCheckOk()
				tester.WaitForScreenContent("✓ quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64")
			},
		},
		{
			name: "success, all checks pass, verify app switches to prompt with timeout",
			steps: func(app *AppTester) {
				appConfig := checks.Config{
					ReleaseImageURL: "quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64",
					LogPath:         "/tmp/delete-me",
				}
				tester := app.Start(appConfig)
				// all checks should pass and we should be prompted
				// with a message asking to continue "<Yes>" or
				// to quit "<No>"
				tester.WaitForScreenContent(
					"Agent-based installer",
					"connectivity checks passed",
					"This prompt will timeout")

				// after selecting "<Yes>" to continue with agent-tui
				// we should see the check screen again
				tester.SelectItem(ui.YES_BUTTON)
				tester.WaitForScreenContent(
					"Agent installer network boot setup",
					"✓ quay.io/openshift-release-dev/ocp-release:4.12.2-x86_64")
			},
		},
		{
			name: "release image not reachable",
			steps: func(app *AppTester) {
				appConfig := checks.Config{
					ReleaseImageURL: "localhost:8888/missing",
					LogPath:         "/tmp/delete-me",
				}
				app.SetPullCheckError("Trying to pull localhost:8888/missing\nError: initializing source docker://localhost:8888/missing: can't talk to a V1 container registry\n")
				app.SetHttpCheckError("no such host")
				app.SetDNSCheckError("dns error")

				tester := app.Start(appConfig)
				tester.WaitForScreenContent(
					"✖ localhost:8888/missing",
					"✖ nslookup localhost",
					"✓ ping localhost",
					"✖ localhost responds to http GET")

				// TODO: There is a limitation in apptester
				// where the full error details are not displayed
				// We may need to make the check details textview
				// contents accessible to the test framework and
				// assert against that. In the example above
				// the similuation screen contains:
				//
				// "nslookup failure:",
				// "Server:        127.0.0.1",
				//
				// but is missing:
				//
				// "Release image pull error:"
			},
		},
	}
	for _, tc := range cases {
		steps := tc.steps
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := NewAppTester(t)
			defer app.Stop()

			steps(app)
		})
	}
}
