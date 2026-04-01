package config

import "github.com/BurntSushi/toml"

func tomlDecodeFile(path string, v any) (toml.MetaData, error) {
	return toml.DecodeFile(path, v)
}
