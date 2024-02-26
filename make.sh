#!/bin/sh

PROJECT=call

VERBOSITY=0
VERBOSITYFLAGS=""
while test "$1" = "-v"; do
	VERBOSITY=$((VERBOSITY+1))
	VERBOSITYFLAGS="$VERBOSITYFLAGS -v"
	shift
done

run()
{
	if test $VERBOSITY -gt 1; then echo "$@"; fi
	"$@" || exit 1
}

# NOTE(anton2920): don't like Google spying on me.
GOPROXY=direct; export GOPROXY
GOSUMDB=off; export GOSUMDB

# NOTE(anton2920): disable Go 1.11+ package management.
GO111MODULE=off; export GO111MODULE
GOPATH=`go env GOPATH`:`pwd`; export GOPATH

CGO_ENABLED=0; export CGO_ENABLED

STARTTIME=`date +%s`

case $1 in
	'' | debug)
		CGO_ENABLED=1; export CGO_ENABLED
		run go build $VERBOSITYFLAGS -o $PROJECT -race -gcflags='all=-N -l' -ldflags='-X main.DebugMode=on' .
		;;
	clean)
		run rm -f $PROJECT $PROJECT.s $PROJECT.esc $PROJECT.test c.out cpu.pprof mem.pprof
		run go clean -cache -modcache -testcache
		run rm -rf `go env GOCACHE`
		run rm -rf /tmp/cover*
		;;
	disas | disasm | disassembly)
		printv go build $VERBOSITYFLAGS -pgo off -gcflags="-S"
		go build $VERBOSITYFLAGS -gcflags="-S" >$PROJECT.s 2>&1
		;;
	esc | escape | escape-analysis)
		printv go build $VERBOSITYFLAGS -pgo off -gcflags="-m -m"
		go build $VERBOSITYFLAGS -gcflags="-m -m" >$PROJECT.m 2>&1
		;;
	fmt)
		if which goimports >/dev/null; then
			run goimports -l -w *.go
		else
			run gofmt -l -s -w *.go
		fi
		;;
	objdump)
		go build $VERBOSITYFLAGS -o $PROJECT -pgo off
		printvv go tool objdump -S -s ^main\. $PROJECT
		go tool objdump -S -s ^main\. $PROJECT >$PROJECT.s
		;;
	release)
		run go build $VERBOSITYFLAGS -o $PROJECT -ldflags="-s -w"
		;;
	vet)
		run go vet $VERBOSITYFLAGS
		;;
	*)
		echo "Target is not supported!"
		;;
esac

ENDTIME=`date +%s`

echo Done $1 in $((ENDTIME-STARTTIME))s
