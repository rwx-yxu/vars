# vars

[![Go Reference](https://pkg.go.dev/badge/github.com/rwx-yxu/vars.svg)](https://pkg.go.dev/github.com/rwx-yxu/vars)
[![Go Report Card](https://goreportcard.com/badge/github.com/rwx-yxu/vars)](https://goreportcard.com/report/github.com/rwx-yxu/vars)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**vars** is a simple, persistent, stateful configuration manager for Go CLI applications.

It provides a key-value store that lives in your user's `XDG_STATE_HOME` (typically `~/.local/state`), to store cli state values.

> **Attribution:** This project is heavily inspired by the design of the vars implementation in [rwxrob/bonzai](https://github.com/rwxrob/bonzai). It adapts that treating state as a flat, file-based key-value store—and re-implements it specifically for the [Cobra](https://github.com/spf13/cobra) ecosystem.

---

## Features

* **Zero Config Persistence:** Automatically manages directory creation and file locking.
* **XDG Compliant:** Respects `XDG_STATE_HOME`, keeping user home directories clean.
* **Editor Support:** Built-in `edit` command opens configuration in `VISUAL` or `EDITOR` (vim, nano, code, etc.).
* **Dual Mode:** Works as a **standalone CLI** tool or an **embedded library** for your own apps.
* **Strict Scoping:** Enforces a clean `namespace/scope` structure to prevent deeply nested complexity.

## Installation

### As a Library (for your Go project)

```
go get -u github.com/rwx-yxu/vars@latest
```

Usage: Library
Vars provides public functions to interact with vars files. Simply import the vars package to get started.

```go
import "github.com/rwx-yxu/vars"

func run() {
    // Initialize the handler
    v := vars.New("my-app")
    
    // Ensure the vars exists
    _ = v.Init()

    // Get a value
    apiKey, err := v.Get("api_key")
    if err != nil {
        // Handle error (key not found)
    }
}
```

### Cobra Subcommand
Use vars.NewCmd to attach a fully featured vars command to your application tree.

```go
package main

import (
    "github.com/rwx-yxu/vars"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "my-app"}

    // Add a 'vars' subcommand to your root command.
    // This will store data in ~/.local/state/my-app/vars.properties
    rootCmd.AddCommand(vars.NewCmd("my-app"))

    // You can also create scoped buckets for specific features
    // e.g. ~/.local/state/my-app/auth/vars.properties
    // authCmd.AddCommand(vars.NewCmd("my-app", "auth"))

    rootCmd.Execute()
}
```

# Usage: Standalone CLI
If you installed the binary, you can use vars as a general-purpose state manager for your shell scripts or local environment.

## Initialization
Before using a namespace, you must initialize it.

```bash

# Initialize storage for a namespace "my-scripts"
vars init my-app
```
## Managing Variables

```bash
# Set a value
vars set my-app api_token "123456"

# Get a value (useful in scripts: token=$(vars get my-scripts api_token))
vars get my-app api_token

# List all variables
vars data my-app

# Open the file in your default $EDITOR
vars edit my-app
```

## Scoped Variables
You can add an optional second argument to create a "scope" (a subdirectory).

```bash
vars init my-app ingest
vars set my-app ingest weather "api_key"
```

# Storage Structure
Data is stored in strict adherence to standard Linux/Unix conventions.

Root: $XDG_STATE_HOME (defaults to ~/.local/state if unset).

File: vars.properties (Java properties style key=value).

Example Tree:

```
~/.local/state/
└── my-app/
    ├── vars.properties         # vars.New("my-app")
    └── ingest/
        └── vars.properties     # vars.New("my-app", "ingest")
```

License
Copyright 2025. Licensed under the Apache License 2.0. See LICENSE for details.
