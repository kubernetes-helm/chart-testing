#!/usr/bin/env bash

# Copyright The Helm Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

readonly SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

show_help() {
cat << EOF
Usage: $(basename "$0") <options>

Create and push a tag.

    -h, --help        Display help
    -d, --debug       Display verbose output
    -r, --remote      The name of the remote to push the tag to (default: upstream)
    -f, --force       Force an existing tag to be overwritten
    -t, --tag         The name of the tag to create
    -s, --skip-push   Skip pushing the tag
EOF
}

main() {
    local debug=
    local tag=
    local remote=upstream
    local force=()
    local skip_push=

    while :; do
        case "${1:-}" in
            -h|--help)
                show_help
                exit
                ;;
            -d|--debug)
                debug=true
                ;;
            -t|--tag)
                if [[ -n "${2:-}" ]]; then
                    tag="$2"
                    shift
                else
                    echo "ERROR: '--tag' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -r|--remote)
                if [[ -n "${2:-}" ]]; then
                    remote="$2"
                    shift
                else
                    echo "ERROR: '--remote' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -f|--force)
                force+=(--force)
                ;;
            -s|--skip-push)
                skip_push=true
                ;;
            *)
                break
                ;;
        esac

        shift
    done

    if [[ -z "$tag" ]]; then
        echo "ERROR: --tag is required!" >&2
        show_help
        exit 1
    fi

    if [[ -n "$debug" ]]; then
        set -x
    fi

    pushd "$SCRIPT_DIR" > /dev/null

    git fetch "$remote"
    git tag -a -m "Release $tag" "$tag" "$remote/dev-v2" "${force[@]}"

    if [[ -z "$skip_push" ]]; then
        git push "$remote" "refs/tags/$tag" "${force[@]}"
    fi

    popd > /dev/null
}

main "$@"
