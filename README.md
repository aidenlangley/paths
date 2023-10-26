[![builds.sr.ht status](https://builds.sr.ht/~nedia/paths/commits/main/.build.yml.svg)](https://builds.sr.ht/~nedia/paths/commits/main/.build.yml?)
[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~nedia/paths.svg)](https://pkg.go.dev/git.sr.ht/~nedia/paths)

# paths

A go module for working with paths.

## Features

 - Provides `Home`, `Config` and `Cache` for retrieving `$HOME`, `$HOME/.config`
   and `$HOME/.cache` respectively.
 - Resolves paths via `paths.Resolve` & `paths.PathResolver.Resolve`.
 - Provides `FileInfo` which stores the size in bytes, modified `time.Time` and
   other relevant information.
