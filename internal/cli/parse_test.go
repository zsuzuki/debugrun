package cli

import "testing"

func TestParseCompleteValues(t *testing.T) {
	parsed, err := Parse([]string{"complete", "values", "main-app", "fields"})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Action != ActionComplete {
		t.Fatalf("action = %q", parsed.Action)
	}
	if parsed.Topic != "values" || parsed.ProfileName != "main-app" || parsed.ParamName != "fields" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseEditLastIndex(t *testing.T) {
	parsed, err := Parse([]string{"edit-last", "2"})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Action != ActionEditLast || parsed.RepeatIndex != 2 {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseAddModifier(t *testing.T) {
	parsed, err := Parse([]string{"main-app", "-add", "dir=TMP", "dir=WORK"})
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.RawParams) != 2 {
		t.Fatalf("raw params = %#v", parsed.RawParams)
	}
	if !parsed.RawParams[0].Append || parsed.RawParams[1].Append {
		t.Fatalf("raw params = %#v", parsed.RawParams)
	}
}

func TestParseAddModifierRequiresValue(t *testing.T) {
	_, err := Parse([]string{"main-app", "-add"})
	if err == nil || err.Error() != "expected name=value after -add" {
		t.Fatalf("err = %v", err)
	}
}
