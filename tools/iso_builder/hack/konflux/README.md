# konflux

Helper scripts for releasing a new OCP version of the OVE UI agent ISO
(`agent-ove.x86_64.iso`) through Konflux. Each new OCP z-stream takes two
changes, made in this order:

1. **`bump-ove-ui-release.sh`** opens a PR against
   [ocp-build-data](https://github.com/openshift-eng/ocp-build-data) that
   updates `RELEASE_VALUE` and `PATCH_VERSION` on the `installer-ove-ui-<X.Y>`
   branch. Merging it starts the ISO build pipeline.
2. **`konflux-release-mr.sh`** runs after the build is done. It opens an MR
   against
   [konflux-release-data](https://gitlab.cee.redhat.com/releng/konflux-release-data)
   that updates the ISO filename and `productVersionName` in the prod/stage
   ReleasePlanAdmissions and adds a developer portal file for the new version.

Each script takes the full OCP version, e.g. `4.22.21` or `5.0.3`. From that
version, the scripts derive the `installer-ove-ui-X.Y` branch and the
`...-X-Y-{prod,stage}.yaml` RPAs. Run either script with `-h` to see all
options.

## Requirements

- `git`, `sed`, `python3`
- `bump-ove-ui-release.sh`: [`gh`](https://cli.github.com/), authenticated
- `konflux-release-mr.sh`: [`yq`](https://github.com/mikefarah/yq) v4 (mikefarah,
  not the pypi package), and optionally `curl` plus `GITLAB_TOKEN`

## bump-ove-ui-release.sh

Run it from the top of a clone of **your fork** of ocp-build-data, where
`origin` is your fork. The script always takes the base branch from upstream,
so your fork doesn't need to be synced.

```bash
cd ocp-build-data                       # your fork
/path/to/bump-ove-ui-release.sh -n 4.22.21   # dry run: show the diff
/path/to/bump-ove-ui-release.sh 4.22.21      # branch + commit, print PR title/description
/path/to/bump-ove-ui-release.sh -r 4.22.21   # also push to your fork and open the PR
```

The script creates the branch `bump-ove-ui-release-<version>`. The PR title
and description follow
[#13690](https://github.com/openshift-eng/ocp-build-data/pull/13690), and the
description links the previous bump PR. After the PR merges, trigger the
layered-products build as described in the PR's test plan.

## konflux-release-mr.sh

Run it from the top of a clone of **upstream** `releng/konflux-release-data`.
The project must not be forked: branches are pushed directly to it, so you
need push access.

```bash
cd konflux-release-data
/path/to/konflux-release-mr.sh -n 4.22.21    # dry run: show the diff
/path/to/konflux-release-mr.sh 4.22.21       # branch + commit, print MR title/description
/path/to/konflux-release-mr.sh -p 4.22.21    # also push and create the MR
```

The script creates the branch `abi-<version>` from `origin/main`, following
[!23685](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/merge_requests/23685).
Other options:

- `-d YYYY-MM-DD` sets the portal `releaseDate`. The default is today.
- `-t` runs `tox -e yamllint,test` before committing.

How `-p` creates the MR depends on `GITLAB_TOKEN`:

- If `GITLAB_TOKEN` is set to a personal access token with `api` scope, the
  script creates the MR through the GitLab API.
- Otherwise, the script pushes the branch and prints a link to a pre-filled
  "New merge request" form.

The script doesn't hide or remove older versions in the developer portal.

## Safety checks

Both scripts refuse to run in these cases:

- the working tree has uncommitted changes
- the branch already exists
- the release is already at the requested version
- the edit doesn't change exactly the expected lines

If a script fails before committing, it switches back to your original branch
and deletes the new one.

## New minor/major versions

For a new `X.Y`, such as 5.0, someone must first create the
`installer-ove-ui-X.Y` branch in ocp-build-data and add the
`art-agent-installer-iso-X-Y-{prod,stage}.yaml` RPAs in konflux-release-data.
After that, the scripts work without changes. If either piece is missing, the
scripts tell you which one.
