#!/usr/bin/env bats

load '/usr/local/lib/bats/load.bash'

load '../lib/download.bash'

#
# Tests for top-level docker bootstrap command. The rest of the plugin runs in Go.
#

# Uncomment the following line to debug stub failures
# export [stub_command]_STUB_DEBUG=/dev/tty
#export DOCKER_STUB_DEBUG=/dev/tty

#TODO: Update this to reflect what we need to test in the task runner code
setup() {
  export BUILDKITE_PLUGIN_ECR_TASK_RUNNER_BUILDKITE_PLUGIN_MESSAGE=true
}

#TODO: Update this to reflect what we need to test in the task runner code
teardown() {
    unset BUILDKITE_PLUGIN_ECR_TASK_RUNNER_BUILDKITE_PLUGIN_MESSAGE
    rm ./migrations-runner-buildkite-plugin || true
}

create_script() {
cat > "$1" << EOM
set -euo pipefail

echo "executing $1:\$@"

EOM
}

@test "Downloads and runs the command for the current architecture" {

  function downloader() {
    echo "$@";
    create_script $2
  }
  export -f downloader

  run download_binary_and_run

  unset downloader

  assert_success
  assert_line --regexp "https://github.com/cultureamp/migrations-runner-buildkite-plugin/releases/latest/download/migrations-runner-buildkite-plugin_linux_amd64 migrations-runner-buildkite-plugin"
  assert_line --regexp "executing migrations-runner-buildkite-plugin"
}
