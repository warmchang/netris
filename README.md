# netris
[![GoDoc](https://godoc.org/git.sr.ht/~tslocum/netris?status.svg)](https://godoc.org/git.sr.ht/~tslocum/netris)
[![builds.sr.ht status](https://builds.sr.ht/~tslocum/netris.svg)](https://builds.sr.ht/~tslocum/netris)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

Multiplayer Tetris clone

This project is not yet stable.  Feedback is welcome.

![](https://netris.rocketnine.space/static/screenshot2.png)

## Demo

To play netris without installing:

```ssh netris.rocketnine.space```

## Install

To install netris and netris-server:

```
GO111MODULE=on
go get -u git.sr.ht/~tslocum/netris
go get -u git.sr.ht/~tslocum/netris/cmd/netris
go get -u git.sr.ht/~tslocum/netris/cmd/netris-server
```

## Configure

See [CONFIGURATION.md](https://man.sr.ht/~tslocum/netris/CONFIGURATION.md)

## Play

A single player game may be played by launching without any options.

To play online, connect to the official server:

```netris --nick <name> --connect netris.rocketnine.space```

To host a private game, start a dedicated server:

```netris-server --listen-tcp :1984```

Then, connect with:

```netris --nick <name> --connect ip.or.dns.address:1984```

## Support

Please share suggestions/issues [here](https://todo.sr.ht/~tslocum/netris).

## Disclaimer

Tetris is a registered trademark of the Tetris Holding, LLC.

netris is no way affiliated with Tetris Holding, LLC.
