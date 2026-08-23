// Package bn formats numbers, money and month names the way a Bangladeshi
// reader expects them.
//
// It lives in the ui module, and both the screens and the PDF documents go
// through it. That is the whole point: a receipt that disagreed with the page
// it was printed from would be two formatters drifting apart, and this is the
// one copy.
package bn

import "strings"

// Bangla digits.
//
// A printed Bangla document uses Bangla numerals throughout — marks, roll
// numbers, dates, taka amounts. Mixing Latin digits into an otherwise Bangla
// marksheet reads as a bug to the parent holding it, so conversion happens at
// render time rather than being scattered through the callers.
//
// Only for display. Never convert a value that will be parsed, compared or
// stored: these are presentation glyphs, not a number format.

var banglaDigits = [10]rune{'০', '১', '২', '৩', '৪', '৫', '৬', '৭', '৮', '৯'}

// BN converts the ASCII digits in s to Bangla, leaving everything else — commas,
// decimal points, the taka sign, letters — untouched.
//
// So "1,250.50" becomes "১,২৫০.৫০" and a grade like "A+" is unchanged, which is
// correct: Bangladeshi report cards write grades in Latin letters.
func BN(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(banglaDigits[r-'0'])
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// BNInt renders an integer in Bangla digits.
func BNInt(n int) string {
	if n == 0 {
		return string(banglaDigits[0])
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var out []rune
	for n > 0 {
		out = append([]rune{banglaDigits[n%10]}, out...)
		n /= 10
	}
	if neg {
		out = append([]rune{'-'}, out...)
	}
	return string(out)
}

// Taka formats an integer poisha amount as Bangla currency: 125050 → "৳ ১,২৫০.৫০".
//
// Money is stored as integer poisha everywhere (docs/architecture.md §8.4), so
// this is the one place the decimal point is introduced — by integer division,
// never by converting to a float.
func Taka(poisha int64) string {
	sign, grouped, frac := takaParts(poisha)
	return sign + "৳ " + BN(grouped) + "." + BN(frac)
}

// TakaEN formats the same amount for an English printout: 125050 → "৳ 1,250.50".
//
// Grouping stays South Asian (last three digits, then pairs) even in English:
// that is how Bangladeshi institutions write money regardless of language, and
// Western grouping on an English-language receipt printed in Bangladesh would
// look wrong to whoever reconciles it. Only the digits stop converting to
// Bangla numerals.
func TakaEN(poisha int64) string {
	sign, grouped, frac := takaParts(poisha)
	return sign + "৳ " + grouped + "." + frac
}

// takaParts does the shared arithmetic Taka and TakaEN both format from: the
// sign, the grouped whole-taka digits, and the two-digit poisha remainder —
// all still in plain Latin digits, so a caller converts them or not.
func takaParts(poisha int64) (sign, grouped, frac string) {
	neg := poisha < 0
	if neg {
		poisha = -poisha
	}
	whole, fracN := poisha/100, poisha%100
	if neg {
		sign = "-"
	}
	return sign, groupBD(whole), pad2(fracN)
}

// groupBD applies the South Asian digit grouping used in Bangladesh: the last
// three digits, then pairs (12,34,567 — not 12,345,67). Western grouping on a
// fee receipt looks wrong to the person reconciling it.
func groupBD(n int64) string {
	s := itoa(n)
	if len(s) <= 3 {
		return s
	}
	head, tail := s[:len(s)-3], s[len(s)-3:]

	var parts []string
	for len(head) > 2 {
		parts = append([]string{head[len(head)-2:]}, parts...)
		head = head[:len(head)-2]
	}
	if head != "" {
		parts = append([]string{head}, parts...)
	}
	return strings.Join(parts, ",") + "," + tail
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var out []byte
	for n > 0 {
		out = append([]byte{byte('0' + n%10)}, out...)
		n /= 10
	}
	return string(out)
}

func pad2(n int64) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// ── Bangla month names ───────────────────────────────────────────────────────
//
// The Gregorian months as Bangladeshi institutions write them. Not the Bengali
// calendar (বৈশাখ, জ্যৈষ্ঠ …): school terms, fee months and exam schedules all
// run on the Gregorian year, and printing a Bengali-calendar month on a fee
// receipt would confuse the guardian reading it.

var banglaMonths = [12]string{
	"জানুয়ারি", "ফেব্রুয়ারি", "মার্চ", "এপ্রিল", "মে", "জুন",
	"জুলাই", "আগস্ট", "সেপ্টেম্বর", "অক্টোবর", "নভেম্বর", "ডিসেম্বর",
}

// MonthBN returns a Gregorian month's Bangla name (1 = January). Out-of-range
// input yields "" rather than panicking — a document with a blank month is
// better than a 500 while a clerk is printing a receipt.
func MonthBN(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	return banglaMonths[month-1]
}

var englishMonths = [12]string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// MonthEN returns a Gregorian month's English name (1 = January), the
// counterpart to MonthBN for a document printed in English. Same
// out-of-range behaviour: "" rather than a panic.
func MonthEN(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	return englishMonths[month-1]
}
