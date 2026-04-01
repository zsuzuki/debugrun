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
env = { APP_MODE = "debug" }

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

[[profiles.main-app.params]]
name = "dir"
arg_name = "-dir"
arg_mode = "split"
kind = "list"
multi = true
```

`values` are treated as suggestions by default. Set `strict_values = true` to reject values outside that list.

`list` params with `multi = true` can be specified more than once on the `run` CLI. Repeated occurrences are appended in order, so `dir=VOL dir=TMP,WORK` becomes the same list as `dir=VOL,TMP,WORK`.

If you want to append to an existing default instead of replacing it, use `-add` before a single `name=value`. For example, `run app01 -add dir=TMP` keeps the configured default list and appends `TMP`.

If you want a `multi = true` list param to default to every configured candidate in `values`, set `default_all_values = true`.

By default params render as `name=value`. You can control the emitted shape with:

- `arg_mode = "kv"`: `name=value`
- `arg_mode = "equals"`: `arg_name=value`
- `arg_mode = "split"`: `arg_name value`
- `arg_name = "-dir"`: emitted argument name for `equals` or `split`
- `list_mode = "join"`: list values are emitted once, joined with `delimiter`
- `list_mode = "repeat"`: list values are emitted once per item

For `kind = "list"`, `arg_mode` and `list_mode` are independent:

- `kv` + `join` => `name=a,b,c`
- `equals` + `join` => `-name=a,b,c`
- `split` + `join` => `-name a,b,c`
- `kv` + `repeat` => `name=a name=b name=c`
- `equals` + `repeat` => `-name=a -name=b -name=c`
- `split` + `repeat` => `-name a -name b -name c`

Example:

```toml
[[profiles.app01.params]]
name = "dir"
arg_name = "-dir"
arg_mode = "split"
list_mode = "repeat"
kind = "list"
multi = true
values = ["VOL", "TMP", "WORK"]
default_all_values = true
```

If you want `-dir=VOL -dir=TMP`, use `arg_mode = "equals"` with `list_mode = "repeat"`.

You can use a few built-in variables in `run.toml` string fields such as `bin`, `literal_args`, `default`, `default_list`, `values`, and `global.history_file`:

- `${HOME}`: user home directory
- `${CWD}`: current working directory where `run` was invoked
- `${CONFIG_DIR}`: directory containing `run.toml`

Profiles can also define environment variables:

```toml
[profiles.main-app]
bin = "build/app/main_app/Debug/main_app"
env = { APP_MODE = "debug", DATA_DIR = "${CWD}/.data" }
```

Child profiles inherit parent `env` values and can override individual keys.

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
env = { APP_MODE = "debug" }

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

`multi = true` の `list` param は、`run` 側の CLI で同じ `name=value` を複数回書けます。繰り返し指定した値は順番どおりに連結されるため、`dir=VOL dir=TMP,WORK` は `dir=VOL,TMP,WORK` と同じリストになります。

既定値を上書きせずに追加したい場合は、1 個の `name=value` の直前に `-add` を付けます。例えば `run app01 -add dir=TMP` とすると、設定済みの既定リストを残したまま `TMP` を追加します。

`multi = true` の `list` param で、`values` に入っている候補を未指定時の既定値としてそのまま使いたい場合は `default_all_values = true` を指定します。

param の最終 argv への描画方法は `arg_mode` で選べます。

- `arg_mode = "kv"`: `name=value`
- `arg_mode = "equals"`: `arg_name=value`
- `arg_mode = "split"`: `arg_name value`
- `arg_name = "-dir"`: `equals` / `split` で使う実際の引数名
- `list_mode = "join"`: list を 1 回だけ出して `delimiter` で連結する
- `list_mode = "repeat"`: list の各要素を 1 回ずつ出す

`kind = "list"` では `arg_mode` と `list_mode` を組み合わせて決めます。

- `kv` + `join` => `name=a,b,c`
- `equals` + `join` => `-name=a,b,c`
- `split` + `join` => `-name a,b,c`
- `kv` + `repeat` => `name=a name=b name=c`
- `equals` + `repeat` => `-name=a -name=b -name=c`
- `split` + `repeat` => `-name a -name b -name c`

`run.toml` の文字列フィールドでは、`bin`、`literal_args`、`default`、`default_list`、`values`、`global.history_file` などでいくつかの組み込み変数を使えます。

- `${HOME}`: ユーザーのホームディレクトリ
- `${CWD}`: `run` を実行したカレントディレクトリ
- `${CONFIG_DIR}`: `run.toml` を含むディレクトリ

profile には環境変数も設定できます。

```toml
[profiles.main-app]
bin = "build/app/main_app/Debug/main_app"
env = { APP_MODE = "debug", DATA_DIR = "${CWD}/.data" }
```

子 profile は親の `env` を引き継ぎ、同じキーだけ個別に上書きできます。

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
