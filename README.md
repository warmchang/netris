# netris
[![CI status](https://gitlab.com/tslocum/netris/badges/master/pipeline.svg)](https://gitlab.com/tslocum/netris/commits/master)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

Multiplayer Tetris clone

## Play Without Installing

To play netris without installing, connect via [SSH](https://en.wikipedia.org/wiki/Secure_Shell):

```bash
ssh netris.rocketnine.space
```

## Screenshot

[![](https://netris.rocketnine.space/static/screenshot5.png)](https://netris.rocketnine.space/static/screenshot5.png)

## Install

Choose one of the following methods:

### Download

[**Download netris**](https://netris.rocketnine.space/download/?sort=name&order=desc)

Windows and Linux binaries are available.

### Compile

```bash
go get gitlab.com/tslocum/netris/cmd/netris
```

## Configure

See [CONFIGURATION.md](https://gitlab.com/tslocum/netris/blob/master/CONFIGURATION.md)

## How to Play

See [GAMEPLAY.md](https://gitlab.com/tslocum/netris/blob/master/GAMEPLAY.md)

## Support

Please share issues and suggestions [here](https://gitlab.com/tslocum/netris/issues).

## Libraries

The following libraries are used to build netris:

* [tslocum/cview](https://gitlab.com/tslocum/cview) - User interface
* [gdamore/tcell](https://github.com/gdamore/tcell) - User interface
* [gliderlabs/ssh](https://github.com/gliderlabs/ssh) - SSH server
* [creack/pty](https://github.com/creack/pty) - Pseudo-terminal interface
* [mattn/go-isatty](https://github.com/mattn/go-isatty) - Terminal detection

## Disclaimer

Tetris is a registered trademark of the Tetris Holding, LLC.

netris is in no way affiliated with Tetris Holding, LLC.

netris is in no way affiliated with Netris by Mark H. Weaver.
