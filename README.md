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

You can use a few built-in variables in `run.toml` string fields such as `bin`, `literal_args`, `default`, `default_list`, `values`, and `global.history_file`:

- `${HOME}`: user home directory
- `${CWD}`: current working directory where `run` was invoked
- `${CONFIG_DIR}`: directory containing `run.toml`

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

## 日本語版

`run` は、プロファイルごとのデフォルト値を使ってプロジェクト固有のバイナリを起動しやすくする、小さな Go 製 CLI です。

## コマンド

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

## 設定ファイルの探索

- `run` は現在のディレクトリから親方向へ `run.toml` を探索します。
- `RUN_CONFIG=/path/to/run.toml` を設定すると探索結果を上書きできます。

## 設定ファイルの形式

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

`values` はデフォルトでは候補値として扱われます。リスト外の値を拒否したい場合は `strict_values = true` を指定します。

`run.toml` の文字列フィールドでは、`bin`、`literal_args`、`default`、`default_list`、`values`、`global.history_file` などでいくつかの組み込み変数を使えます。

- `${HOME}`: ユーザーのホームディレクトリ
- `${CWD}`: `run` を実行したカレントディレクトリ
- `${CONFIG_DIR}`: `run.toml` を含むディレクトリ

## 補完用ヘルパー

以下のコマンドは、シェル統合向けに 1 行 1 項目で出力します。

```text
run complete profiles
run complete params main-app
run complete values main-app fields
```

## zsh 補完

このリポジトリには `completions/_run` に zsh 補完スクリプトが含まれています。

設定例:

```zsh
fpath=(/path/to/debugrun/completions $fpath)
autoload -Uz compinit
compinit
```

このスクリプトは `run complete profiles|params|values` を利用するため、`run.toml` と補完内容がずれにくくなっています。

## 直前コマンドの再編集

`run edit-last` は、直近の履歴エントリを再利用可能な `run ...` コマンドとして再構築して出力します。

```text
run edit-last
run edit-last 2
```
