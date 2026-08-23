package i18n

import (
	"context"
	"strconv"
	"time"

	"github.com/msrsiddik/apicorex-ui/bn"
)

// Locale-aware number, money and date formatting.
//
// These delegate to ui/bn rather than reimplementing the rules, so a figure on
// screen and the same figure on a printed marksheet are formatted by one piece
// of code. A second formatter would drift, and the drift would show
// up as a receipt that disagrees with the screen it was printed from.

// N formats an integer in the request's locale: Bangla numerals under Bangla,
// Latin under English.
func N(ctx context.Context, n int) string {
	if FromContext(ctx).IsBangla() {
		return bn.BNInt(n)
	}
	return strconv.Itoa(n)
}

// N64 formats an int64 (a count, not money — see Money).
func N64(ctx context.Context, n int64) string {
	if FromContext(ctx).IsBangla() {
		return bn.BNInt(int(n))
	}
	return strconv.FormatInt(n, 10)
}

// Digits converts the ASCII digits inside an arbitrary string, leaving
// everything else alone. For values that are numeric but not numbers — an
// admission number like "2026-0457", a phone number.
func Digits(ctx context.Context, s string) string {
	if FromContext(ctx).IsBangla() {
		return bn.BN(s)
	}
	return s
}

// Money formats integer poisha as currency.
//
// Bangla gets the taka sign and South Asian digit grouping (৳ ১০,০০,০০০.০০ —
// last three digits, then pairs), which is what whoever reconciles the ledger
// expects. English keeps Latin digits but the same grouping: the grouping is a
// property of the region, not the script.
func Money(ctx context.Context, poisha int64) string {
	formatted := bn.Taka(poisha)
	if FromContext(ctx).IsBangla() {
		return formatted
	}
	// bn.Taka renders Bangla numerals; map them back for the English locale
	// rather than duplicating the grouping logic.
	return latinDigits(formatted)
}

// Date formats a calendar date.
//
// Rendered in Asia/Dhaka, never UTC: a payment taken at 9pm Dhaka time must
// not display tomorrow's date.
func Date(ctx context.Context, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	local := t.In(dhaka)
	if FromContext(ctx).IsBangla() {
		return bn.BNInt(local.Day()) + " " + bn.MonthBN(int(local.Month())) + " " + bn.BNInt(local.Year())
	}
	return local.Format("2 January 2006")
}

// MonthName returns a Gregorian month's name (1 = January) in the request's
// locale. Gregorian, not the Bengali calendar: school terms, fee months and
// exam schedules all run on the Gregorian year.
func MonthName(ctx context.Context, month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	if FromContext(ctx).IsBangla() {
		return bn.MonthBN(month)
	}
	return time.Month(month).String()
}

// DateInput formats a date for an <input type="date">, which requires
// ISO 8601 with Latin digits regardless of the interface language.
func DateInput(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(dhaka).Format("2006-01-02")
}

// dhaka is the display timezone. Loaded once; falls back to a fixed +06:00 when
// the host has no zoneinfo database, which is common in a scratch container.
var dhaka = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Dhaka"); err == nil {
		return loc
	}
	return time.FixedZone("BST", 6*60*60)
}()

var banglaToLatin = map[rune]rune{
	'০': '0', '১': '1', '২': '2', '৩': '3', '৪': '4',
	'৫': '5', '৬': '6', '৭': '7', '৮': '8', '৯': '9',
}

func latinDigits(s string) string {
	out := []rune(s)
	for i, r := range out {
		if l, ok := banglaToLatin[r]; ok {
			out[i] = l
		}
	}
	return string(out)
}
