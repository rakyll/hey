#!/bin/bash
# Since this script will be run in a rkt container, use "/bin/sh" instead of "/bin/bash"
# parse the current git commit hash
COMMIT=`git rev-parse HEAD`
# check if the current commit has a matching tag
TAG=$(git describe --exact-match --abbrev=0 --tags ${COMMIT} 2> /dev/null || true)

# use the matching tag as the version, if available
if [ -z "$TAG" ]; then
    VERSION=${COMMIT}
else
    VERSION=${TAG}
fi

# check for changed files (not untracked files)
if [ -n "$(git diff --shortstat 2> /dev/null | tail -n1)" ]; then
    VERSION="${VERSION}-dirty"
fi

echo ${VERSION}
