[![GitHub release](https://img.shields.io/github/v/release/bodgit/psx)](https://github.com/bodgit/psx/releases)
[![Build Status](https://img.shields.io/github/workflow/status/bodgit/psx/build)](https://github.com/bodgit/psx/actions?query=workflow%3Abuild)
[![Coverage Status](https://coveralls.io/repos/github/bodgit/psx/badge.svg?branch=main)](https://coveralls.io/github/bodgit/psx?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/bodgit/psx)](https://goreportcard.com/report/github.com/bodgit/psx)
[![GoDoc](https://godoc.org/github.com/bodgit/psx?status.svg)](https://godoc.org/github.com/bodgit/psx)
![Go version](https://img.shields.io/badge/Go-1.16-brightgreen.svg)
![Go version](https://img.shields.io/badge/Go-1.15-brightgreen.svg)
![Go version](https://img.shields.io/badge/Go-1.14-brightgreen.svg)

psx
===

A collection of libraries and utilities for dealing with Sony PlayStation 1 file formats.

Full installation:
```
go get github.com/bodgit/psx/...
```
Or grab a pre-built binary from the [releases page](https://github.com/bodgit/psx/releases).

## psx

The `psx` utility currently allows you to split generic memory cards (such as those created by the [8BitMods MemCard PRO](https://8bitmods.com/memcard-pro-for-playstation-1-smoke-black/)) into per-game memory cards for use with a supported ODE.

A quick demo:

<img src="./psx.svg">
