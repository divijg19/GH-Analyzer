package signals

import (
	"testing"
	"time"
)

func TestFromReposEmpty(t *testing.T) {
	f := FromRepos(nil)
	assertZeroFacts(t, f)

	f = FromRepos([]Repo{})
	assertZeroFacts(t, f)
}

func TestFromReposCountsTotalRepos(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(30)},
	}

	f := FromRepos(repos)

	if f.TotalRepos != 3 {
		t.Fatalf("expected TotalRepos 3, got %d", f.TotalRepos)
	}
}

func TestFromReposCountsOriginalAndForkRepos(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(30)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(40)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(50)},
	}

	f := FromRepos(repos)

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
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(89)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(91)},
	}

	f := FromRepos(repos)

	if f.RecentRepos != 2 {
		t.Fatalf("expected RecentRepos 2 (10 and 89 days), got %d", f.RecentRepos)
	}
}

func TestFromReposForksNotCountedAsRecent(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
	}

	f := FromRepos(repos)

	if f.RecentRepos != 1 {
		t.Fatalf("expected RecentRepos 1 (fork excluded), got %d", f.RecentRepos)
	}
}

func TestFromReposDeepReposAboveThreshold(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: minDepthRepoSize, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: minDepthRepoSize + 1, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: minDepthRepoSize - 1, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

	if f.DeepRepos != 2 {
		t.Fatalf("expected DeepRepos 2 (exactly at threshold and above), got %d", f.DeepRepos)
	}
}

func TestFromReposForksNotCountedAsDeep(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 1000, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

	if f.DeepRepos != 1 {
		t.Fatalf("expected DeepRepos 1 (fork excluded despite size), got %d", f.DeepRepos)
	}
}

func TestFromReposValidReposExcludesZeroSize(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

	if f.ValidRepos != 2 {
		t.Fatalf("expected ValidRepos 2 (size > 0), got %d", f.ValidRepos)
	}
}

func TestFromReposValidOriginalReposExcludesForksAndZeroSize(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 50, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

	if f.ValidOriginalRepos != 2 {
		t.Fatalf("expected ValidOriginalRepos 2 (non-fork, size > 0), got %d", f.ValidOriginalRepos)
	}
}

func TestFromReposLargestRepoSize(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 999, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 500, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

	if f.LargestRepoSize != 999 {
		t.Fatalf("expected LargestRepoSize 999, got %d", f.LargestRepoSize)
	}
}

func TestFromReposLargestRepoSizeWithAllZeros(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 0, UpdatedAt: daysAgo(10)},
		{Fork: true, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

	if f.LargestRepoSize != 0 {
		t.Fatalf("expected LargestRepoSize 0, got %d", f.LargestRepoSize)
	}
}

func TestFromReposLatestActivity(t *testing.T) {
	now := time.Now()
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: now.Add(-48 * time.Hour)},
		{Fork: false, Size: 100, UpdatedAt: now.Add(-72 * time.Hour)},
		{Fork: false, Size: 100, UpdatedAt: now.Add(-24 * time.Hour)},
	}

	f := FromRepos(repos)

	expected := now.Add(-24 * time.Hour)
	if !f.LatestActivity.Equal(expected) {
		t.Fatalf("expected LatestActivity %v, got %v", expected, f.LatestActivity)
	}
}

func TestFromReposLatestActivityAcrossForksAndOriginals(t *testing.T) {
	now := time.Now()
	repos := []Repo{
		{Fork: true, Size: 100, UpdatedAt: now.Add(-1 * time.Hour)},
		{Fork: false, Size: 100, UpdatedAt: now.Add(-48 * time.Hour)},
	}

	f := FromRepos(repos)

	if !f.LatestActivity.Equal(now.Add(-1 * time.Hour)) {
		t.Fatalf("expected LatestActivity to be fork's recent update, got %v", f.LatestActivity)
	}
}

func TestFromReposAllOriginalAllRecentAllDeepAllValid(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 200, UpdatedAt: daysAgo(5)},
		{Fork: false, Size: 150, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(15)},
	}

	f := FromRepos(repos)

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
	repos := []Repo{
		{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
		{Fork: true, Size: 200, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

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
	repos := []Repo{
		{Fork: false, Size: 75, UpdatedAt: daysAgo(30)},
	}

	f := FromRepos(repos)

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

func TestFromReposNoOriginalRepos(t *testing.T) {
	repos := []Repo{
		{Fork: true, Size: 0, UpdatedAt: daysAgo(5)},
		{Fork: true, Size: 0, UpdatedAt: daysAgo(10)},
	}

	f := FromRepos(repos)

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

func assertZeroFacts(t *testing.T, f Facts) {
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
