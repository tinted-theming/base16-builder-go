# base16-builder-go

A simple builder for base16 templates and schemes.

This currently implements version 0.10.0 of the
[base16 spec](https://github.com/base16-project/base16).

## Building

Currently version 1.16 or higher of the Go compiler is needed.

Unfortunately, because the schemes are stored in a separate repo, the schemes
repo needs to be cloned before building.

The following command will clone the schemes directory

```
$ git clone https://github.com/base16-project/base16-schemes.git schemes
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

### Base24 Support

This builder has experimental support for [base24](https://github.com/Base24)
schemes and templates.
