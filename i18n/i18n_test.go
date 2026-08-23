package i18n

import (
	"context"
	"sort"
	"strings"
	"testing"
)

// The two catalogs must stay in lockstep. A string added to one file and
// forgotten in the other surfaces to a user as a stray English label in a
// Bangla interface — visible to them, invisible in review. Failing here is the
// only reliable way to catch it.
func TestCatalogsMatch(t *testing.T) {
	for _, pair := range []struct{ from, to Locale }{
		{Bangla, English},
		{English, Bangla},
	} {
		missing := Missing(pair.from, pair.to)
		if len(missing) == 0 {
			continue
		}
		sort.Strings(missing)
		t.Errorf("%d key(s) in %s.toml missing from %s.toml:\n  %s",
			len(missing), pair.from, pair.to, strings.Join(missing, "\n  "))
	}
}

// Every catalog entry must actually say something. An empty value renders as a
// blank label, which looks like a layout bug rather than a missing translation.
func TestNoEmptyStrings(t *testing.T) {
	for _, l := range Supported {
		for _, k := range Keys(l) {
			if strings.TrimSpace(TranslateTo(l, k)) == "" {
				t.Errorf("%s.toml: %q is empty", l, k)
			}
		}
	}
}

// Format placeholders must match across locales: a string with %s in Bangla and
// none in English would print a literal "%!s(MISSING)" to whoever switched.
func TestPlaceholdersMatch(t *testing.T) {
	for _, k := range Keys(Bangla) {
		bn := TranslateTo(Bangla, k)
		en := TranslateTo(English, k)
		if got, want := strings.Count(en, "%s"), strings.Count(bn, "%s"); got != want {
			t.Errorf("%q has %d %%s in Bangla but %d in English", k, want, got)
		}
	}
}

func TestT_UsesContextLocale(t *testing.T) {
	// A kit.* key, because that is all this module ships now. Product strings
	// belong to the product and are registered by it.
	bn := T(WithLocale(context.Background(), Bangla), "kit.action.cancel")
	en := T(WithLocale(context.Background(), English), "kit.action.cancel")

	if bn != "বাতিল" {
		t.Errorf("Bangla cancel = %q", bn)
	}
	if en != "Cancel" {
		t.Errorf("English cancel = %q", en)
	}
}

// Bangla is the default because the institutions this serves are Bangladeshi.
// A context with no locale must not fall back to English.
func TestT_DefaultsToBangla(t *testing.T) {
	if got := T(context.Background(), "kit.action.cancel"); got != "বাতিল" {
		t.Errorf("default locale cancel = %q, want Bangla", got)
	}
}

// An unknown key returns itself rather than an empty string: visible in review,
// not silently blank in the interface.
func TestT_UnknownKeyReturnsKey(t *testing.T) {
	if got := T(context.Background(), "no.such.key"); got != "no.such.key" {
		t.Errorf("unknown key = %q, want the key itself", got)
	}
}

func TestT_Formats(t *testing.T) {
	got := TranslateTo(English, "kit.table.showing", "1", "50", "1,247")
	if got != "Showing 1–50 of 1,247" {
		t.Errorf("formatted = %q", got)
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	cases := map[string]Locale{
		"bn":                      Bangla,
		"bn-BD,bn;q=0.9,en;q=0.8": Bangla,
		"en-US,en;q=0.9":          English,
		"fr-FR,fr;q=0.9,en;q=0.5": English, // first *supported* match
		"fr,de":                   "",      // nothing supported
		"":                        "",
	}
	for header, want := range cases {
		if got := ParseAcceptLanguage(header); got != want {
			t.Errorf("ParseAcceptLanguage(%q) = %q, want %q", header, got, want)
		}
	}
}

// The picker shows each language in its own script — someone looking for
// Bangla should see "বাংলা", not "Bangla".
func TestLocaleName(t *testing.T) {
	if Bangla.Name() != "বাংলা" {
		t.Errorf("Bangla name = %q", Bangla.Name())
	}
	if English.Name() != "English" {
		t.Errorf("English name = %q", English.Name())
	}
}

func TestValid(t *testing.T) {
	for _, ok := range []string{"bn", "en"} {
		if !Valid(ok) {
			t.Errorf("Valid(%q) = false", ok)
		}
	}
	for _, bad := range []string{"", "fr", "BN", "bn-BD"} {
		if Valid(bad) {
			t.Errorf("Valid(%q) = true", bad)
		}
	}
}

// A duplicate table heading makes the TOML unparseable, which panics in init
// and takes the whole binary down at startup.
//
// Note what this test can and cannot do: init runs before it, so a duplicate
// already fails the package with a panic and a stack trace. Keeping the check
// is still worth it — when someone hits this the assertion names both line
// numbers, which the panic does not, and that is the difference between a
// minute and twenty.
func TestTOMLHasNoDuplicateTables(t *testing.T) {
	for _, name := range []string{"bn.toml", "en.toml"} {
		raw, err := localeFS.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		seen := map[string]int{}
		for i, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
				continue
			}
			if first, dup := seen[trimmed]; dup {
				t.Errorf("%s: table %s declared twice, at lines %d and %d — this panics at startup",
					name, trimmed, first, i+1)
				continue
			}
			seen[trimmed] = i + 1
		}
	}
}
