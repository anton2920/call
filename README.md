# Call

Call is an application for NxN peer-to-peer audio/video conferencing.

You can share your voice, video from camera or your computer's screen/tab (as well as tab's audio in Chrome). Additionally, you can enable/disable «audio enhancements» such as automatic gain control, echo cancellation or noise reduction. Generally the default (only automatic gain is enabled) is pretty good, but it's up to you to experiment and find the best setup, because unlike Google Meet, you are in control, not software manufacturer.

# Installation

You don't really need to install it, as you can just use it right away by going to https://call.anton2920.ru. A room will be created for you (TBD). Just share a link with your friends and start conferencing.

If you want to build and deploy it on your own server, you need to download [golang.org/x/net/websocket]() for accepting websocket connections from browser. For a convenience, it's backported into this project via git modules. Run `git submodule init`, `git submodule update` and then `./make.sh release` to build it from scratch.

As an alternative, put websocket library into your `$GOPATH` and use `go build` to build everything.

# Supported browsers

Your browser needs to support JavaScript, WebSockets and WebRTC. It works in Firefox 52 (Windows XP), as well as in any newer browser. Works in mobile browsers too.

# Copyright

Pavlovskii Anton, 2024 (MIT). See [LICENSE](LICENSE) for more details.
