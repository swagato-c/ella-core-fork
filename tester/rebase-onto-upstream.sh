#!/usr/bin/env bash
# SPDX-FileCopyrightText: Ella Networks Inc.
# SPDX-License-Identifier: BUSL-1.1
#
# Rebase the flexicore fork's commits onto a newer Ella Core commit.
#
#   tester/rebase-onto-upstream.sh <upstream-commit>   # e.g. a stage boundary's pin
#   tester/rebase-onto-upstream.sh                     # upstream/main's tip
#
# The fork is the branch `flexicore`: Ella Core at a pinned commit plus our
# commits on top, bottom one first ("Move internal/tester to tester ...").
# The old base is wherever the branch leaves upstream's history, so no
# state file is needed. Run it only at stage boundaries (ADR 0020 in
# swagato-c/flexicore), then run the tester's tests and Ella's scenarios,
# push with --force-with-lease, and bump the pin in flexicore's
# acceptance/go.mod.
set -euo pipefail

cd "$(dirname "$0")/.."

if ! git remote get-url upstream >/dev/null 2>&1; then
  git remote add upstream https://github.com/ellanetworks/core.git
fi
git fetch --quiet upstream

target="${1:-upstream/main}"
target_sha="$(git rev-parse --verify "${target}^{commit}")"

if ! git merge-base --is-ancestor "$target_sha" upstream/main; then
  echo "error: $target is not on upstream/main" >&2
  exit 1
fi

branch="$(git rev-parse --abbrev-ref HEAD)"
if [ "$branch" != "flexicore" ]; then
  echo "error: check out the flexicore branch first (on $branch)" >&2
  exit 1
fi

old_base="$(git merge-base HEAD upstream/main)"
if [ "$old_base" = "$target_sha" ]; then
  echo "already based on ${target_sha:0:8}"
  exit 0
fi

echo "rebasing $(git rev-list --count "$old_base"..HEAD) fork commits from ${old_base:0:8} onto ${target_sha:0:8}"
git rebase --onto "$target_sha" "$old_base" flexicore

echo
echo "done: flexicore is now based on ${target_sha:0:8}. Next:"
echo "  go build ./tester/... ./cmd/core-tester && go test ./tester/..."
echo "  run Ella's scenarios (see tester/README.md), then"
echo "  git push --force-with-lease origin flexicore"
