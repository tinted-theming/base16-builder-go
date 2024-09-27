# base16-builder-go

A simple builder for base16 templates and schemes.

This currently implements version 0.11.0 of the [base16 spec](https://github.com/tinted-theming/home).

## Building

Currently version 1.16 or higher of the Go compiler is needed.

Unfortunately, because the schemes are stored in a separate repo, the schemes
submodule needs to be cloned before building.

The following command will clone the schemes directory

```
$ git submodule update --init
```

Now that the repo is cloned, you can use `go build` to create a binary. You may
wish to update the schemes dir to get new included schemes.

## Install

### Arch Linux

```
yay base16-builder-go
```

Available as command after installation: `base16-builder`

## Commands

By default, this builder will build the template in the the current directory
using the compiled-in schemes. If you want to update schemes independently, you
can use the -schemes-dir flag to point to another directory.

```
Usage of base16-builder-go:
  -schemes-dir string
    	Target directory for scheme data. The default value uses internal schemes. (default "-")
  -template-dir string
    	Target template directory to build. (default ".")
  -verbose
    	Log all debug messages
```
