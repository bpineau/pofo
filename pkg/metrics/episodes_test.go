package metrics

import "testing"

func TestDrawdownEpisodesTwo(t *testing.T) {
	// 100,120,90,120,100,130: two completed drawdowns from the 120 peak.
	eps := DrawdownEpisodes(days(6), []float64{100, 120, 90, 120, 100, 130})
	if len(eps) != 2 {
		t.Fatalf("episodes = %d, want 2", len(eps))
	}
	near(t, "depth1", eps[0].Depth, 90.0/120-1, 1e-12)
	near(t, "depth2", eps[1].Depth, 100.0/120-1, 1e-12)
	if eps[0].Ongoing || eps[1].Ongoing {
		t.Error("both episodes should have recovered")
	}
	if eps[0].DrawdownDays != 1 || eps[0].RecoveryDays != 1 {
		t.Errorf("episode1 timing = %d/%d d", eps[0].DrawdownDays, eps[0].RecoveryDays)
	}
}

func TestDrawdownEpisodesOngoing(t *testing.T) {
	eps := DrawdownEpisodes(days(3), []float64{100, 120, 90})
	if len(eps) != 1 || !eps[0].Ongoing {
		t.Fatalf("want one ongoing episode, got %+v", eps)
	}
	near(t, "depth", eps[0].Depth, 90.0/120-1, 1e-12)
}

func TestDrawdownEpisodesNone(t *testing.T) {
	if eps := DrawdownEpisodes(days(4), []float64{100, 101, 102, 103}); len(eps) != 0 {
		t.Errorf("monotone series should have no drawdown episodes, got %d", len(eps))
	}
}
