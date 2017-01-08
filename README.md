# Go Plugin Example

This is a simple example that shows how to use the newly added `plugin` package in Go 1.8 (see https://tip.golang.org/pkg/plugin/).  A Go plugin is package compiled with the `-buildmode=plugin` which creaetes a shared library (`.so`) file.  Go can then dynamically load the shared library at runtime and access exported functions an variables.

## Requirements
- Go 1.8 

## Restrictions
Go 1.8 (beta 2) only supports plugin on Linux OS.

This may change by the time GA comes out, double check.

## Pluggable Greeting System
The plugin example in this repository uses the Go plugin system to implement a simple greeting system.  Each plugin package (`./eng`, `./chi`) implements a greeting meesage in a different lanaguage.
File `greeter.go` uses the Go `plugin` package to load the plugin module and displays the proper message based on passed command parameters.

### The Plugin
