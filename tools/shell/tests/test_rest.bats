#!/usr/bin/env bats

setup () {
  echo "Setting up test" >&2
}

teardown () {
  echo "Tearing down test" >&2
}

@test "Upload file using custom rest command" {
  tmpfile=$( mktemp )
  echo "example contents" > "$tmpfile"
  result=$( c8y rest POST /inventory/binaries --file $tmpfile )
  [ "$?" -eq 0 ]

  rm -f $tmpfile

  # download file
  id=$( echo $result | jq -r ".id" )
  result2=$( c8y rest /inventory/binaries/$id )
  [ "$?" -eq 0 ]
  [ "$result2" == "example contents" ]
}
