// Package i18n provides the web panel's bilingual strings.
//
// Every user-facing string in the admin panel goes through Translate from the
// first line of code. Retrofitting i18n means revisiting every template, so
// there is no "add it later" — a bare Bangla or English literal in markup is a
// bug, not a shortcut.
//
// Bangla is the default. The institutions this serves are Bangladeshi, their
// office staff read Bangla first, and defaulting to English would make the
// common case the one that needs configuring.
package i18n

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

// Locale identifies a supported language.
type Locale string

const (
	// Bangla is the default: the software is for Bangladeshi institutions.
	Bangla  Locale = "bn"
	English Locale = "en"
)

// Default is used when nothing else resolves.
const Default = Bangla

// Supported lists the locales a user may choose, in display order.
var Supported = []Locale{Bangla, English}

// Valid reports whether s names a supported locale.
func Valid(s string) bool {
	for _, l := range Supported {
		if string(l) == s {
			return true
		}
	}
	return false
}

// Name returns a locale's name written in that locale, for the language picker.
// A user looking for Bangla should see "বাংলা", not "Bangla".
func (l Locale) Name() string {
	switch l {
	case Bangla:
		return "বাংলা"
	case English:
		return "English"
	default:
		return string(l)
	}
}

// IsBangla reports whether this locale renders Bangla, which decides numeral
// formatting and the font stack.
func (l Locale) IsBangla() bool { return l == Bangla }

//go:embed *.toml
var localeFS embed.FS

// catalogs holds the loaded strings, keyed by locale.
//
// Guarded, because a second consumer registers its own catalogue at startup
// while the first request may already be in flight in a test.
var (
	catalogMu sync.RWMutex
	catalogs  = map[Locale]map[string]string{}
)

func init() {
	// The kit's own strings, namespaced kit.* — see Register for why the
	// namespace matters. Consumers add theirs on top.
	for _, l := range Supported {
		c, err := load(string(l) + ".toml")
		if err != nil {
			// A missing or malformed catalog is a build-time mistake: the files
			// are embedded, so this cannot fail at runtime for environmental
			// reasons. Failing loudly at startup beats shipping an interface
			// full of raw key names.
			panic(fmt.Sprintf("i18n: load %s: %v", l, err))
		}
		catalogs[l] = c
	}
}

// Register merges a consumer's strings into the catalogue for one locale.
//
// The kit used to ship exactly one catalogue — Schoolyze's, 795 lines of it,
// loaded by init() into a package-level map with no way for anyone else to add
// a key. That was fine while there was one consumer and became the first thing
// in the way when there were two: a second plugin's title rendered as
// "স্কুলাইজ", because app.name could only mean one product.
//
// Keys already present are overwritten, so a consumer can also replace a kit
// string it disagrees with. Registering the same key from two consumers in one
// process would be a genuine conflict, but that cannot happen: one process runs
// one plugin.
func Register(l Locale, catalog map[string]string) {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	if catalogs[l] == nil {
		catalogs[l] = map[string]string{}
	}
	for k, v := range catalog {
		catalogs[l][k] = v
	}
}

// RegisterFS reads a TOML catalogue out of an embedded filesystem and registers
// it, which is what every consumer actually wants: the file lives next to the
// screens it belongs to and is embedded into the binary.
func RegisterFS(l Locale, fsys fs.FS, name string) error {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("i18n: read %s: %w", name, err)
	}
	flat, err := parse(raw)
	if err != nil {
		return fmt.Errorf("i18n: parse %s: %w", name, err)
	}
	Register(l, flat)
	return nil
}

func load(name string) (map[string]string, error) {
	raw, err := localeFS.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return parse(raw)
}

func parse(raw []byte) (map[string]string, error) {
	var nested map[string]any
	if err := toml.Unmarshal(raw, &nested); err != nil {
		return nil, err
	}
	flat := map[string]string{}
	flatten("", nested, flat)
	return flat, nil
}

// flatten turns TOML's nested tables into dotted keys, so the files can be
// grouped for humans ([student] name = "...") while lookup stays a flat map
// ("student.name").
func flatten(prefix string, in map[string]any, out map[string]string) {
	for k, v := range in {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]any:
			flatten(key, val, out)
		}
	}
}

// contextKey is the locale's slot on a request context.
type contextKey struct{}

// WithLocale returns ctx carrying the resolved locale.
func WithLocale(ctx context.Context, l Locale) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext returns the request's locale, or the default when none was set.
func FromContext(ctx context.Context) Locale {
	if l, ok := ctx.Value(contextKey{}).(Locale); ok {
		return l
	}
	return Default
}

// T translates a key for the context's locale.
//
// A missing key falls back to the other locale before giving up, so a string
// added in Bangla but not yet in English still renders something readable
// rather than a bare key. An untranslated key returns itself, which is ugly on
// purpose: it should be obvious in review, not silently blank.
func T(ctx context.Context, key string, args ...any) string {
	return TranslateTo(FromContext(ctx), key, args...)
}

// TranslateTo translates a key for a specific locale.
func TranslateTo(l Locale, key string, args ...any) string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	if s, ok := catalogs[l][key]; ok {
		return format(s, args...)
	}
	// Fall back to the other locale: half-translated is better than missing.
	for _, alt := range Supported {
		if alt == l {
			continue
		}
		if s, ok := catalogs[alt][key]; ok {
			return format(s, args...)
		}
	}
	return key
}

func format(s string, args ...any) string {
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Missing reports keys present in one locale but absent from another.
//
// Used by a test so a string added to one file and forgotten in the other
// fails the build rather than surfacing to a user as a stray English label in
// a Bangla interface.
func Missing(from, to Locale) []string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	var out []string
	for k := range catalogs[from] {
		if _, ok := catalogs[to][k]; !ok {
			out = append(out, k)
		}
	}
	return out
}

// Keys returns every key in a locale, sorted-insensitive (callers sort if they
// need order). Exposed for tests and tooling.
func Keys(l Locale) []string {
	out := make([]string, 0, len(catalogs[l]))
	for k := range catalogs[l] {
		out = append(out, k)
	}
	return out
}

// ParseAcceptLanguage picks the best supported locale from an Accept-Language
// header, or "" when none matches.
//
// Deliberately simple: this is a fallback for a user who has not chosen yet,
// not a full RFC 4647 negotiation. The quality values are ignored because with
// two locales the first recognised match is the right answer.
func ParseAcceptLanguage(header string) Locale {
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(part)
		if i := strings.Index(tag, ";"); i >= 0 {
			tag = tag[:i]
		}
		// "bn-BD" and "bn" both mean Bangla.
		if i := strings.Index(tag, "-"); i >= 0 {
			tag = tag[:i]
		}
		if Valid(tag) {
			return Locale(tag)
		}
	}
	return ""
}
