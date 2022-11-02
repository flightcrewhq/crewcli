# Copyright 2022 Bindl Authors
# Modifications Copyright 2022 Flightcrew
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Original file found at https://github.com/bindl-dev/bindl/blob/main/bootstrap.sh
# Modified to handle crewcli, and updated how the coloring is formatted.

#!/usr/bin/env bash

set -e
set -o pipefail

OS="$(uname -s)"
ARCH="$(uname -m)"

PROGRAM_NAME="crewcli"
REPOSITORY="flightcrewhq/crewcli"

WORKDIR="$(mktemp -d /tmp/bootstrap-${PROGRAM_NAME}-XXXXX)"
ARCHIVE="${PROGRAM_NAME}-${OS}-${ARCH}.tar.gz"

OUTDIR="${OUTDIR:-$(pwd)}"
TAG="${TAG:-latest}"

CODE_GREEN="32"
CODE_RED="31"
COLOR_RESET="\e[0m"
CODE_BOLD="1"
COLOR_BOLD="\e[${CODE_BOLD}m"
COLOR_BOLDGREEN="\e[${CODE_BOLD};${CODE_GREEN}m"
COLOR_BOLDRED="\e[${CODE_BOLD};${CODE_RED}m"

OUTDIR="${OUTDIR:-$(pwd)}"
TAG="${TAG:-latest}"

function log() {
  printf "${COLOR_BOLD}[bootstrap]${COLOR_RESET} $1\n"
}

function prompt() {
  if [ -t 0 ]; then
    read -p "❗️ Proceed? (y/N) " answer </dev/tty
    if [ $answer != "y" ]; then
      echo "Aborted: only 'y' is accepted answer to continue (received '${answer}')"
      exit 1
    fi
  else
    log "Detected non-interactive mode, prompt implictly proceeds"
    return
  fi
}

function copy_local() {
  set +e
  exe=$(which ${PROGRAM_NAME} 2>/dev/null)
  set -e

  if [ -z $exe ]; then
    return
  fi

  dst=${OUTDIR}/${PROGRAM_NAME}

  if [ $exe == $dst ]; then
    log "Found ${PROGRAM_NAME} in ${OUTDIR}, my job here is done."
    exit 0
  fi

  log "I found ${exe}, I will now create symbolic link to ${OUTDIR}"
  ln -s ${exe} ${dst} || return
  log "Done!"
  exit 0
}

log "Hello! The sole purpose of my existence is to bootstrap ${PROGRAM_NAME}."

copy_local

log "I have found myself in ${ARCH} machine running ${OS}."
log "I expect the archive to be named ${ARCHIVE}."

prompt

log "Working in ${WORKDIR}"
pushd "${WORKDIR}" >/dev/null
  log "Downloading (1/2): checksums.txt"
  curl --silent --location --remote-name "https://github.com/${REPOSITORY}/releases/${TAG}/download/checksums.txt"

  log "Downloading (2/2): ${ARCHIVE}"
  curl --silent --location --remote-name "https://github.com/${REPOSITORY}/releases/${TAG}/download/${ARCHIVE}"

  downloaded=$(ls -A | tr '\n' ' ')
  log "Downloaded: ${downloaded}"

  log "Verifying checksums"
  shasum --algorithm 256 --check checksums.txt --ignore-missing

  tar -xzf ${ARCHIVE} ${PROGRAM_NAME}

  log "Printing program version"
  ./${PROGRAM_NAME} version
popd >/dev/null

trap "rm -r ${WORKDIR}" EXIT

mv ${WORKDIR}/${PROGRAM_NAME} ${OUTDIR}/.
log "Done! The binary is in ${OUTDIR}"
