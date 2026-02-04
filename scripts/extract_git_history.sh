#!/usr/bin/env bash
#
# Extract every revision of a file from a git repository into timestamped
# files under TMPFOLDER. Used to build historical pricing data from
# google-cloud-pricing-cost-calculator (e.g. pricing.yml).
#
# Usage:
#   extract_git_history.sh <path-in-repo> [repo-dir] [repo-url]
#
# If repo-dir is given: clone the repo if it does not exist, otherwise pull.
# Then run extraction from repo-dir. repo-url defaults to the Cyclenerd
# pricing calculator repo if not provided.
#
# If repo-dir is omitted: run from current directory (must be inside the repo).
#
# Examples:
#   ../scripts/extract_git_history.sh pricing.yml
#
# Output: /tmp/pricing-data/YYYY-MM-DD.HHMM.SS.<rev>

set -euo pipefail

DEFAULT_REPO_URL="https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator"

if [[ $# -lt 1 ]]; then
	echo "Usage: $(basename "$0") <path-in-repo> [repo-dir] [repo-url]" >&2
	echo "" >&2
	echo "  path-in-repo  File path inside the repo (e.g. pricing.yml)" >&2
	echo "  repo-dir      Optional. If set: clone repo here if missing, else git pull" >&2
	echo "  repo-url      Optional. Used only when cloning (default: Cyclenerd repo)" >&2
	exit 1
fi

FILE="$1"
REPO_DIR="${2:-}"
REPO_URL="${3:-$DEFAULT_REPO_URL}"

TMPFOLDER="/tmp/pricing-data"
mkdir -p "$TMPFOLDER"

if [[ -n "$REPO_DIR" ]]; then
	if [[ -d "$REPO_DIR/.git" ]]; then
		echo "Repository $REPO_DIR exists, pulling latest changes..."
		(cd "$REPO_DIR" && git pull)
	else
		echo "Cloning $REPO_URL into $REPO_DIR..."
		git clone "$REPO_URL" "$REPO_DIR"
	fi
	cd "$REPO_DIR"
fi

for rev in $(git log --pretty=format:"%h" -- "$FILE"); do
	datestamp="$(git show --no-patch --no-notes --date=format:'%Y-%m-%d.%H%M.%S' --pretty=format:"%ad" "$rev")"
	filestamp="${datestamp}.${rev}"
	git show "${rev}:${FILE}" > "${TMPFOLDER}/${filestamp}"
done
