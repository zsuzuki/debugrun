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
