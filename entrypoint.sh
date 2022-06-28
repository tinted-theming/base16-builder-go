#!/bin/bash

ARGS=""

if [[ -n $INPUT_PATH ]]; then
  ARGS+="-schemes-dir $INPUT_PATH"
else
  ARGS+="-online"
fi

/bin/base16-builder-go $ARGS
