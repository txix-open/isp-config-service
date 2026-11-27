package variable

import (
	"encoding/json"
	"strings"
	"testing"

	"isp-config-service/entity"
)

func textVar(name, value string) entity.Variable {
	return entity.Variable{Name: name, Type: entity.TextType, Value: value}
}

func secretVar(name, value string) entity.Variable {
	return entity.Variable{Name: name, Type: entity.SecretType, Value: value}
}

func TestRenderSubstitutesNonUI(t *testing.T) {
	out, skipped, undefined, err := Render(
		[]byte(`{"a":"_Var(v1)"}`),
		nil,
		map[string]entity.Variable{"v1": textVar("v1", "x")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"a":"x"}` {
		t.Fatalf("out = %s, want %s", out, `{"a":"x"}`)
	}
	if len(skipped) != 0 || len(undefined) != 0 {
		t.Fatalf("skipped = %v, undefined = %v, want both empty", skipped, undefined)
	}
}

func TestRenderKeepsSecretInUI(t *testing.T) {
	out, skipped, undefined, err := Render(
		[]byte(`{"a":"_Var(s1)"}`),
		[]string{"$.a"},
		map[string]entity.Variable{"s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"a":"_Var(s1)"}` {
		t.Fatalf("out = %s, want the literal placeholder", out)
	}
	if strings.Contains(string(out), "top") {
		t.Fatalf("the secret value leaked into the output: %s", out)
	}
	if len(skipped) != 1 || skipped[0].Path != "$.a" || skipped[0].Variable != "s1" {
		t.Fatalf("skipped = %v, want [{$.a s1}]", skipped)
	}
	if len(undefined) != 0 {
		t.Fatalf("undefined = %v, want empty", undefined)
	}
}

func TestRenderSameVarUINonUI(t *testing.T) {
	out, skipped, _, err := Render(
		[]byte(`{"ui":"_Var(s1)","plain":"_Var(s1)"}`),
		[]string{"$.ui"},
		map[string]entity.Variable{"s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"plain":"top","ui":"_Var(s1)"}` {
		t.Fatalf("out = %s, want the non-UI leaf substituted and the UI leaf left as a literal", out)
	}
	if len(skipped) != 1 || skipped[0].Path != "$.ui" {
		t.Fatalf("skipped = %v, want [{$.ui s1}]", skipped)
	}
}

func TestRenderMixedUILeaf(t *testing.T) {
	out, skipped, _, err := Render(
		[]byte(`{"a":"p/_Var(t1)/_Var(s1)"}`),
		[]string{"$.a"},
		map[string]entity.Variable{"t1": textVar("t1", "tv"), "s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"a":"p/tv/_Var(s1)"}` {
		t.Fatalf("out = %s, want %s", out, `{"a":"p/tv/_Var(s1)"}`)
	}
	if len(skipped) != 1 || skipped[0].Variable != "s1" {
		t.Fatalf("skipped = %v, want [{$.a s1}]", skipped)
	}
}

func TestRenderSecretOutsideUI(t *testing.T) {
	out, skipped, _, err := Render(
		[]byte(`{"a":"_Var(s1)"}`),
		nil,
		map[string]entity.Variable{"s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"a":"top"}` {
		t.Fatalf("out = %s, want %s", out, `{"a":"top"}`)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v, want empty", skipped)
	}
}

func TestRenderArrayPath(t *testing.T) {
	out, skipped, _, err := Render(
		[]byte(`{"oauth":{"clients":[{"e":"_Var(s1)"},{"e":"x"}]}}`),
		[]string{"$.oauth.clients[].e"},
		map[string]entity.Variable{"s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"oauth":{"clients":[{"e":"_Var(s1)"},{"e":"x"}]}}` {
		t.Fatalf("out = %s, want the literal in the UI array leaf", out)
	}
	if len(skipped) != 1 || skipped[0].Path != "$.oauth.clients[].e" || skipped[0].Variable != "s1" {
		t.Fatalf("skipped = %v, want [{$.oauth.clients[].e s1}]", skipped)
	}
}

func TestRenderUndefined(t *testing.T) {
	out, skipped, undefined, err := Render(
		[]byte(`{"a":"_Var(ghost)"}`),
		nil,
		map[string]entity.Variable{},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(undefined) != 1 || undefined[0] != "ghost" {
		t.Fatalf("undefined = %v, want [ghost]", undefined)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v, want empty", skipped)
	}
	if string(out) != `{"a":"_Var(ghost)"}` {
		t.Fatalf("out = %s, the unknown reference is left as-is", out)
	}
}

func TestRenderUndefinedInUI(t *testing.T) {
	// A variable that does not exist cannot be proven SECRET, so it is
	// reported undefined instead of silently kept as a literal.
	_, skipped, undefined, err := Render(
		[]byte(`{"a":"_Var(ghost)"}`),
		[]string{"$.a"},
		map[string]entity.Variable{},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(undefined) != 1 || undefined[0] != "ghost" {
		t.Fatalf("undefined = %v, want [ghost]", undefined)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v, want empty", skipped)
	}
}

func TestRenderToInt(t *testing.T) {
	out, _, _, err := Render(
		[]byte(`{"n":"_ToInt(_Var(n1))"}`),
		nil,
		map[string]entity.Variable{"n1": textVar("n1", "42")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"n":42}` {
		t.Fatalf("out = %s, want the number 42", out)
	}
}

func TestRenderToIntSecretInUI(t *testing.T) {
	out, skipped, _, err := Render(
		[]byte(`{"n":"_ToInt(_Var(s1))"}`),
		[]string{"$.n"},
		map[string]entity.Variable{"s1": secretVar("s1", "42")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"n":"_ToInt(_Var(s1))"}` {
		t.Fatalf("out = %s, the _ToInt wrapper must not break the literal", out)
	}
	if len(skipped) != 1 || skipped[0].Variable != "s1" {
		t.Fatalf("skipped = %v, want [{$.n s1}]", skipped)
	}
}

func TestRenderKeysUntouched(t *testing.T) {
	out, _, _, err := Render(
		[]byte(`{"_Var(k)":"_Var(v1)"}`),
		nil,
		map[string]entity.Variable{"v1": textVar("v1", "x")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"_Var(k)":"x"}` {
		t.Fatalf("out = %s, JSON keys must not be substituted", out)
	}
}

func TestRenderSkippedDeduped(t *testing.T) {
	_, skipped, _, err := Render(
		[]byte(`{"a":"_Var(s1)_Var(s1)"}`),
		[]string{"$.a"},
		map[string]entity.Variable{"s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(skipped) != 1 {
		t.Fatalf("skipped = %v, want a single entry", skipped)
	}
}

func TestRenderNonStringUILeaf(t *testing.T) {
	out, skipped, _, err := Render(
		[]byte(`{"n":5}`),
		[]string{"$.n"},
		map[string]entity.Variable{},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"n":5}` {
		t.Fatalf("out = %s, non-string leaves are untouched", out)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v, want empty", skipped)
	}
}

func TestRenderInvalidJSON(t *testing.T) {
	_, _, _, err := Render(
		[]byte(`{"a":`),
		nil,
		map[string]entity.Variable{},
	)
	if err == nil {
		t.Fatal("Render: want an error for invalid JSON")
	}
}

func TestRenderEscapes(t *testing.T) {
	out, _, _, err := Render(
		[]byte(`{"a":"line1\nline2 _Var(v1)"}`),
		nil,
		map[string]entity.Variable{"v1": textVar("v1", "x")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(out) != `{"a":"line1\nline2 x"}` {
		t.Fatalf("out = %s, JSON escapes must survive the re-serialization", out)
	}
}

func TestRenderCompositeInUISubstituted(t *testing.T) {
	// a composite JSON string on a UI path is report-only: references are
	// substituted, the secret policy does not apply
	composite, err := json.Marshal(map[string]any{"a": "_Var(s1)"})
	if err != nil {
		t.Fatalf("marshal composite: %v", err)
	}
	doc, err := json.Marshal(map[string]any{
		"secrets": string(composite),
		"plain":   "_Var(s1)",
	})
	if err != nil {
		t.Fatalf("marshal document: %v", err)
	}

	out, skipped, undefined, err := Render(
		doc,
		[]string{"$.secrets", "$.plain"},
		map[string]entity.Variable{"s1": secretVar("s1", "top")},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got := m["secrets"].(string); !strings.Contains(got, "top") {
		t.Fatalf("$.secrets = %s, want the secret value substituted", got)
	}
	// the simple leaf on a UI path still follows the policy
	if got := m["plain"]; got != "_Var(s1)" {
		t.Fatalf("$.plain = %v, want the literal placeholder", got)
	}
	if len(skipped) != 1 || skipped[0].Path != "$.plain" || skipped[0].Variable != "s1" {
		t.Fatalf("skipped = %v, want [{$.plain s1}]", skipped)
	}
	if len(undefined) != 0 {
		t.Fatalf("undefined = %v, want empty", undefined)
	}
}

func TestRenderCompositeUndefinedCollected(t *testing.T) {
	composite, err := json.Marshal([]any{"_Var(ghost)"})
	if err != nil {
		t.Fatalf("marshal composite: %v", err)
	}
	doc, err := json.Marshal(map[string]any{"secrets": string(composite)})
	if err != nil {
		t.Fatalf("marshal document: %v", err)
	}

	out, skipped, undefined, err := Render(
		doc,
		[]string{"$.secrets"},
		map[string]entity.Variable{},
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got, want := m["secrets"], string(composite); got != want {
		t.Fatalf("$.secrets = %v, want %s (the unknown reference stays a literal)", got, want)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v, want empty", skipped)
	}
	if len(undefined) != 1 || undefined[0] != "ghost" {
		t.Fatalf("undefined = %v, want [ghost]", undefined)
	}
}
