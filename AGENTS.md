# AGENTS.md

## Overview

This repository contains a small Go CLI named `run`.

Its purpose is to make repeated execution of slightly different build outputs easier, especially for cases like:

- `Debug` vs `Release`
- `CUI` vs `GUI`
- test-only variants such as DB-focused binaries
- repeated runs where only a few `name=value` arguments change

The core design goal is:

- keep everyday usage CLI-first
- make previous invocations easy to rerender and edit
- keep the internal model generic across projects

## Command Model

Primary command shape:

```text
run <profile> [name=value ...] [-- extra args...]
```

Related commands:

```text
run show <profile> [name=value ...] [-- extra args...]
run params <profile>
run list
run complete profiles
run complete params <profile>
run complete values <profile> <param>
run last
run edit-last [n]
run repeat [n]
run history
```

### Intent

- `profile` selects the executable and its default parameter model.
- `name=value` overrides project-specific runtime parameters.
- `-- extra args...` passes through arbitrary trailing arguments without interpretation.
- `edit-last` prints a reusable `run ...` command reconstructed from history.
- `repeat` reruns a history entry directly.

## Design Decisions

### Why CLI-first

This tool is intentionally CLI-first rather than TUI-first because the main workflow is:

- reuse the previous command
- tweak one or two arguments
- rerun immediately

TUI can be added later for selection, but the core value comes from:

- profiles
- defaults
- completion
- replay/edit flows

### Generic params instead of hardcoded flags

Parameters such as `targets`, `column`, `users`, `place`, etc. vary per project.

Because of that, the model is not based on fixed flags like `--targets` or `--columns`.

Instead, profiles declare arbitrary named parameters and users pass:

```text
name=value
```

Examples:

```text
run main-app entities=alpha,beta fields=status,owner
run report-tool items=group-a,group-b region=region-1
```

### Defaults over “fixed” key-value args

Key-value arguments such as `dbdir=...` and `config=...` are modeled as params with defaults, not hard fixed args.

Reason:

- they usually look fixed
- but in development they often need temporary overrides

So the design is:

- key-value style arguments live in `params`
- if they are usually stable, they get a default
- only true raw fixed flags belong in `literal_args`

### Literal args

`literal_args` exist for non-`name=value` flags that should always be appended.

Example:

```toml
literal_args = ["--gui"]
```

Inheritance behavior for `literal_args` is concatenation:

- parent args first
- child args appended after them

### Values are suggestions by default

`values = [...]` are treated as candidate values for completion and warnings.

They are not strict constraints unless:

```toml
strict_values = true
```

This keeps the tool flexible during development while still surfacing suspicious inputs.

## Config Format

Config file name:

```text
run.toml
```

Discovery:

- search upward from the current directory
- `RUN_CONFIG=/path/to/run.toml` overrides discovery

### Current schema

Profiles are stored as a TOML table map.

Params are stored as an ordered TOML array because Go implementation is simpler and preserves output order.

Example:

```toml
version = 1

[global]
history_file = ".run/history.jsonl"
default_exec_mode = "exec"

[profiles.main-app]
bin = "build/app/main_app/Debug/main_app"

[[profiles.main-app.params]]
name = "data_dir"
kind = "string"
default = "/path/to/workdir/main-app-data"

[[profiles.main-app.params]]
name = "entities"
kind = "list"
multi = true
delimiter = ","
values = ["alpha", "beta", "gamma"]

[profiles.main-app-gui]
inherits = "main-app"
bin = "build/app/main_app_gui/Debug/main_app_gui"
literal_args = ["--gui"]
```

### Param semantics

Supported kinds:

- `string`
- `list`

Supported fields:

- `name`
- `kind`
- `required`
- `default` for `string`
- `default_list` for `list`
- `multi`
- `delimiter`
- `values`
- `strict_values`
- `help`

Rules:

- `params` render in definition order
- child profiles override parent params by matching `name`
- new child params append to the end
- same param specified multiple times on CLI: last wins
- except `kind=list` with `multi=true`, where repeated CLI occurrences append in order
- empty values are rejected

## Runtime Model

Internal flow is:

```text
Config -> resolved Profile -> Invocation -> argv
```

### Important internal rules

- `show` does not require the target binary to exist
- `exec` and `repeat` do require the binary to exist
- `repeat` rebuilds from structured history, not from a raw shell string
- `edit-last` rerenders history into a `run ...` command, not the final binary argv

## History

History is stored as JSONL at the configured `history_file`.

Each entry stores:

- timestamp
- profile name
- structured params
- extra args

This enables:

- `last`
- `history`
- `repeat`
- `edit-last`

## Shell Completion

There is a zsh completion script at:

`completions/_run`

Completion is intentionally driven by runtime subcommands:

- `run complete profiles`
- `run complete params <profile>`
- `run complete values <profile> <param>`

This keeps completion synchronized with `run.toml` instead of duplicating profile knowledge in shell code.

## Implementation Notes

Language:

- Go

Style:

- keep dependencies minimal
- use standard library where practical
- current external dependency is TOML parsing via `github.com/BurntSushi/toml`

Current package layout:

- `cmd/run/main.go`
- `internal/app/app.go`
- `internal/cli/parse.go`
- `internal/config/load.go`
- `internal/config/path.go`
- `internal/config/types.go`
- `internal/profile/resolve.go`
- `internal/invoke/build.go`
- `internal/history/history.go`
- `internal/render/render.go`
- `internal/exec/exec.go`

## Current Status

Implemented:

- config discovery
- profile inheritance
- ordered params
- `list`
- `params`
- `show`
- `exec`
- `history`
- `last`
- `repeat`
- `edit-last`
- completion helper subcommands
- zsh completion script
- basic unit tests

Validated during development:

- `go build ./...`
- `go test ./...`
- runtime checks for `list`, `params`, `show`, `exec`, `history`, `repeat`, `edit-last`

## Likely Next Steps

Most useful next improvements:

- replace generic sample profiles in `run.toml` with real project binaries
- add richer shell integration around `edit-last`
- add an optional “edit then execute” flow
- add bash/fish completion if needed
- add more tests around history replay and config validation edge cases
