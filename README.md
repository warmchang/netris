# netris
[![GoDoc](https://godoc.org/git.sr.ht/~tslocum/netris?status.svg)](https://godoc.org/git.sr.ht/~tslocum/netris)
[![builds.sr.ht status](https://builds.sr.ht/~tslocum/netris.svg)](https://builds.sr.ht/~tslocum/netris)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

Multiplayer Tetris clone

## Play Without Installing

To play netris without installing, connect via [SSH](https://en.wikipedia.org/wiki/Secure_Shell):

```ssh netris.rocketnine.space```

## Screenshot

[![](https://netris.rocketnine.space/static/screenshot5.png)](https://netris.rocketnine.space/static/screenshot5.png)

## Install

Choose one of the following methods:

### Download

[**Download netris**](https://netris.rocketnine.space/download/?sort=name&order=desc)

Windows and Linux binaries are available.

### Compile

```
GO111MODULE=on
go get -u git.sr.ht/~tslocum/netris
go get -u git.sr.ht/~tslocum/netris/cmd/netris
go get -u git.sr.ht/~tslocum/netris/cmd/netris-server
```

## Configure

See [CONFIGURATION.md](https://man.sr.ht/~tslocum/netris/CONFIGURATION.md)

## How to Play

See [GAMEPLAY.md](https://man.sr.ht/~tslocum/netris/GAMEPLAY.md)

## Support

Please share suggestions/issues [here](https://todo.sr.ht/~tslocum/netris).

## Libraries

The following libraries are used to build netris:

* [tcell](https://github.com/gdamore/tcell) - User interface
* [tview](https://github.com/rivo/tview) - User interface
* [ssh](https://github.com/gliderlabs/ssh) - SSH server
* [pty](https://github.com/creack/pty) - Pseudo-terminal interface
* [go-isatty](https://github.com/mattn/go-isatty) - Terminal detection

## Disclaimer

Tetris is a registered trademark of the Tetris Holding, LLC.

netris is in no way affiliated with Tetris Holding, LLC.
