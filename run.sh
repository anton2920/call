#!/bin/sh

`pwd`/call $@ >/var/log/call.log 2>&1 &
