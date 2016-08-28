# base16-builder-go

A simple builder for base16 templates and schemes, modeled off of
[base16-builder-php](https://github.com/chriskempson/base16-builder-php).

This currently implements version 0.8 of the [base16
spec](https://github.com/chriskempson/base16).

## Commands

There are two main commands: update and build.

`update` will pull in any template and scheme updates (or clone the repos if
they don't exist).

`build` will build all templates following the spec for all templates and
schemes defined.
