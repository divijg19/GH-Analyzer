package intelligence

import (
	"fmt"
	"time"

	"github.com/divijg19/Atlas/internal/evidence"
	"github.com/divijg19/Atlas/internal/profile"
	"github.com/divijg19/Atlas/internal/repositoryintelligence"
)

// PortfolioIntelligence is the overall characterization of the candidate's body
// of work. Aggregated from per-repository Repository Intelligence plus account
// metadata.
type PortfolioIntelligence struct {
	dimensionBase
	Level           Level
	RepositoryCount int
	AtRiskRepos     int
	AccountAgeDays  int
	Confidence      Confidence
}

func buildPortfolio(repos []repositoryintelligence.RepositoryIntelligence, meta *profile.UserMetadata, referenceTime time.Time) PortfolioIntelligence {
	dim := PortfolioIntelligence{}
	dim.name = "portfolio"

	total := len(repos)
	dim.RepositoryCount = total

	var atRisk int
	for _, r := range repos {
		if r.Risk.Level == repositoryintelligence.LevelAtRisk {
			atRisk++
		}
	}
	dim.AtRiskRepos = atRisk

	if meta != nil && !meta.CreatedAt.IsZero() {
		dim.AccountAgeDays = int(referenceTime.Sub(meta.CreatedAt).Hours() / 24)
		if dim.AccountAgeDays < 0 {
			dim.AccountAgeDays = 0
		}
	}

	switch {
	case total == 0:
		dim.Level = LevelUnknown
		dim.Confidence = ConfidenceLow
	case total >= 10:
		dim.Level = LevelBroad
	case total >= 3:
		dim.Level = LevelModerate
	default:
		dim.Level = LevelNarrow
	}
	if dim.Level != LevelUnknown {
		dim.Confidence = ConfidenceHigh
	}
	dim.evidence = []evidence.EvidenceGroup{
		groupFrom(repos, "portfolio", []string{"risk"},
			factItem("repositories", total),
			factItem("at-risk repositories", atRisk),
			factItem("account age (days)", dim.AccountAgeDays),
		),
	}
	dim.summary = fmt.Sprintf("Portfolio is %s (%d repositories; %d at risk).", dim.Level, total, atRisk)
	return dim
}
