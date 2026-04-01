package render

import (
	"reflect"
	"testing"

	"debugrun/internal/config"
	"debugrun/internal/history"
	"debugrun/internal/invoke"
)

func TestParamHelpers(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Params: []config.ParamSpec{
			{Name: "data_dir", Kind: "string"},
			{Name: "fields", Kind: "list", Values: []string{"status", "score"}},
		},
	}

	if got := ParamNames(profile); !reflect.DeepEqual(got, []string{"data_dir", "fields"}) {
		t.Fatalf("ParamNames() = %#v", got)
	}
	if got := ParamValues(profile, "fields"); !reflect.DeepEqual(got, []string{"status", "score"}) {
		t.Fatalf("ParamValues() = %#v", got)
	}
}

func TestFormatReplayCommand(t *testing.T) {
	entry := history.Entry{
		ProfileName: "main-app",
		Params: []history.Param{
			{Name: "data_dir", Kind: "string", Scalar: "/tmp/data"},
			{Name: "fields", Kind: "list", List: []string{"status", "score"}},
		},
		ExtraArgs: []string{"--verbose", "--path=/tmp/a b"},
	}

	got := FormatReplayCommand("run", entry)
	want := "run main-app data_dir=/tmp/data fields=status,score -- --verbose '--path=/tmp/a b'"
	if got != want {
		t.Fatalf("FormatReplayCommand() = %q, want %q", got, want)
	}
}

func TestArgvRendersSplitModeParams(t *testing.T) {
	inv := &invoke.Invocation{
		Bin: "/bin/echo",
		Params: []invoke.BoundParam{
			{Spec: config.ParamSpec{Name: "config", ArgName: "--config", ArgMode: "split", Kind: "string"}, Value: invoke.Value{Scalar: "/tmp/app.toml"}},
			{Spec: config.ParamSpec{Name: "dir", ArgName: "-dir", ArgMode: "split", Kind: "list", Multi: true, Delimiter: ","}, Value: invoke.Value{List: []string{"VOL", "TMP", "WORK"}}},
		},
	}

	got := Argv(inv, "/tmp")
	want := []string{"/bin/echo", "--config", "/tmp/app.toml", "-dir", "VOL", "-dir", "TMP", "-dir", "WORK"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv() = %#v, want %#v", got, want)
	}
}

func TestArgvRendersDifferentArgModes(t *testing.T) {
	inv := &invoke.Invocation{
		Bin: "/bin/echo",
		Params: []invoke.BoundParam{
			{Spec: config.ParamSpec{Name: "region", Kind: "string", ArgMode: "kv"}, Value: invoke.Value{Scalar: "jp"}},
			{Spec: config.ParamSpec{Name: "config", ArgName: "--config", ArgMode: "equals", Kind: "string"}, Value: invoke.Value{Scalar: "/tmp/app.toml"}},
			{Spec: config.ParamSpec{Name: "dir", ArgName: "-dir", ArgMode: "split", Kind: "list", Multi: true, Delimiter: ","}, Value: invoke.Value{List: []string{"VOL", "TMP"}}},
		},
	}

	got := Argv(inv, "/tmp")
	want := []string{"/bin/echo", "region=jp", "--config=/tmp/app.toml", "-dir", "VOL", "-dir", "TMP"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv() = %#v, want %#v", got, want)
	}
}
