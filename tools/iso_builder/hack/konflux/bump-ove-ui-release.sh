#!/usr/bin/env bash
# Bump the OCP release embedded in the OVE UI agent ISO build on an
# installer-ove-ui-<X.Y> branch of ocp-build-data.
# Follows the pattern of https://github.com/openshift-eng/ocp-build-data/pull/13690
#
# Run it from the top of a clone of your fork of ocp-build-data, where
# 'origin' is your fork; the branch is pushed there and the PR is opened
# from it. The base branch always comes from upstream, so the fork doesn't
# need to be up to date. Needs an authenticated 'gh' for the PR steps.
#
# Usage: bump-ove-ui-release.sh [options] <X.Y.Z>
#   e.g. bump-ove-ui-release.sh 4.22.21
#        bump-ove-ui-release.sh 5.0.3     (uses branch installer-ove-ui-5.0)
#
# Options:
#   -C <dir>   ocp-build-data fork clone (default: current directory)
#   -p         push the branch to origin (your fork)
#   -r         push and open the PR against openshift-eng/ocp-build-data
#   -n         dry run: show the diff, don't commit
#   -h         help
set -euo pipefail

UPSTREAM_REPO="openshift-eng/ocp-build-data"
UPSTREAM_URL="https://github.com/${UPSTREAM_REPO}.git"
ARCH="x86_64"

repo_dir=.
push=false
open_pr=false
dry_run=false

usage() { sed -n '2,/^set -euo/p' "$0" | sed '$d; s/^# \{0,1\}//'; exit "${1:-0}"; }
die() { echo "error: $*" >&2; exit 1; }

while getopts "C:prnh" opt; do
    case $opt in
        C) repo_dir=$OPTARG ;;
        p) push=true ;;
        r) push=true; open_pr=true ;;
        n) dry_run=true ;;
        h) usage 0 ;;
        *) usage 1 ;;
    esac
done
shift $((OPTIND - 1))
[[ $# -eq 1 ]] || usage 1

version=$1
[[ $version =~ ^([0-9]+\.[0-9]+)\.([0-9]+)$ ]] || die "version must look like 4.22.21, got '$version'"
major_minor=${BASH_REMATCH[1]}
patch=${BASH_REMATCH[2]}

base_branch="installer-ove-ui-${major_minor}"
new_branch="bump-ove-ui-release-${version}"
release_value="quay.io/openshift-release-dev/ocp-release:${version}-${ARCH}"

cd "$repo_dir" || die "cannot cd to $repo_dir"
[[ $(git rev-parse --show-toplevel 2>/dev/null) == "$(pwd -P)" && $(git remote get-url origin 2>/dev/null) == */ocp-build-data* ]] \
    || die "$(pwd) is not the top of an ocp-build-data fork clone (origin = your fork); cd there or use -C"
[[ $(git remote get-url origin) != *"/${UPSTREAM_REPO}"* ]] \
    || die "origin is ${UPSTREAM_REPO} itself; it must be your fork"
git diff --quiet && git diff --cached --quiet || die "$(pwd) has uncommitted changes"
git show-ref --quiet "refs/heads/${new_branch}" && die "branch ${new_branch} already exists"

# Always branch from upstream so a stale fork doesn't matter.
echo "Fetching ${base_branch} from ${UPSTREAM_REPO}..."
git fetch -q "$UPSTREAM_URL" "refs/heads/${base_branch}" \
    || die "no ${base_branch} branch upstream; it must exist before the first ${major_minor} bump"
git checkout -q -b "$new_branch" FETCH_HEAD

old_patch=$(sed -n 's/^  - "PATCH_VERSION=\([0-9]*\)"$/\1/p' group.yml)
[[ -n $old_patch ]] || die "PATCH_VERSION not found in group.yml"
old_version="${major_minor}.${old_patch}"
[[ $old_patch != "$patch" ]] || { git checkout -q -; git branch -q -D "$new_branch"; die "group.yml is already at ${version}"; }

sed -i \
    -e "s|^  - \"RELEASE_VALUE=.*\"$|  - \"RELEASE_VALUE=${release_value}\"|" \
    -e "s|^  - \"PATCH_VERSION=.*\"$|  - \"PATCH_VERSION=${patch}\"|" \
    group.yml

git --no-pager diff
if [[ $(git diff --numstat group.yml | awk '{print $1}') != 2 ]]; then
    git checkout -q group.yml; git checkout -q -; git branch -q -D "$new_branch"
    die "expected exactly 2 changed lines in group.yml; aborting"
fi

if $dry_run; then
    git checkout -q group.yml; git checkout -q -; git branch -q -D "$new_branch"
    echo "Dry run: no commit made."
    exit 0
fi

title="Bump OVE UI ISO embedded OCP release to ${version}"
git commit -q -a -F - <<EOF
${title}

Update RELEASE_VALUE and PATCH_VERSION in group.yml konflux build_args
for the next agent-ove.${ARCH}.iso build on ${base_branch}.
EOF
echo "Committed on branch ${new_branch}:"
git --no-pager log --oneline -1

# Link the previous bump PR in the description, if we can find it.
prev_pr=$(gh pr list -R "$UPSTREAM_REPO" --base "$base_branch" --state merged \
    --search "\"Bump OVE UI ISO embedded OCP release to ${old_version}\" in:title" \
    --json number --jq '.[0].number // empty' 2>/dev/null || true)
followup="Follow-up to the ${old_version} bump${prev_pr:+ (#${prev_pr})}."

body=$(cat <<EOF
## Summary
- Update \`konflux.build_args\` in \`group.yml\` to embed OCP **${version}** in the OVE agent installer ISO build
  - Changes \`RELEASE_VALUE\` to \`${release_value}\` and \`PATCH_VERSION\` to \`${patch}\`

## Context
Prepares the \`${base_branch}\` branch for a new \`art-agent-installer-iso\` Konflux/Jenkins build.

${followup}

## Test plan
- [ ] Merge PR
- [ ] Trigger layered-products build with \`GROUP=${base_branch}\`, \`ASSEMBLY=stream\`, \`IMAGE_LIST=art-agent-installer-iso\`, and this branch as \`DOOZER_DATA_GITREF\`
- [ ] Confirm build embeds OCP ${version} in the ISO
EOF
)

if $push; then
    git push -u origin "$new_branch"
fi

fork_owner=$(git remote get-url origin | sed -E 's#.*[:/]([^/]+)/[^/.]+(\.git)?$#\1#')
pr_cmd=(gh pr create -R "$UPSTREAM_REPO" --base "$base_branch" --head "${fork_owner}:${new_branch}" --title "$title")

if $open_pr; then
    "${pr_cmd[@]}" --body "$body"
else
    body_file="${TMPDIR:-/tmp}/${new_branch}-pr-body.md"
    printf '%s\n' "$body" > "$body_file"
    printf '\nPR title:\n%s\n\nPR description (saved to %s):\n%s\n\n' "$title" "$body_file" "$body"
    echo "To open the PR:"
    $push || echo "  git push -u origin ${new_branch}"
    printf '  %s --body-file %q\n' "$(printf '%q ' "${pr_cmd[@]}")" "$body_file"
fi
