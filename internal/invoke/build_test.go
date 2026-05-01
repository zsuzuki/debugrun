package invoke

import (
	"path/filepath"
	"testing"

	"debugrun/internal/cli"
	"debugrun/internal/config"
)

func TestBuildAppliesDefaultsAndOverrides(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "data_dir", Kind: "string", Default: "/tmp/data"},
			{Name: "entities", Kind: "list", Multi: true, Delimiter: ",", DefaultList: []string{"item0"}},
			{Name: "fields", Kind: "list", Multi: true, Delimiter: ","},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams: []cli.RawParam{
			{Name: "entities", Value: "item1,item2"},
			{Name: "fields", Value: "status,owner"},
		},
		ExtraArgs: []string{"--verbose"},
	}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Params[0].Value.Scalar != "/tmp/data" {
		t.Fatalf("data_dir default = %q", inv.Params[0].Value.Scalar)
	}
	if len(inv.Params[1].Value.List) != 2 || inv.Params[1].Value.List[0] != "item1" {
		t.Fatalf("entities = %#v", inv.Params[1].Value.List)
	}
	if len(inv.ExtraArgs) != 1 || inv.ExtraArgs[0] != "--verbose" {
		t.Fatalf("extra args = %#v", inv.ExtraArgs)
	}
}

func TestBuildAcceptsParamAlias(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "value", Alias: "v", Kind: "string"},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams:   []cli.RawParam{{Name: "v", Value: "1"}},
	}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Params[0].Value.Scalar != "1" {
		t.Fatalf("value = %q", inv.Params[0].Value.Scalar)
	}
	if inv.Params[0].Spec.Name != "value" {
		t.Fatalf("canonical name = %q", inv.Params[0].Spec.Name)
	}
}

func TestBuildAppendsRepeatedMultiListParams(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "dir", ArgName: "-dir", ArgMode: "split", Kind: "list", Multi: true, Delimiter: ",", DefaultList: []string{"OLD"}},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams: []cli.RawParam{
			{Name: "dir", Value: "VOL"},
			{Name: "dir", Value: "TMP,WORK"},
		},
	}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if got := inv.Params[0].Value.List; len(got) != 3 || got[0] != "VOL" || got[1] != "TMP" || got[2] != "WORK" {
		t.Fatalf("dir = %#v", got)
	}
}

func TestBuildUsesAllValuesAsDefaultForMultiList(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "dir", Kind: "list", Multi: true, Delimiter: ",", Values: []string{"VOL", "TMP", "WORK"}, DefaultAllValues: true},
		},
	}
	parsed := &cli.Parsed{ProfileName: "main-app"}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if got := inv.Params[0].Value.List; len(got) != 3 || got[0] != "VOL" || got[1] != "TMP" || got[2] != "WORK" {
		t.Fatalf("dir default = %#v", got)
	}
}

func TestBuildExplicitMultiListOverridesDefaultAllValues(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "dir", Kind: "list", Multi: true, Delimiter: ",", Values: []string{"VOL", "TMP", "WORK"}, DefaultAllValues: true},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams:   []cli.RawParam{{Name: "dir", Value: "TMP"}},
	}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if got := inv.Params[0].Value.List; len(got) != 1 || got[0] != "TMP" {
		t.Fatalf("dir override = %#v", got)
	}
}

func TestBuildAddAppendsToDefaultList(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "dir", Kind: "list", Multi: true, Delimiter: ",", DefaultList: []string{"BASE"}},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams:   []cli.RawParam{{Name: "dir", Value: "TMP,WORK", Append: true}},
	}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if got := inv.Params[0].Value.List; len(got) != 3 || got[0] != "BASE" || got[1] != "TMP" || got[2] != "WORK" {
		t.Fatalf("dir add = %#v", got)
	}
}

func TestBuildAddRejectsNonMultiList(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "region", Kind: "string"},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams:   []cli.RawParam{{Name: "region", Value: "jp", Append: true}},
	}

	_, err := Build(profile, parsed)
	if err == nil || err.Error() != `param "region" does not support -add` {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateReturnsWarningsForCandidateMiss(t *testing.T) {
	profile := &config.Profile{
		Name: "main-app",
		Bin:  "/bin/echo",
		Params: []config.ParamSpec{
			{Name: "region", Kind: "string", Values: []string{"region-1", "region-2"}},
		},
	}
	parsed := &cli.Parsed{
		ProfileName: "main-app",
		RawParams:   []cli.RawParam{{Name: "region", Value: "region-x"}},
	}

	inv, err := Build(profile, parsed)
	if err != nil {
		t.Fatal(err)
	}
	warnings, err := Validate(inv, profile, filepath.Dir("/bin/echo"), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings len = %d", len(warnings))
	}
	if warnings[0].Param != "region" {
		t.Fatalf("warning param = %q", warnings[0].Param)
	}
}
