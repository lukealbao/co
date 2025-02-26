# co: code owners command-line tool

[![build](https://github.com/lukealbao/co/workflows/build/badge.svg)](https://github.com/lukealbao/co/actions/workflows/ci.yml)
[![build](https://github.com/lukealbao/co/workflows/release/badge.svg)](https://github.com/lukealbao/co/actions/workflows/release.yml)

A CLI for GitHub's [CODEOWNERS file](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/about-code-owners#codeowners-syntax).

This repository is forked from [https://github.com/hmarr/codeowners](https://github.com/hmarr/codeowners). This is the most correct implementation
of Github's rules engine I've found. Other solutions (e.g., Gitlab) may have differences which may not be featured here.

## Installation

### Local Usage

You can download a binary from the [releases page](https://github.com/lukealbao/co/releases) and put the `co` binary somewhere in your path.

### Usage

```
co help

Usage:
  co [command]

Available Commands:
  diff        Print a unified diff of file ownership
  fmt         Normalize CODEOWNERS format
  help        Help about any command
  lint        Validate codeowners file
  stats       Display code ownership statistics
  version     Print code version
  who         List code owners for file(s)
  why         Identify which rule effects ownership for a single file.

Flags:
  -f, --file string   CODEOWNERS file path
  -h, --help          help for co

Use "co [command] --help" for more information about a command.
```
