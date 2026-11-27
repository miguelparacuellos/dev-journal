package store

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	t, _ := time.ParseInLocation(DateLayout, s, time.Local)
	return t
}

func TestDocRoundTrip(t *testing.T) {
	src := "# Daily — 2026-11-27\n\n## Yesterday\n- DEV-1 did it\n\n## Today\n- [ ] open\n- [x] done\n\n## Blockers\nfree text line\n"
	if got := ParseDoc(src).Render(); got != src {
		t.Fatalf("round trip mismatch:\n%s\n---\n%s", got, src)
	}
}

func TestNewDailyCarriesOverFromPreviousDaily(t *testing.T) {
	st := New(t.TempDir())
	prev := ParseDoc("# Daily\n## Yesterday\n- old\n## Today\n- [x] shipped DEV-1\n- [ ] still open\n## Blockers\n- waiting on infra\n")
	if err := SaveDoc(st.DailyPath(date("2026-11-21")), prev); err != nil {
		t.Fatal(err)
	}

	doc, err := st.LoadDaily(date("2026-11-24"))
	if err != nil {
		t.Fatal(err)
	}
	if y := doc.Section(SecYesterday).Items; len(y) != 1 || y[0].Text != "shipped DEV-1" || y[0].Checkbox {
		t.Errorf("yesterday = %+v", y)
	}
	if td := doc.Section(SecToday).Items; len(td) != 1 || td[0].Text != "still open" || td[0].Done {
		t.Errorf("today = %+v", td)
	}
	if b := doc.Section(SecBlockers).Items; len(b) != 1 {
		t.Errorf("blockers = %+v", b)
	}
}

func TestOpenActionsAndWins(t *testing.T) {
	st := New(t.TempDir())
	oct := ParseDoc("## Actions\n- [ ] write ADR\n- [x] pair with QA\n")
	_ = SaveDoc(st.OneOnOnePath(date("2026-10-29")), oct)
	_ = SaveDoc(st.WinsPath(), ParseDoc("## Wins\n- 2026-10-01 · old\n- 2026-11-20 · shipped refunds\n"))

	acts := st.OpenActionsBefore(date("2026-12-03"))
	if len(acts) != 1 || acts[0].Text != "write ADR" {
		t.Errorf("actions = %+v", acts)
	}
	if w := st.WinsSince(date("2026-10-29")); len(w) != 1 || w[0] != "shipped refunds" {
		t.Errorf("wins = %+v", w)
	}
}

func TestBacklogRoundTripAndQuickAdd(t *testing.T) {
	it := ParseQuickBacklog("refactor: Extract pricing rules #rates-service !m")
	want := BacklogItem{Title: "Extract pricing rules", Type: "refactor", Repo: "rates-service", Effort: "M", Status: "open"}
	if it != want {
		t.Fatalf("quick = %+v", it)
	}
	it.Notes = "Rules live in the controller.\n\nSecond paragraph."
	items := []BacklogItem{it, {Title: "Plain idea", Type: "idea", Status: "started"}}
	got := ParseBacklog(RenderBacklog(items))
	if len(got) != 2 || got[0] != items[0] || got[1] != items[1] {
		t.Fatalf("round trip = %+v", got)
	}
}
