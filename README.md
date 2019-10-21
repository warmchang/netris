# netris
[![GoDoc](https://godoc.org/git.sr.ht/~tslocum/netris?status.svg)](https://godoc.org/git.sr.ht/~tslocum/netris)
[![builds.sr.ht status](https://builds.sr.ht/~tslocum/netris.svg)](https://builds.sr.ht/~tslocum/netris)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

Multiplayer Tetris clone

![](https://netris.rocketnine.space/static/screenshot4.png)

## Demo

To play netris without installing:

```ssh netris.rocketnine.space```

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
* [ssh](github.com/gliderlabs/ssh) - SSH server
* [pty](github.com/creack/pty) - Pseudo-terminal interface
* [go-isatty](github.com/mattn/go-isatty) - Terminal detection

## Disclaimer

Tetris is a registered trademark of the Tetris Holding, LLC.

netris is no way affiliated with Tetris Holding, LLC.
