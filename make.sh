#!/bin/bash
LDFLAGS_STATIC='-linkmode external'
EXTLDFLAGS_STATIC='-static'
BUILDFLAGS=( -a -tags "netgo static_build" )
EXTLDFLAGS_STATIC="$EXTLDFLAGS_STATIC -lpthread -Wl,--unresolved-symbols=ignore-in-object-files"
LDFLAGS_STATIC="
  $LDFLAGS_STATIC
  -extldflags \"$EXTLDFLAGS_STATIC\"
"
go get

go build  "${BUILDFLAGS[@]}" \
    -ldflags "
        $LDFLAGS
        $LDFLAGS_STATIC
    "
