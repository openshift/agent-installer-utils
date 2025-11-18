#!/bin/bash
# Script to update Konflux release version for OCP
# Usage: ./update-konflux-release.sh <version>
# Example: ./update-konflux-release.sh 4.20.2

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 4.20.2"
    exit 1
fi

VERSION="$1"
RELEASE_URL="https://amd64.ocp.releases.ci.openshift.org/releasestream/4-stable/release/${VERSION}"

echo "This script will:"
echo "1. Fetch the release page and extract the quay.io URL"
echo "2. Parse version numbers from the URL"
echo "3. Update the two .tekton YAML files"
echo "4. Create a PR to the release-4.20 branch"
echo "5. Provide Jira card details"
echo ""
echo "Version: $VERSION"
echo "Release URL: $RELEASE_URL"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

echo "Starting Claude Code to run the automation..."
echo "Please provide this prompt to Claude:"
echo ""
echo "---"
echo "Please update the Konflux release version using this release URL: $RELEASE_URL"
echo ""
echo "Follow these steps:"
echo "1. Fetch the release page and extract the quay.io URL (e.g., quay.io/openshift-release-dev/ocp-release:X.Y.Z-x86_64)"
echo "2. Parse version numbers: major-minor (e.g., 4.20) and patch (e.g., 2)"
echo "3. Ensure you're working from upstream/release-4.20 branch"
echo "4. Update these 2 files with the new values:"
echo "   - .tekton/ove-ui-iso-4-20-pull-request.yaml"
echo "   - .tekton/ove-ui-iso-4-20-push.yaml"
echo "   Update params: release-value, major-minor-version, patch-version"
echo "5. Push the branch and create a PR:"
echo "   After Claude creates the commit, push the branch with:"
echo "   git push -u origin <branch-name>"
echo ""
echo "   IMPORTANT: Create the PR to UPSTREAM, not your fork!"
echo "   Option 1 - Use this URL format:"
echo "   https://github.com/openshift/agent-installer-utils/compare/release-4.20...<your-github-username>:agent-installer-utils:<branch-name>"
echo ""
echo "   Option 2 - Navigate manually:"
echo "   a. Go to https://github.com/openshift/agent-installer-utils"
echo "   b. Click 'Pull requests'"
echo "   c. Click 'New pull request'"
echo "   d. Click 'compare across forks'"
echo "   e. Set base repository: openshift/agent-installer-utils, base: release-4.20"
echo "   f. Set head repository: <your-fork>/agent-installer-utils, compare: <branch-name>"
echo ""
echo "6. Generate Jira card details:"
echo "   - Project: OCPBUGS"
echo "   - Issue Type: Bug"
echo "   - Summary: Update Konflux release version"
echo "   - Component: Installer/Agent based installation"
echo "   - Target Version: 4.20.z"
echo "   - Description: Claude generated bug to update the OCP version for Konflux builds."
echo "---"
