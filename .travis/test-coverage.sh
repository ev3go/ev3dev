#!/bin/bash

[ -n "$COVERALLS_TOKEN" ] && $HOME/gopath/bin/goveralls -repotoken $COVERALLS_TOKEN