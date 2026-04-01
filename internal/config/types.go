package config

type Config struct {
	Version  int                `toml:"version"`
	Global   GlobalConfig       `toml:"global"`
	Profiles map[string]Profile `toml:"profiles"`
}

type GlobalConfig struct {
	HistoryFile     string `toml:"history_file"`
	DefaultExecMode string `toml:"default_exec_mode"`
}

type Profile struct {
	Name        string      `toml:"-"`
	Bin         string      `toml:"bin"`
	Inherits    string      `toml:"inherits"`
	LiteralArgs []string    `toml:"literal_args"`
	Params      []ParamSpec `toml:"params"`
}

type ParamSpec struct {
	Name         string   `toml:"name"`
	Kind         string   `toml:"kind"`
	Required     bool     `toml:"required"`
	Multi        bool     `toml:"multi"`
	Delimiter    string   `toml:"delimiter"`
	Values       []string `toml:"values"`
	StrictValues bool     `toml:"strict_values"`
	Help         string   `toml:"help"`
	Default      string   `toml:"default"`
	DefaultList  []string `toml:"default_list"`
}
