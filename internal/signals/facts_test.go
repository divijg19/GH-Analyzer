package signals

import (
	"testing"
	"time"
)

// minDepthRepoSize mirrors facts.minDepthRepoSize for use as a test fixture
// boundary. The canonical constant lives in internal/facts.
const minDepthRepoSize = 50

func TestFromReposEmpty(t *testing.T) {
	f := FromRepos(nil, refTime)
	assertZeroFacts(t, f)

	f = FromRepos([]RepositoryVestige{}, refTime)
	assertZeroFacts(t, f)
}

func TestFromReposCountsTotalRepos(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(30)},
	}

	f := FromRepos(repos, refTime)

	if f.TotalRepos != 3 {
		t.Fatalf("expected TotalRepos 3, got %d", f.TotalRepos)
	}
}

func TestFromReposCountsOriginalAndForkRepos(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(30)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(40)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(50)},
	}

	f := FromRepos(repos, refTime)

	if f.OriginalRepos != 3 {
		t.Fatalf("expected OriginalRepos 3, got %d", f.OriginalRepos)
	}
	if f.ForkRepos != 2 {
		t.Fatalf("expected ForkRepos 2, got %d", f.ForkRepos)
	}
	if f.TotalRepos != f.OriginalRepos+f.ForkRepos {
		t.Fatalf("TotalRepos (%d) != OriginalRepos (%d) + ForkRepos (%d)", f.TotalRepos, f.OriginalRepos, f.ForkRepos)
	}
}

func TestFromReposRecentReposWithinWindow(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(89)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(91)},
	}

	f := FromRepos(repos, refTime)

	if f.RecentRepos != 2 {
		t.Fatalf("expected RecentRepos 2 (10 and 89 days), got %d", f.RecentRepos)
	}
}

func TestFromReposForksNotCountedAsRecent(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
	}

	f := FromRepos(repos, refTime)

	if f.RecentRepos != 1 {
		t.Fatalf("expected RecentRepos 1 (fork excluded), got %d", f.RecentRepos)
	}
}

func TestFromReposDeepReposAboveThreshold(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: minDepthRepoSize, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: minDepthRepoSize + 1, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: minDepthRepoSize - 1, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.DeepRepos != 2 {
		t.Fatalf("expected DeepRepos 2 (exactly at threshold and above), got %d", f.DeepRepos)
	}
}

func TestFromReposForksNotCountedAsDeep(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 1000, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.DeepRepos != 1 {
		t.Fatalf("expected DeepRepos 1 (fork excluded despite size), got %d", f.DeepRepos)
	}
}

func TestFromReposValidReposExcludesZeroSize(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.ValidRepos != 2 {
		t.Fatalf("expected ValidRepos 2 (size > 0), got %d", f.ValidRepos)
	}
}

func TestFromReposValidOriginalReposExcludesForksAndZeroSize(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 50, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.ValidOriginalRepos != 2 {
		t.Fatalf("expected ValidOriginalRepos 2 (non-fork, size > 0), got %d", f.ValidOriginalRepos)
	}
}

func TestFromReposLargestRepoSize(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 999, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 500, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.LargestRepoSize != 999 {
		t.Fatalf("expected LargestRepoSize 999, got %d", f.LargestRepoSize)
	}
}

func TestFromReposLargestRepoSizeWithAllZeros(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.LargestRepoSize != 0 {
		t.Fatalf("expected LargestRepoSize 0, got %d", f.LargestRepoSize)
	}
}

func TestFromReposLatestActivity(t *testing.T) {
	now := time.Now()
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: now.Add(-48 * time.Hour)},
		{Fork: false, Size: 100, UpdatedAt: now.Add(-72 * time.Hour)},
		{Fork: false, Size: 100, UpdatedAt: now.Add(-24 * time.Hour)},
	}

	f := FromRepos(repos, refTime)

	expected := now.Add(-24 * time.Hour)
	if !f.LatestActivity.Equal(expected) {
		t.Fatalf("expected LatestActivity %v, got %v", expected, f.LatestActivity)
	}
}

func TestFromReposLatestActivityAcrossForksAndOriginals(t *testing.T) {
	now := time.Now()
	repos := []RepositoryVestige{
		{Fork: true, Size: 100, UpdatedAt: now.Add(-1 * time.Hour)},
		{Fork: false, Size: 100, UpdatedAt: now.Add(-48 * time.Hour)},
	}

	f := FromRepos(repos, refTime)

	if !f.LatestActivity.Equal(now.Add(-1 * time.Hour)) {
		t.Fatalf("expected LatestActivity to be fork's recent update, got %v", f.LatestActivity)
	}
}

func TestFromReposAllOriginalAllRecentAllDeepAllValid(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 200, UpdatedAt: daysAgo(5)},
		{Fork: false, Size: 150, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(15)},
	}

	f := FromRepos(repos, refTime)

	if f.TotalRepos != 3 {
		t.Fatalf("TotalRepos: 3 != %d", f.TotalRepos)
	}
	if f.OriginalRepos != 3 {
		t.Fatalf("OriginalRepos: 3 != %d", f.OriginalRepos)
	}
	if f.ForkRepos != 0 {
		t.Fatalf("ForkRepos: 0 != %d", f.ForkRepos)
	}
	if f.RecentRepos != 3 {
		t.Fatalf("RecentRepos: 3 != %d", f.RecentRepos)
	}
	if f.DeepRepos != 3 {
		t.Fatalf("DeepRepos: 3 != %d", f.DeepRepos)
	}
	if f.ValidRepos != 3 {
		t.Fatalf("ValidRepos: 3 != %d", f.ValidRepos)
	}
	if f.ValidOriginalRepos != 3 {
		t.Fatalf("ValidOriginalRepos: 3 != %d", f.ValidOriginalRepos)
	}
}

func TestFromReposAllForks(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
		{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.OriginalRepos != 0 {
		t.Fatalf("OriginalRepos: 0 != %d", f.OriginalRepos)
	}
	if f.ForkRepos != 2 {
		t.Fatalf("ForkRepos: 2 != %d", f.ForkRepos)
	}
	if f.RecentRepos != 0 {
		t.Fatalf("RecentRepos: 0 != %d", f.RecentRepos)
	}
	if f.DeepRepos != 0 {
		t.Fatalf("DeepRepos: 0 != %d", f.DeepRepos)
	}
	if f.ValidOriginalRepos != 0 {
		t.Fatalf("ValidOriginalRepos: 0 != %d", f.ValidOriginalRepos)
	}
	if f.ValidRepos != 2 {
		t.Fatalf("ValidRepos: 2 != %d", f.ValidRepos)
	}
}

func TestFromReposSingleRepo(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 75, UpdatedAt: daysAgo(30)},
	}

	f := FromRepos(repos, refTime)

	if f.TotalRepos != 1 {
		t.Fatalf("TotalRepos: 1 != %d", f.TotalRepos)
	}
	if f.OriginalRepos != 1 {
		t.Fatalf("OriginalRepos: 1 != %d", f.OriginalRepos)
	}
	if f.ForkRepos != 0 {
		t.Fatalf("ForkRepos: 0 != %d", f.ForkRepos)
	}
	if f.RecentRepos != 1 {
		t.Fatalf("RecentRepos: 1 != %d", f.RecentRepos)
	}
	if f.DeepRepos != 1 {
		t.Fatalf("DeepRepos: 1 != %d", f.DeepRepos)
	}
	if f.ValidRepos != 1 {
		t.Fatalf("ValidRepos: 1 != %d", f.ValidRepos)
	}
	if f.ValidOriginalRepos != 1 {
		t.Fatalf("ValidOriginalRepos: 1 != %d", f.ValidOriginalRepos)
	}
	if f.LargestRepoSize != 75 {
		t.Fatalf("LargestRepoSize: 75 != %d", f.LargestRepoSize)
	}
}

func TestFromReposMetadataFacts(t *testing.T) {
	createdOld := time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)
	createdNew := time.Date(2023, 6, 7, 0, 0, 0, 0, time.UTC)

	repos := []RepositoryVestige{
		{
			Fork: false, Size: 100, UpdatedAt: daysAgo(10),
			Visibility: "public", License: "mit", Topics: []string{"go", "cli"},
			Stars: 50, Forks: 5, Watchers: 7, OpenIssues: 3,
			CreatedAt: createdOld,
		},
		{
			Fork: false, Size: 200, UpdatedAt: daysAgo(20),
			Visibility: "private", Archived: true, Template: true,
			License: "", Topics: nil,
			Stars: 0, Forks: 0, Watchers: 0, OpenIssues: 0,
			CreatedAt: createdNew,
		},
		{
			Fork: true, Size: 10, UpdatedAt: daysAgo(30),
			Visibility: "public", License: "apache-2.0", Topics: []string{"lib"},
			Stars: 1, Forks: 0, Watchers: 2, OpenIssues: 1,
			CreatedAt: createdNew,
		},
	}

	f := FromRepos(repos, refTime)

	if f.ArchivedRepos != 1 {
		t.Fatalf("ArchivedRepos: 1 != %d", f.ArchivedRepos)
	}
	if f.TemplateRepos != 1 {
		t.Fatalf("TemplateRepos: 1 != %d", f.TemplateRepos)
	}
	if f.PublicRepos != 2 {
		t.Fatalf("PublicRepos: 2 != %d", f.PublicRepos)
	}
	if f.PrivateRepos != 1 {
		t.Fatalf("PrivateRepos: 1 != %d", f.PrivateRepos)
	}
	if f.LicensedRepos != 2 {
		t.Fatalf("LicensedRepos: 2 != %d", f.LicensedRepos)
	}
	if f.TotalStars != 51 {
		t.Fatalf("TotalStars: 51 != %d", f.TotalStars)
	}
	if f.TotalForks != 5 {
		t.Fatalf("TotalForks: 5 != %d", f.TotalForks)
	}
	if f.TotalWatchers != 9 {
		t.Fatalf("TotalWatchers: 9 != %d", f.TotalWatchers)
	}
	if f.TotalOpenIssues != 4 {
		t.Fatalf("TotalOpenIssues: 4 != %d", f.TotalOpenIssues)
	}
	if f.TotalTopics != 3 {
		t.Fatalf("TotalTopics: 3 != %d", f.TotalTopics)
	}
	if !f.OldestCreated.Equal(createdOld) {
		t.Fatalf("OldestCreated: %v != %v", f.OldestCreated, createdOld)
	}
	if !f.NewestCreated.Equal(createdNew) {
		t.Fatalf("NewestCreated: %v != %v", f.NewestCreated, createdNew)
	}
}

func TestFromReposNoOriginalRepos(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: true, Size: 0, UpdatedAt: daysAgo(5)},
		{Fork: true, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos, refTime)

	if f.OriginalRepos != 0 {
		t.Fatalf("OriginalRepos: 0 != %d", f.OriginalRepos)
	}
	if f.ValidRepos != 0 {
		t.Fatalf("ValidRepos: 0 != %d", f.ValidRepos)
	}
	if f.ValidOriginalRepos != 0 {
		t.Fatalf("ValidOriginalRepos: 0 != %d", f.ValidOriginalRepos)
	}
}

func assertZeroFacts(t *testing.T, f RepositoryFacts) {
	t.Helper()

	if f.TotalRepos != 0 {
		t.Fatalf("expected TotalRepos 0, got %d", f.TotalRepos)
	}
	if f.OriginalRepos != 0 {
		t.Fatalf("expected OriginalRepos 0, got %d", f.OriginalRepos)
	}
	if f.ForkRepos != 0 {
		t.Fatalf("expected ForkRepos 0, got %d", f.ForkRepos)
	}
	if f.RecentRepos != 0 {
		t.Fatalf("expected RecentRepos 0, got %d", f.RecentRepos)
	}
	if f.DeepRepos != 0 {
		t.Fatalf("expected DeepRepos 0, got %d", f.DeepRepos)
	}
	if f.ValidRepos != 0 {
		t.Fatalf("expected ValidRepos 0, got %d", f.ValidRepos)
	}
	if f.ValidOriginalRepos != 0 {
		t.Fatalf("expected ValidOriginalRepos 0, got %d", f.ValidOriginalRepos)
	}
	if f.LargestRepoSize != 0 {
		t.Fatalf("expected LargestRepoSize 0, got %d", f.LargestRepoSize)
	}
	if !f.LatestActivity.IsZero() {
		t.Fatalf("expected LatestActivity zero, got %v", f.LatestActivity)
	}
}

func TestFromReposDerivedIntelligence(t *testing.T) {
	createdOld := time.Date(2015, 3, 4, 0, 0, 0, 0, time.UTC)
	createdNew := time.Date(2024, 8, 9, 0, 0, 0, 0, time.UTC)
	released := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)

	repos := []RepositoryVestige{
		{
			Fork: false, Size: 100, UpdatedAt: daysAgo(10), CreatedAt: createdOld,
			Stars: 40, License: "mit", Topics: []string{"go", "cli"},
			ReleaseCount: 2, LatestReleaseAt: released,
			PullRequestCount: 5, CollaboratorCount: 3,
			DefaultBranchProtected: true, DiscussionEnabled: true,
			LanguageDistribution: map[string]int64{"Go": 800, "Shell": 200},
		},
		{
			Fork: false, Size: 200, UpdatedAt: daysAgo(20), CreatedAt: createdNew,
			Stars: 10, License: "apache-2.0", Topics: []string{"rust"},
			ReleaseCount: 1, LatestReleaseAt: time.Time{},
			PullRequestCount: 1, CollaboratorCount: 1,
			DefaultBranchProtected: false, DiscussionEnabled: false,
			LanguageDistribution: map[string]int64{"Rust": 900, "Go": 100},
		},
		{
			Fork: true, Size: 10, UpdatedAt: daysAgo(30), CreatedAt: createdNew,
			Stars: 5, License: "", Topics: nil,
			ReleaseCount: 0, LatestReleaseAt: time.Time{},
			PullRequestCount: 0, CollaboratorCount: 0,
			DefaultBranchProtected: false, DiscussionEnabled: false,
			LanguageDistribution: map[string]int64{"JavaScript": 500},
		},
	}

	f := FromRepos(repos, refTime)

	// Star distribution
	if f.MaxRepoStars != 40 {
		t.Fatalf("MaxRepoStars: 40 != %d", f.MaxRepoStars)
	}
	if got := f.MeanRepoStars; !almostEqual(got, 55.0/3.0) {
		t.Fatalf("MeanRepoStars: %.4f != %.4f", got, 55.0/3.0)
	}

	// Technology breadth (union of Go, Shell, Rust, JavaScript = 4 distinct)
	if f.LanguageCount != 4 {
		t.Fatalf("LanguageCount: 4 != %d", f.LanguageCount)
	}
	want := []string{"Go", "Rust", "JavaScript", "Shell"}
	if len(f.RankedLanguages) != len(want) {
		t.Fatalf("RankedLanguages length: %d != %d (%v)", len(f.RankedLanguages), len(want), f.RankedLanguages)
	}
	for i, w := range want {
		if f.RankedLanguages[i] != w {
			t.Fatalf("RankedLanguages[%d]: %q != %q (full: %v)", i, f.RankedLanguages[i], w, f.RankedLanguages)
		}
	}

	// Release & maintenance aggregates
	if f.TotalReleases != 3 {
		t.Fatalf("TotalReleases: 3 != %d", f.TotalReleases)
	}
	if f.ReleasedRepos != 2 {
		t.Fatalf("ReleasedRepos: 2 != %d", f.ReleasedRepos)
	}
	if !f.LatestReleaseAt.Equal(released) {
		t.Fatalf("LatestReleaseAt: %v != %v", f.LatestReleaseAt, released)
	}
	if f.TotalPullRequests != 6 {
		t.Fatalf("TotalPullRequests: 6 != %d", f.TotalPullRequests)
	}
	if f.TotalCollaborators != 4 {
		t.Fatalf("TotalCollaborators: 4 != %d", f.TotalCollaborators)
	}
	if f.ProtectedBranchRepos != 1 {
		t.Fatalf("ProtectedBranchRepos: 1 != %d", f.ProtectedBranchRepos)
	}
	if f.DiscussionRepos != 1 {
		t.Fatalf("DiscussionRepos: 1 != %d", f.DiscussionRepos)
	}

	// Ratio facts (in [0, 1])
	if !almostEqual(f.ForkRatio, 1.0/3.0) {
		t.Fatalf("ForkRatio: %.4f != %.4f", f.ForkRatio, 1.0/3.0)
	}
	if !almostEqual(f.LicensedRatio, 2.0/3.0) {
		t.Fatalf("LicensedRatio: %.4f != %.4f", f.LicensedRatio, 2.0/3.0)
	}
	if f.ArchivedRatio != 0 {
		t.Fatalf("ArchivedRatio: 0 != %.4f", f.ArchivedRatio)
	}
	if f.ForkRatio < 0 || f.ForkRatio > 1 || f.LicensedRatio < 0 || f.LicensedRatio > 1 || f.ArchivedRatio < 0 || f.ArchivedRatio > 1 {
		t.Fatalf("ratio facts out of [0,1]: fork=%.4f licensed=%.4f archived=%.4f", f.ForkRatio, f.LicensedRatio, f.ArchivedRatio)
	}

	// Freshness & age
	if f.PortfolioAgeDays != int(refTime.Sub(createdOld).Hours()/24) {
		t.Fatalf("PortfolioAgeDays: %d != %d", f.PortfolioAgeDays, int(refTime.Sub(createdOld).Hours()/24))
	}
	if f.NewestRepoAgeDays != int(refTime.Sub(createdNew).Hours()/24) {
		t.Fatalf("NewestRepoAgeDays: %d != %d", f.NewestRepoAgeDays, int(refTime.Sub(createdNew).Hours()/24))
	}
	if f.DaysSinceLatestRelease != int(refTime.Sub(released).Hours()/24) {
		t.Fatalf("DaysSinceLatestRelease: %d != %d", f.DaysSinceLatestRelease, int(refTime.Sub(released).Hours()/24))
	}
	if !almostEqual(f.MeanRepoSize, 310.0/3.0) {
		t.Fatalf("MeanRepoSize: %.4f != %.4f", f.MeanRepoSize, 310.0/3.0)
	}
	if !almostEqual(f.TopicBreadth, 1.0) {
		t.Fatalf("TopicBreadth: %.4f != %.4f", f.TopicBreadth, 1.0)
	}
}

func TestFromReposDerivedIntelligenceEmpty(t *testing.T) {
	f := FromRepos(nil, refTime)

	if f.MaxRepoStars != 0 {
		t.Fatalf("MaxRepoStars: 0 != %d", f.MaxRepoStars)
	}
	if f.MeanRepoStars != 0 {
		t.Fatalf("MeanRepoStars: 0 != %.4f", f.MeanRepoStars)
	}
	if f.LanguageCount != 0 {
		t.Fatalf("LanguageCount: 0 != %d", f.LanguageCount)
	}
	if f.RankedLanguages != nil {
		t.Fatalf("RankedLanguages: nil != %v", f.RankedLanguages)
	}
	if f.TotalReleases != 0 || f.ReleasedRepos != 0 || f.TotalPullRequests != 0 || f.TotalCollaborators != 0 || f.ProtectedBranchRepos != 0 || f.DiscussionRepos != 0 {
		t.Fatalf("release/maintenance aggregates should be 0 for empty input")
	}
	if f.ForkRatio != 0 || f.LicensedRatio != 0 || f.ArchivedRatio != 0 {
		t.Fatalf("ratio facts should be 0 for empty input")
	}
	if f.PortfolioAgeDays != 0 || f.NewestRepoAgeDays != 0 || f.DaysSinceLatestRelease != 0 {
		t.Fatalf("age facts should be 0 for empty input")
	}
	if f.MeanRepoSize != 0 || f.TopicBreadth != 0 {
		t.Fatalf("size/breadth facts should be 0 for empty input")
	}
}

func TestFromReposRankedLanguagesTiebreak(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, LanguageDistribution: map[string]int64{"Zeta": 100, "Alpha": 100, "Beta": 200}},
	}

	f := FromRepos(repos, refTime)

	want := []string{"Beta", "Alpha", "Zeta"}
	if len(f.RankedLanguages) != len(want) {
		t.Fatalf("RankedLanguages length: %d != %d (%v)", len(f.RankedLanguages), len(want), f.RankedLanguages)
	}
	for i, w := range want {
		if f.RankedLanguages[i] != w {
			t.Fatalf("RankedLanguages[%d]: %q != %q (full: %v)", i, f.RankedLanguages[i], w, f.RankedLanguages)
		}
	}
}

func TestExtractSignalsEquivalence(t *testing.T) {
	cases := []struct {
		name  string
		repos []RepositoryVestige
	}{
		{
			name:  "empty",
			repos: nil,
		},
		{
			name:  "single original",
			repos: []RepositoryVestige{{Fork: false, Size: 100, UpdatedAt: daysAgo(10)}},
		},
		{
			name: "mixed original and fork",
			repos: []RepositoryVestige{
				{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
				{Fork: true, Size: 500, UpdatedAt: daysAgo(5)},
				{Fork: false, Size: 60, UpdatedAt: daysAgo(89)},
				{Fork: false, Size: 20, UpdatedAt: daysAgo(91)},
				{Fork: false, Size: 0, UpdatedAt: daysAgo(150)},
			},
		},
		{
			name: "all forks",
			repos: []RepositoryVestige{
				{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
				{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
			},
		},
		{
			name: "all zero size",
			repos: []RepositoryVestige{
				{Fork: false, Size: 0, UpdatedAt: daysAgo(5)},
				{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
			},
		},
		{
			name: "boundary recent",
			repos: []RepositoryVestige{
				{Fork: false, Size: 100, UpdatedAt: daysAgo(90)},
				{Fork: false, Size: 100, UpdatedAt: daysAgo(91)},
			},
		},
		{
			name: "many repos",
			repos: func() []RepositoryVestige {
				r := make([]RepositoryVestige, 0, 20)
				for i := 0; i < 16; i++ {
					r = append(r, RepositoryVestige{Fork: false, Size: 100, UpdatedAt: daysAgo(10)})
				}
				for i := 0; i < 4; i++ {
					r = append(r, RepositoryVestige{Fork: false, Size: 100, UpdatedAt: daysAgo(120)})
				}
				return r
			}(),
		},
		{
			name: "single stale old repo",
			repos: []RepositoryVestige{
				{Fork: false, Size: 100, UpdatedAt: daysAgo(300)},
			},
		},
		{
			name: "varied activity levels",
			repos: []RepositoryVestige{
				{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
				{Fork: false, Size: 100, UpdatedAt: daysAgo(60)},
				{Fork: false, Size: 100, UpdatedAt: daysAgo(150)},
				{Fork: false, Size: 100, UpdatedAt: daysAgo(300)},
			},
		},
		{
			name: "deep threshold boundary",
			repos: []RepositoryVestige{
				{Fork: false, Size: minDepthRepoSize, UpdatedAt: daysAgo(10)},
				{Fork: false, Size: minDepthRepoSize - 1, UpdatedAt: daysAgo(10)},
				{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
			},
		},
		{
			name: "ownership mixed valid zero size",
			repos: []RepositoryVestige{
				{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
				{Fork: true, Size: 0, UpdatedAt: daysAgo(10)},
				{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
				{Fork: true, Size: 100, UpdatedAt: daysAgo(10)},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s1 := ExtractSignals(tc.repos, refTime)
			s2 := ExtractSignalsFromFacts(FromRepos(tc.repos, refTime), refTime)

			if !almostEqual(s1.Ownership, s2.Ownership) {
				t.Fatalf("Ownership: ExtractSignals=%.4f ExtractSignalsFromFacts=%.4f", s1.Ownership, s2.Ownership)
			}
			if !almostEqual(s1.Consistency, s2.Consistency) {
				t.Fatalf("Consistency: ExtractSignals=%.4f ExtractSignalsFromFacts=%.4f", s1.Consistency, s2.Consistency)
			}
			if !almostEqual(s1.Depth, s2.Depth) {
				t.Fatalf("Depth: ExtractSignals=%.4f ExtractSignalsFromFacts=%.4f", s1.Depth, s2.Depth)
			}
			if !almostEqual(s1.Activity, s2.Activity) {
				t.Fatalf("Activity: ExtractSignals=%.4f ExtractSignalsFromFacts=%.4f", s1.Activity, s2.Activity)
			}
		})
	}
}
