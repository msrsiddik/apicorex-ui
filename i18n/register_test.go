package i18n

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/*.toml
var testCatalogs embed.FS

// The kit shipped exactly one catalogue — Schoolyze's — loaded by init() with
// no way for anyone else to add a key. These pin the behaviour that lets a
// second plugin have its own strings without taking Schoolyze's away.

func TestRegisterAddsKeysWithoutDisturbingTheKit(t *testing.T) {
	before := TranslateTo(Bangla, "app.name")
	Register(Bangla, map[string]string{"ledger.title": "খতিয়ান"})

	if got := TranslateTo(Bangla, "ledger.title"); got != "খতিয়ান" {
		t.Fatalf("registered key reads %q", got)
	}
	if after := TranslateTo(Bangla, "app.name"); after != before {
		t.Fatalf("registering moved an existing key: %q became %q", before, after)
	}
}

func TestRegisterOverwrites(t *testing.T) {
	// A consumer must be able to replace a string it disagrees with. Two
	// consumers cannot collide over this, because one process runs one plugin.
	Register(English, map[string]string{"probe.overwrite": "first"})
	Register(English, map[string]string{"probe.overwrite": "second"})
	if got := TranslateTo(English, "probe.overwrite"); got != "second" {
		t.Fatalf("got %q, want the later registration", got)
	}
}

func TestRegisterFSReadsEmbeddedTOML(t *testing.T) {
	// What a consumer actually does: keep the catalogue next to its screens and
	// embed it into its own binary.
	if err := RegisterFS(English, testCatalogs, "testdata/probe.toml"); err != nil {
		t.Fatalf("RegisterFS: %v", err)
	}
	if got := TranslateTo(English, "probe.nested.key"); got != "nested value" {
		t.Fatalf("nested TOML flattened to %q", got)
	}
}

func TestRegisterFSReportsAMissingFile(t *testing.T) {
	if err := RegisterFS(English, testCatalogs, "testdata/absent.toml"); err == nil {
		t.Fatal("a missing catalogue registered silently")
	}
}

func TestUnknownKeyStillRendersItself(t *testing.T) {
	// Unchanged, and deliberately ugly: a missing string should be obvious in
	// review rather than silently blank.
	if got := TranslateTo(Bangla, "no.such.key"); got != "no.such.key" {
		t.Fatalf("got %q", got)
	}
}

func TestFromRequestFallsBackToTheDefault(t *testing.T) {
	// The bug this exists to stop: ParseAcceptLanguage returns "" for a header
	// naming no language we serve, and an empty locale renders <html lang="">
	// with a Bangla interface showing every name in English.
	r := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	if got := FromRequest(r); got != Default {
		t.Fatalf("no headers at all gave %q, want the default", got)
	}
	r.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	if got := FromRequest(r); got != Default {
		t.Fatalf("an unserved language gave %q, want the default", got)
	}
}

func TestFromRequestPrefersTheExplicitChoice(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/accounts?lang=en", nil)
	r.Header.Set("Accept-Language", "bn")
	r.AddCookie(&http.Cookie{Name: "lang", Value: "bn"})
	if got := FromRequest(r); got != English {
		t.Fatalf("?lang lost to something else: %q", got)
	}
}

func TestFromRequestRemembersTheCookie(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	r.AddCookie(&http.Cookie{Name: "lang", Value: "en"})
	r.Header.Set("Accept-Language", "bn")
	if got := FromRequest(r); got != English {
		t.Fatalf("cookie lost to Accept-Language: %q", got)
	}
}

func TestFromRequestIgnoresNonsense(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/accounts?lang=klingon", nil)
	if got := FromRequest(r); got != Default {
		t.Fatalf("an invalid ?lang gave %q", got)
	}
}
