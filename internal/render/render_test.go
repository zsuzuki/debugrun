package render

import (
	"reflect"
	"testing"

	"debugrun/internal/config"
	"debugrun/internal/history"
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
