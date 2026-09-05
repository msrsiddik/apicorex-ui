package components

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// Two things every list control has to agree on, both of which fail silently:
// what it sends with the request, and what happens when two requests overlap.
// The list still renders either way — it just answers the wrong question, or
// answers an older one last.

func renderToolbar(t *testing.T, p TableToolbar) string {
	t.Helper()
	var sb strings.Builder
	if err := Toolbar(p).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func renderTable(t *testing.T, p TableProps) string {
	t.Helper()
	var sb strings.Builder
	if err := Table(p, templ.NopComponent).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func sampleToolbar() TableToolbar {
	return TableToolbar{
		Endpoint:   "/students/rows",
		Target:     "#list",
		SearchName: "q",
		Filters: []SelectFilter{
			{Name: "class_id", Label: "Class", AllLabel: "All"},
			{Name: "status", Label: "Status", AllLabel: "All"},
		},
	}
}

// TestToolbarControlsIncludeTheWholeToolbar is the regression guard for the
// bug this test file was written for: hx-include said "closest .flex", and the
// search input's nearest flex box is the one holding its own icon, not the
// toolbar. Typing a name therefore sent q without class_id — the filter
// stopped applying while its dropdown still showed the class selected.
func TestToolbarControlsIncludeTheWholeToolbar(t *testing.T) {
	tb := sampleToolbar()
	html := renderToolbar(t, tb)

	if !strings.Contains(html, "list-toolbar") {
		t.Fatalf("toolbar root lost its marker class, so hx-include names nothing: %s", html)
	}
	// The search box and every filter, all naming the same ancestor.
	if got, want := strings.Count(html, `hx-include="closest .list-toolbar"`), 1+len(tb.Filters); got != want {
		t.Fatalf("hx-include on %d controls, want %d: %s", got, want, html)
	}
	if strings.Contains(html, "closest .flex") {
		t.Fatalf("a control still includes the nearest flex box rather than the toolbar: %s", html)
	}
}

// TestListControlsCancelEachOther covers the other half: a debounce spaces
// requests out but does not order the responses, so a slow query started three
// letters ago can still land after a fast one and overwrite it.
func TestListControlsCancelEachOther(t *testing.T) {
	tb := sampleToolbar()
	html := renderToolbar(t, tb)
	if got, want := strings.Count(html, `hx-sync="#list:replace"`), 1+len(tb.Filters); got != want {
		t.Fatalf("hx-sync on %d toolbar controls, want %d: %s", got, want, html)
	}

	// Sort headers and page buttons swap the same list, so they share the queue.
	p := TableProps{
		Columns:  []Column{{Label: "Roll", Sort: "roll"}, {Label: "Name", Sort: "name"}},
		Endpoint: "/students/rows",
		Target:   "#list",
		Page:     PageInfo{Limit: 50, Offset: 50, Total: 200},
	}
	html = renderTable(t, p)
	// Two sortable headers, plus previous and next.
	if got, want := strings.Count(html, `hx-sync="#list:replace"`), 4; got != want {
		t.Fatalf("hx-sync on %d table controls, want %d: %s", got, want, html)
	}
}

// TestSearchKeepsTheSort covers the third way a list can quietly answer the
// wrong question: sort by roll, then type a name, and the sort is gone because
// the toolbar never knew about it.
func TestSearchKeepsTheSort(t *testing.T) {
	tb := sampleToolbar()
	tb.SortKey, tb.SortDesc = "roll", true
	html := renderToolbar(t, tb)

	// Inside the toolbar, so hx-include picks them up, and present even when a
	// search returns nothing and the table is replaced by an empty state.
	if !strings.Contains(html, `id="list-sort" name="sort" value="roll"`) {
		t.Fatalf("toolbar does not carry the sort key: %s", html)
	}
	if !strings.Contains(html, `id="list-desc" name="desc" value="true"`) {
		t.Fatalf("toolbar does not carry the sort direction: %s", html)
	}
}

// TestSortHeaderUpdatesTheToolbar is the other half of it: the toolbar renders
// once and is never swapped, so a sort click has to write the new ordering
// into it or the hidden fields go stale on the first click.
func TestSortHeaderUpdatesTheToolbar(t *testing.T) {
	p := TableProps{
		Columns:  []Column{{Label: "Roll", Sort: "roll"}, {Label: "Name", Sort: "name"}},
		Endpoint: "/students/rows",
		Target:   "#list",
		SortKey:  "roll", // ascending, so clicking Roll again reverses
		Page:     PageInfo{Limit: 50, Total: 200},
	}
	html := renderTable(t, p)

	// templ escapes the quotes inside the attribute; the browser unescapes
	// them before the handler is parsed.
	for _, want := range []string{
		`s.value=&#34;roll&#34;}`,  // the active column...
		`d.value=&#34;true&#34;}`,  // ...reverses
		`s.value=&#34;name&#34;}`,  // a different column...
		`d.value=&#34;false&#34;}`, // ...starts ascending
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("sort header does not write %s into the toolbar: %s", want, html)
		}
	}
	if got := strings.Count(html, "list-sort"); got != len(p.Columns) {
		t.Fatalf("%d headers update the toolbar, want %d: %s", got, len(p.Columns), html)
	}
}
