# base16-builder-go

A simple builder for base16 templates and schemes.

This currently implements version 0.10.0 of the
[base16 spec](https://github.com/chriskempson/base16).

## Building

Currently version 1.16 or higher of the Go compiler is needed.

Currently 0.10.0 is not an official release of the base16 spec and some things
are still being ironed out, so an additional step is needed.

The following command will clone the schemes directory

```
$ git clone https://github.com/belak/base16-schemes.git schemes
```

Now that the repo is cloned, you can use `go build` to create a binary. You may
wish to update the schemes dir to get new included schemes. In the future this
will most likely be provided as a submodule, updated on a regular basis.

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

## Notes

I'm open to making a few template-specific tweaks as long as they'll be useful
to other templates. Below is a listing of the additions to the base16 spec which
this builder supports.

### Additional variables

* `scheme-slug-underscored` - A version of the scheme slug where dashes have
  been replaced with underscores.
