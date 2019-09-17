# stick
[![GoDoc](https://godoc.org/git.sr.ht/~tslocum/stick?status.svg)](https://godoc.org/git.sr.ht/~tslocum/stick)
[![builds.sr.ht status](https://builds.sr.ht/~tslocum/stick.svg)](https://builds.sr.ht/~tslocum/stick?)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

Shareable Git-backed Markdown-formatted notes

## Features

- Notebooks are standard git repositories containing .md files
- Notebooks may be used privately, shared with one or more users privately, or shared publicly
- Notes may be customized (e.g., to-do style lists with automatic sorting and de-duplication)

## Demo

[**Try stick**](https://stick.rocketnine.space/#login/c89f5381659ad34bd84967fdbbb5e76834495063dc93ed3902bf4e186f4def90)

**Note:** Read-only except checking/un-checking items.

## Install

Choose one of the following methods:

### Download

[**Download stick**](https://stick.rocketnine.space/download/?sort=name&order=desc)

Windows and Linux binaries are available.

### Compile

```
GO111MODULE=on go get -u git.sr.ht/~tslocum/stick
```

## Configure

See [CONFIGURATION.md](https://man.sr.ht/~tslocum/stick/CONFIGURATION.md)

## Run

```
stick serve
```

## Customize

See [OPTIONS.md](https://man.sr.ht/~tslocum/stick/OPTIONS.md)

## Support

Please share suggestions/issues [here](https://todo.sr.ht/~tslocum/stick).
