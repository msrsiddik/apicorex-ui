package components

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

// A control's DOM id is what its own label points at, so getting it wrong is
// quiet: the page renders, and clicking a label focuses somebody else's field.

func renderField(t *testing.T, p FieldProps) string {
	t.Helper()
	var sb strings.Builder
	if err := Field(p).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func renderSelect(t *testing.T, p SelectProps) string {
	t.Helper()
	var sb strings.Builder
	if err := SelectField(p).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func TestFieldIDDefaultsToTheName(t *testing.T) {
	// Every caller that predates FieldProps.ID must render exactly what it did.
	html := renderField(t, FieldProps{Name: "email", Label: "Email"})
	if !strings.Contains(html, `id="email"`) || !strings.Contains(html, `for="email"`) {
		t.Fatalf("id or label lost the name: %s", html)
	}
}

func TestFieldIDCanBeSetPerRow(t *testing.T) {
	// The case this exists for: a table of rows each carrying one "account_code".
	// Without a per-row id every row's label points at the first row's control.
	rows := []string{"gl-method-cash", "gl-method-bkash"}
	var all strings.Builder
	for _, id := range rows {
		all.WriteString(renderField(t, FieldProps{Name: "account_code", ID: id, Label: id}))
	}
	seen := map[string]bool{}
	for _, m := range regexp.MustCompile(`id="([^"]+)"`).FindAllStringSubmatch(all.String(), -1) {
		if seen[m[1]] {
			t.Fatalf("duplicate DOM id %q across rows", m[1])
		}
		seen[m[1]] = true
	}
	// The submitted name stays shared — that is what makes them one field
	// repeated rather than two different ones.
	if strings.Count(all.String(), `name="account_code"`) != len(rows) {
		t.Error("the form field name should not change with the id")
	}
}

func TestFieldErrorIsDescribedByTheRowsOwnMessage(t *testing.T) {
	// aria-describedby has to point at this row's error, not at the first row's.
	html := renderField(t, FieldProps{
		Name: "account_code", ID: "gl-method-cash", Error: "unknown account",
	})
	if !strings.Contains(html, `aria-describedby="gl-method-cash-error"`) {
		t.Errorf("aria-describedby not scoped to the row: %s", html)
	}
	if !strings.Contains(html, `id="gl-method-cash-error"`) {
		t.Errorf("the error paragraph is not the one described: %s", html)
	}
}

func TestSelectIDBehavesTheSameWay(t *testing.T) {
	html := renderSelect(t, SelectProps{Name: "account_code", Label: "Account"})
	if !strings.Contains(html, `id="account_code"`) {
		t.Errorf("default id lost: %s", html)
	}
	html = renderSelect(t, SelectProps{
		Name: "account_code", ID: "gl-receivable", Label: "Receivable",
		Error: "required",
	})
	for _, want := range []string{`id="gl-receivable"`, `for="gl-receivable"`, `aria-describedby="gl-receivable-error"`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in %s", want, html)
		}
	}
}
