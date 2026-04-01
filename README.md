# run

Small Go CLI for launching project-specific binaries with profile-based defaults.

## Commands

```text
run <profile> [name=value ...] [-- extra args...]
run show <profile> [name=value ...] [-- extra args...]
run params <profile>
run complete profiles
run complete params <profile>
run complete values <profile> <param>
run list
run last
run edit-last [n]
run repeat [n]
run history
```

## Config discovery

- `run` looks for `run.toml` from the current directory upward.
- Set `RUN_CONFIG=/path/to/run.toml` to override discovery.

## Config shape

```toml
version = 1

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
```

`values` are treated as suggestions by default. Set `strict_values = true` to reject values outside that list.

## Completion helpers

These commands print one item per line for shell integration:

```text
run complete profiles
run complete params main-app
run complete values main-app fields
```

## Zsh completion

The repository includes a zsh completion script at `completions/_run`.

Example setup:

```zsh
fpath=(/path/to/debugrun/completions $fpath)
autoload -Uz compinit
compinit
```

The script uses `run complete profiles|params|values`, so it stays in sync with your `run.toml`.

## Re-editing the last command

`run edit-last` prints the most recent history entry back as a reusable `run ...` command.

```text
run edit-last
run edit-last 2
```
