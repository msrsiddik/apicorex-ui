package components

import "testing"

// A merged sidebar's headings are the only thing separating two products'
// entries, and getting them wrong is quiet: the menu still renders.

func TestNavSectionsLeavesOneProductUnlabelled(t *testing.T) {
	// A heading over the only group in the menu names something the user
	// cannot be confused about, and costs a line in a sidebar with none spare.
	got := navSections([]NavItem{
		{Key: "a", Group: "schoolyze", GroupLabel: "School"},
		{Key: "b", Group: "schoolyze", GroupLabel: "School"},
	})
	if len(got) != 1 || got[0].Label != "" || len(got[0].Items) != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestNavSectionsSplitsTwoProducts(t *testing.T) {
	got := navSections([]NavItem{
		{Key: "dashboard", Group: "schoolyze", GroupLabel: "School"},
		{Key: "students", Group: "schoolyze", GroupLabel: "School"},
		{Key: "journal", Group: "accounting", GroupLabel: "Accounting"},
	})
	if len(got) != 2 {
		t.Fatalf("want 2 sections, got %d: %+v", len(got), got)
	}
	if got[0].Label != "School" || len(got[0].Items) != 2 {
		t.Fatalf("first section: %+v", got[0])
	}
	if got[1].Label != "Accounting" || len(got[1].Items) != 1 {
		t.Fatalf("second section: %+v", got[1])
	}
}

func TestNavSectionsKeepsTheOrderItWasGiven(t *testing.T) {
	// Which product leads is decided in Core, where the whole suite is visible.
	// Re-sorting here would quietly overrule it.
	got := navSections([]NavItem{
		{Key: "journal", Group: "accounting", GroupLabel: "Accounting"},
		{Key: "dashboard", Group: "schoolyze", GroupLabel: "School"},
	})
	if got[0].Label != "Accounting" {
		t.Fatalf("got %+v", got)
	}
}

func TestNavSectionsHandlesAnEmptyMenu(t *testing.T) {
	if got := navSections(nil); len(got) != 1 || len(got[0].Items) != 0 {
		t.Fatalf("got %+v", got)
	}
}
