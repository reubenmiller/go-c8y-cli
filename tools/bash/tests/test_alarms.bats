#!/usr/bin/env bats

setup () {
  echo "setting up test" >&2
}

teardown () {
  echo "Tearing down test" >&2
}

@test "check alarms list" {
  result="$(c8y alarms list --raw)"
  [ "$?" -eq 0 ]
}
