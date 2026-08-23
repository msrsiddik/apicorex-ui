package bn

import "testing"

func TestBN(t *testing.T) {
	cases := map[string]string{
		"0":          "০",
		"1234567890": "১২৩৪৫৬৭৮৯০",
		"78":         "৭৮",
		// Separators and non-digits pass through: a grade stays Latin, which is
		// how Bangladeshi report cards write it.
		"1,250.50":  "১,২৫০.৫০",
		"A+":        "A+",
		"2026-0457": "২০২৬-০৪৫৭",
		"":          "",
		"ক":         "ক",
	}
	for in, want := range cases {
		if got := BN(in); got != want {
			t.Errorf("BN(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBNInt(t *testing.T) {
	cases := map[int]string{
		0: "০", 5: "৫", 12: "১২", 100: "১০০", 2026: "২০২৬", -7: "-৭",
	}
	for in, want := range cases {
		if got := BNInt(in); got != want {
			t.Errorf("BNInt(%d) = %q, want %q", in, got, want)
		}
	}
}

// Money is integer poisha end to end, so the decimal point is introduced here
// by integer division and never by a float conversion.
func TestTaka(t *testing.T) {
	cases := map[int64]string{
		0:       "৳ ০.০০",
		50:      "৳ ০.৫০",
		100:     "৳ ১.০০",
		125050:  "৳ ১,২৫০.৫০",
		999:     "৳ ৯.৯৯",
		-125050: "-৳ ১,২৫০.৫০",
		// South Asian grouping: last three digits, then pairs. 1234567 poisha
		// is ৳12,345.67 — grouped "১২,৩৪৫", not the Western "12,345".
		1234567: "৳ ১২,৩৪৫.৬৭",
		// 10,00,000 taka — the grouping that differs most from Western style.
		100000000: "৳ ১০,০০,০০০.০০",
	}
	for in, want := range cases {
		if got := Taka(in); got != want {
			t.Errorf("Taka(%d) = %q, want %q", in, got, want)
		}
	}
}

// A rounding or float bug in fee display would show a guardian the wrong
// balance, so check the poisha boundary explicitly.
func TestTaka_PoishaBoundary(t *testing.T) {
	for _, tc := range []struct {
		poisha int64
		want   string
	}{
		{1, "৳ ০.০১"},
		{9, "৳ ০.০৯"},
		{10, "৳ ০.১০"},
		{99, "৳ ০.৯৯"},
		{101, "৳ ১.০১"},
	} {
		if got := Taka(tc.poisha); got != tc.want {
			t.Errorf("Taka(%d) = %q, want %q", tc.poisha, got, tc.want)
		}
	}
}
