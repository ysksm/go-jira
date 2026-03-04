package cli

import "testing"

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		current, total, width int
		wantFilled            int
	}{
		{0, 100, 10, 0},
		{50, 100, 10, 5},
		{100, 100, 10, 10},
		{75, 100, 20, 15},
		{0, 0, 10, 0}, // zero total
	}

	for _, tt := range tests {
		bar := renderProgressBar(tt.current, tt.total, tt.width)
		// Check bar starts with [ and ends with ]
		if bar[0] != '[' || bar[len(bar)-1] != ']' {
			t.Errorf("bar should be wrapped in brackets: %s", bar)
		}
		// Check total width
		if len(bar) != tt.width+2 { // +2 for brackets
			t.Errorf("expected width %d, got %d: %s", tt.width+2, len(bar), bar)
		}
	}
}

func TestPhaseNumber(t *testing.T) {
	tests := []struct {
		phase string
		want  int
	}{
		{"fetch_issues", 1},
		{"sync_metadata", 2},
		{"generate_snapshots", 3},
		{"verify_integrity", 4},
		{"unknown", 0},
	}

	for _, tt := range tests {
		got := phaseNumber(tt.phase)
		if got != tt.want {
			t.Errorf("phaseNumber(%s): got %d, want %d", tt.phase, got, tt.want)
		}
	}
}

func TestExtractProjectKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PROJ-123", "PROJ"},
		{"MY-PROJECT-1", "MY"},
		{"X-1", "X"},
		{"NOHYPHEN", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractProjectKey(tt.input)
		if got != tt.want {
			t.Errorf("extractProjectKey(%s): got %s, want %s", tt.input, got, tt.want)
		}
	}
}
