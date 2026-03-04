package cli

import (
	"fmt"
	"strings"

	"github.com/ysksm/go-jira/core/domain/models"
)

// progressDisplay handles terminal progress output.
type progressDisplay struct {
	lastProject string
	lastPhase   string
}

func newProgressDisplay() *progressDisplay {
	return &progressDisplay{}
}

func (p *progressDisplay) update(progress models.SyncProgress) {
	if progress.ProjectKey != p.lastProject {
		if p.lastProject != "" {
			fmt.Println()
		}
		fmt.Printf("  Syncing project: %s\n", progress.ProjectKey)
		p.lastProject = progress.ProjectKey
		p.lastPhase = ""
	}

	phaseName := phaseDisplayName(progress.Phase)
	phaseNum := phaseNumber(progress.Phase)

	if progress.Phase != p.lastPhase {
		p.lastPhase = progress.Phase
		fmt.Printf("    Phase %d/4: %s\n", phaseNum, phaseName)
	}

	if progress.Total > 0 && progress.Phase != models.PhaseSyncMetadata {
		bar := renderProgressBar(progress.Current, progress.Total, 40)
		fmt.Printf("\r      %s %d/%d (%d%%)",
			bar, progress.Current, progress.Total, progress.Current*100/progress.Total)
		if progress.Current >= progress.Total {
			fmt.Println()
		}
	} else {
		fmt.Printf("      %s\n", progress.Message)
	}
}

func phaseDisplayName(phase string) string {
	switch phase {
	case models.PhaseFetchIssues:
		return "Fetching issues..."
	case models.PhaseSyncMetadata:
		return "Syncing metadata..."
	case models.PhaseGenerateSnapshots:
		return "Generating snapshots..."
	case models.PhaseVerifyIntegrity:
		return "Verifying integrity..."
	default:
		return phase
	}
}

func phaseNumber(phase string) int {
	switch phase {
	case models.PhaseFetchIssues:
		return 1
	case models.PhaseSyncMetadata:
		return 2
	case models.PhaseGenerateSnapshots:
		return 3
	case models.PhaseVerifyIntegrity:
		return 4
	default:
		return 0
	}
}

func renderProgressBar(current, total, width int) string {
	if total == 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}
	filled := current * width / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"
}
