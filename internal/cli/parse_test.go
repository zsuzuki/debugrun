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
