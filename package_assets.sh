#!/bin/sh
set -e
cd brogger/

# Package assets into go-file
bin2go                  \
-s="base_assets.go"     \
-p="${PWD##*/}"         \
-a                      \
-g                      \
base/assets/css/*.css   \
base/assets/js/*.js     \
base/templates/*.gohtml \
base/posts/*.md
