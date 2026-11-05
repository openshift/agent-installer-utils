#!/usr/bin/env bash
# After the OVE UI agent ISO build for a new OCP release is done, update
# konflux-release-data to release it: bump the ISO filename and
# productVersionName in the prod/stage ReleasePlanAdmissions and add a
# developer portal file for the new version.
# Follows the pattern of
# https://gitlab.cee.redhat.com/releng/konflux-release-data/-/merge_requests/23685
#
# Run it from the top of a clone of releng/konflux-release-data (not a fork;
# the project must not be forked), where 'origin' is the upstream project and
# you can push branches. The branch is created from origin/main.
#
# Usage: konflux-release-mr.sh [options] <X.Y.Z>
#   e.g. konflux-release-mr.sh 4.22.21
#        konflux-release-mr.sh 5.0.3      (uses the ...-5-0-{prod,stage} RPAs)
#
# Options:
#   -C <dir>         konflux-release-data clone (default: current directory)
#   -d <YYYY-MM-DD>  releaseDate for the portal file (default: today)
#   -t               run 'tox -e yamllint,test' before committing
#   -p               push the branch and create the MR. If $GITLAB_TOKEN is set
#                    (a personal access token with 'api' scope) the MR is created
#                    via the GitLab API; otherwise the branch is pushed and a link
#                    to a pre-filled MR form is printed.
#   -n               dry run: show the changes, don't commit
#   -h               help
set -euo pipefail

GITLAB_HOST="gitlab.cee.redhat.com"
PROJECT_PATH="releng/konflux-release-data"
TARGET_BRANCH="main"
RPA_DIR="config/kflux-ocp-p01.7ayg.p1/product/ReleasePlanAdmission/art-installer-agent"
PORTAL_DIR="data/external/developer-portal/assisted-installer-disconnected"
ARCH="x86_64"

repo_dir=.
release_date=$(date +%F)
run_tox=false
push=false
dry_run=false

usage() { sed -n '2,/^set -euo/p' "$0" | sed '$d; s/^# \{0,1\}//'; exit "${1:-0}"; }
die() { echo "error: $*" >&2; exit 1; }

while getopts "C:d:tpnh" opt; do
    case $opt in
        C) repo_dir=$OPTARG ;;
        d) release_date=$OPTARG ;;
        t) run_tox=true ;;
        p) push=true ;;
        n) dry_run=true ;;
        h) usage 0 ;;
        *) usage 1 ;;
    esac
done
shift $((OPTIND - 1))
[[ $# -eq 1 ]] || usage 1

version=$1
[[ $version =~ ^([0-9]+)\.([0-9]+)\.[0-9]+$ ]] || die "version must look like 4.22.21, got '$version'"
major_minor="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}"
major_minor_dash="${BASH_REMATCH[1]}-${BASH_REMATCH[2]}"
[[ $release_date =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] && date -d "$release_date" >/dev/null 2>&1 \
    || die "release date must be YYYY-MM-DD, got '$release_date'"

new_branch="abi-${version}"
rpa_files=(
    "${RPA_DIR}/art-agent-installer-iso-${major_minor_dash}-prod.yaml"
    "${RPA_DIR}/art-agent-installer-iso-${major_minor_dash}-stage.yaml"
)
portal_file="${PORTAL_DIR}/${version}.yaml"

cd "$repo_dir" || die "cannot cd to $repo_dir"
[[ -d $RPA_DIR && -d $PORTAL_DIR ]] && git rev-parse --git-dir >/dev/null 2>&1 \
    || die "$(pwd) is not a konflux-release-data clone; run from its top directory or use -C"
[[ $(git remote get-url origin) == *"${GITLAB_HOST}/${PROJECT_PATH}"* ]] \
    || die "origin must be ${GITLAB_HOST}/${PROJECT_PATH} (branches go to upstream, not a fork)"
git diff --quiet && git diff --cached --quiet || die "$(pwd) has uncommitted changes"
git show-ref --quiet "refs/heads/${new_branch}" && die "local branch ${new_branch} already exists"

echo "Fetching ${TARGET_BRANCH} from origin..."
git fetch -q origin "$TARGET_BRANCH"
git ls-remote --exit-code --heads origin "$new_branch" >/dev/null && die "branch ${new_branch} already exists on origin"
start_branch=$(git rev-parse --abbrev-ref HEAD)
git checkout -q -b "$new_branch" "origin/${TARGET_BRANCH}"

abort() {
    git reset -q --hard
    git clean -q -f -- "$PORTAL_DIR"
    git checkout -q "$start_branch"
    git branch -q -D "$new_branch"
    [[ $# -eq 0 ]] || die "$*"
}

for f in "${rpa_files[@]}"; do
    [[ -f $f ]] || abort "$f not found; the ${major_minor} release config (RPAs) must be added first"
done

old_version=$(yq '.spec.data.mapping.components[0].contentGateway.productVersionName' "${rpa_files[0]}")
[[ $old_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || abort "could not read current productVersionName from ${rpa_files[0]}"
[[ $old_version != "$version" ]] || abort "RPAs are already at ${version}"
[[ ! -e $portal_file ]] || abort "$portal_file already exists"
echo "Updating ${old_version} -> ${version}"

for f in "${rpa_files[@]}"; do
    sed -i \
        -e "s|^\( *filename: \)agent-ove-${old_version//./\\.}\.${ARCH}\.iso$|\1agent-ove-${version}.${ARCH}.iso|" \
        -e "s|^\( *productVersionName: \)\"${old_version//./\\.}\"$|\1\"${version}\"|" \
        "$f"
    [[ $(git diff --numstat -- "$f" | awk '{print $1}') == 2 ]] \
        || abort "expected exactly 2 changed lines in $f (is it at ${old_version}?)"
done

cat > "$portal_file" <<EOF
---
versionName: "${version}"
ga: false
termsAndConditions: "RHD supportable user profile"
hidden: false
invisible: true
releaseDate: "${release_date}"
EOF
git add -N "$portal_file"

git --no-pager diff

if $dry_run; then
    abort
    echo "Dry run: no commit made."
    exit 0
fi

if $run_tox; then
    tox -e yamllint,test || abort "tox failed"
fi

title="Use ${version} for agent-based installer release"
what="Updating the version for the agent-based installer release and adding a new file to the portal for this version."
git add "${rpa_files[@]}" "$portal_file"
git commit -q -F - <<EOF
${title}

${what}
EOF
echo "Committed on branch ${new_branch}:"
git --no-pager log --oneline -1

description=$(cat <<EOF
#### What:
<!--- What is this change doing? --->

${what}

#### Why:
<!--- Please include the context and background for your change.  --->

#### Tickets:
<!--- Please link to any related Jira issue here. --->
EOF
)

urlencode() { python3 -c 'import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1], safe=""))' "$1"; }
project_url="https://${GITLAB_HOST}/${PROJECT_PATH}"
new_mr_url="${project_url}/-/merge_requests/new?merge_request%5Bsource_branch%5D=$(urlencode "$new_branch")&merge_request%5Btarget_branch%5D=${TARGET_BRANCH}&merge_request%5Btitle%5D=$(urlencode "$title")&merge_request%5Bdescription%5D=$(urlencode "$description")"

if ! $push; then
    printf '\nMR title:\n%s\n\nMR description:\n%s\n\n' "$title" "$description"
    echo "To create the MR, push the branch:"
    echo "  git push -u origin ${new_branch}"
    echo "then open:"
    echo "  $new_mr_url"
    exit 0
fi

git push -q -u origin "$new_branch"

if [[ -n ${GITLAB_TOKEN:-} ]]; then
    mr_url=$(curl -sS --fail-with-body -X POST \
        -H "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
        --data-urlencode "source_branch=${new_branch}" \
        --data-urlencode "target_branch=${TARGET_BRANCH}" \
        --data-urlencode "title=${title}" \
        --data-urlencode "description=${description}" \
        --data-urlencode "remove_source_branch=true" \
        --data-urlencode "squash=true" \
        "https://${GITLAB_HOST}/api/v4/projects/$(urlencode "$PROJECT_PATH")/merge_requests" \
        | yq -p json '.web_url') || die "MR creation failed; branch is pushed, open: $new_mr_url"
    echo "Created MR: $mr_url"
else
    echo "Branch pushed. GITLAB_TOKEN not set, so open this link to create the MR:"
    echo "  $new_mr_url"
fi
