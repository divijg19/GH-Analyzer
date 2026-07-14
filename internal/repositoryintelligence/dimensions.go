package repositoryintelligence

import (
	"fmt"

	"github.com/divijg19/Atlas/internal/evidence"
)

// 1. Identity — is this repository clearly identified and distinguishable?
type IdentityDimension struct {
	dimensionBase
	Level      Level
	Confidence Confidence
	HasName    bool
	HasTopic   bool
	HasLicense bool
	HasDefault bool
	Template   bool
}

func buildIdentity(rc repoContext) IdentityDimension {
	d := IdentityDimension{}
	d.name = "identity"
	d.HasName = rc.vestige.Name != ""
	d.HasTopic = rc.topicCount > 0
	d.HasLicense = rc.hasLicense
	d.HasDefault = rc.vestige.DefaultBranch != ""
	d.Template = rc.vestige.Template

	var hits int
	if d.HasTopic {
		hits++
	}
	if d.HasLicense {
		hits++
	}
	if d.HasDefault {
		hits++
	}
	if !d.Template {
		hits++
	}
	den := 4
	ratio := float64(hits) / float64(den)
	d.Level = levelFromRatio(ratio, 0.8, 0.4)
	if rc.vestige.Name == "" {
		d.Confidence = ConfidenceLow
	} else {
		d.Confidence = ConfidenceHigh
	}
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "identity", []string{"Topics", "License", "DefaultBranch", "Template"},
			factItem("has topic", d.HasTopic),
			factItem("has license", rc.hasLicense),
			factItem("has default branch", d.HasDefault),
			factItem("is template", d.Template),
		),
	}
	d.summary = fmt.Sprintf("Identity is %s (%d of %d identifying signals present).", d.Level, hits, den)
	return d
}

// 2. Health — is this repository alive and receiving attention?
type HealthDimension struct {
	dimensionBase
	Level       Level
	Confidence  Confidence
	PushAgeDays int
	Stars       int
	OpenIssues  int
}

func buildHealth(rc repoContext) HealthDimension {
	d := HealthDimension{}
	d.name = "health"
	d.PushAgeDays = rc.pushAgeDays
	d.Stars = rc.vestige.Stars
	d.OpenIssues = rc.vestige.OpenIssues

	switch {
	case rc.pushAgeDays <= 90:
		d.Level = LevelActive
	case rc.pushAgeDays <= 365:
		d.Level = LevelDormant
	default:
		d.Level = LevelInactive
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "health", []string{"PushedAt", "Stars", "OpenIssues"},
			factItem("days since last push", rc.pushAgeDays),
			factItem("stars", d.Stars),
			factItem("open issues", d.OpenIssues),
		),
	}
	d.summary = fmt.Sprintf("Health is %s (last push %d days ago; %d stars).", d.Level, rc.pushAgeDays, d.Stars)
	return d
}

// 3. Maintenance — is this repository actively maintained?
type MaintenanceDimension struct {
	dimensionBase
	Level        Level
	Confidence   Confidence
	PushAgeDays  int
	Archived     bool
	ReleaseCount int
}

func buildMaintenance(rc repoContext) MaintenanceDimension {
	d := MaintenanceDimension{}
	d.name = "maintenance"
	d.PushAgeDays = rc.pushAgeDays
	d.Archived = rc.vestige.Archived
	d.ReleaseCount = rc.vestige.ReleaseCount

	switch {
	case rc.vestige.Archived:
		d.Level = LevelInactive
	case rc.pushAgeDays <= 90:
		d.Level = LevelMaintained
	case rc.pushAgeDays <= 365:
		d.Level = LevelDormant
	default:
		d.Level = LevelInactive
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "maintenance", []string{"PushedAt", "Archived", "ReleaseCount"},
			factItem("days since last push", rc.pushAgeDays),
			factItem("archived", rc.vestige.Archived),
			factItem("release count", rc.vestige.ReleaseCount),
		),
	}
	d.summary = fmt.Sprintf("Maintenance is %s (last push %d days ago; archived=%t).", d.Level, rc.pushAgeDays, rc.vestige.Archived)
	return d
}

// 4. Delivery — does this repository ship releases with discipline?
type DeliveryDimension struct {
	dimensionBase
	Level          Level
	Confidence     Confidence
	ReleaseCount   int
	ReleaseAgeDays int
}

func buildDelivery(rc repoContext) DeliveryDimension {
	d := DeliveryDimension{}
	d.name = "delivery"
	d.ReleaseCount = rc.vestige.ReleaseCount
	d.ReleaseAgeDays = rc.releaseAgeDays

	switch {
	case rc.vestige.ReleaseCount == 0:
		d.Level = LevelUndisciplined
	case rc.releaseAgeDays <= 365:
		d.Level = LevelDisciplined
	default:
		d.Level = LevelModerate
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "delivery", []string{"ReleaseCount", "LatestReleaseAt"},
			factItem("release count", rc.vestige.ReleaseCount),
			factItem("days since latest release", rc.releaseAgeDays),
		),
	}
	if rc.vestige.ReleaseCount == 0 {
		d.summary = "Delivery is undisciplined (no releases published)."
	} else {
		d.summary = fmt.Sprintf("Delivery is %s (%d releases; latest %d days ago).", d.Level, rc.vestige.ReleaseCount, rc.releaseAgeDays)
	}
	return d
}

// 5. Architecture — how is this repository structured?
type ArchitectureDimension struct {
	dimensionBase
	Level         Level
	Confidence    Confidence
	Size          int
	LanguageCount int
	PrimaryLang   string
}

func buildArchitecture(rc repoContext) ArchitectureDimension {
	d := ArchitectureDimension{}
	d.name = "architecture"
	d.Size = rc.vestige.Size
	d.LanguageCount = rc.languageCount
	d.PrimaryLang = rc.primaryLang

	switch {
	case rc.vestige.Size == 0:
		d.Level = LevelEmpty
	case rc.languageCount >= 3:
		d.Level = LevelBroad
	case rc.languageCount == 1:
		d.Level = LevelNarrow
	default:
		d.Level = LevelModerate
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "architecture", []string{"Size", "LanguageDistribution"},
			factItem("size (KB)", rc.vestige.Size),
			factItem("language count", rc.languageCount),
			factItem("primary language", d.PrimaryLang),
		),
	}
	d.summary = fmt.Sprintf("Architecture is %s (%dKB; %d languages; primary %s).", d.Level, rc.vestige.Size, rc.languageCount, d.PrimaryLang)
	return d
}

// 6. Technology — what stack does this repository use?
type TechnologyDimension struct {
	dimensionBase
	Level         Level
	Confidence    Confidence
	LanguageCount int
	PrimaryLang   string
	Stack         []string
}

func buildTechnology(rc repoContext) TechnologyDimension {
	d := TechnologyDimension{}
	d.name = "technology"
	d.LanguageCount = rc.languageCount
	d.PrimaryLang = rc.primaryLang
	d.Stack = rc.rankedLangs

	switch {
	case rc.languageCount >= 3:
		d.Level = LevelBroad
	case rc.languageCount == 2:
		d.Level = LevelModerate
	case rc.languageCount == 1:
		d.Level = LevelNarrow
	default:
		d.Level = LevelEmpty
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "technology", []string{"LanguageDistribution"},
			factItem("language count", rc.languageCount),
			factItem("primary language", d.PrimaryLang),
			factItem("languages", fmt.Sprintf("%v", rc.rankedLangs)),
		),
	}
	d.summary = fmt.Sprintf("Technology is %s (primary %s across %d languages).", d.Level, d.PrimaryLang, rc.languageCount)
	return d
}

// 7. Community — how much external engagement does this repository attract?
type CommunityDimension struct {
	dimensionBase
	Level         Level
	Confidence    Confidence
	Forks         int
	Watchers      int
	Collaborators int
	PullRequests  int
	Discussions   bool
}

func buildCommunity(rc repoContext) CommunityDimension {
	d := CommunityDimension{}
	d.name = "community"
	d.Forks = rc.vestige.Forks
	d.Watchers = rc.vestige.Watchers
	d.Collaborators = rc.vestige.CollaboratorCount
	d.PullRequests = rc.vestige.PullRequestCount
	d.Discussions = rc.vestige.DiscussionEnabled

	switch {
	case rc.vestige.Watchers >= 50 || rc.vestige.Forks >= 10 || rc.vestige.CollaboratorCount >= 5:
		d.Level = LevelHigh
	case rc.vestige.Watchers >= 5 || rc.vestige.Forks >= 1 || rc.vestige.CollaboratorCount >= 1 || rc.vestige.PullRequestCount >= 1 || rc.vestige.DiscussionEnabled:
		d.Level = LevelModerate
	default:
		d.Level = LevelLow
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "community", []string{"Forks", "Watchers", "CollaboratorCount", "PullRequestCount", "DiscussionEnabled"},
			factItem("forks", rc.vestige.Forks),
			factItem("watchers", rc.vestige.Watchers),
			factItem("collaborators", rc.vestige.CollaboratorCount),
			factItem("pull requests", rc.vestige.PullRequestCount),
			factItem("discussions enabled", rc.vestige.DiscussionEnabled),
		),
	}
	d.summary = fmt.Sprintf("Community is %s (%d forks, %d watchers, %d collaborators).", d.Level, rc.vestige.Forks, rc.vestige.Watchers, rc.vestige.CollaboratorCount)
	return d
}

// 8. Documentation — is this repository documented?
type DocumentationDimension struct {
	dimensionBase
	Level       Level
	Confidence  Confidence
	TopicCount  int
	HasLicense  bool
	Discussions bool
}

func buildDocumentation(rc repoContext) DocumentationDimension {
	d := DocumentationDimension{}
	d.name = "documentation"
	d.TopicCount = rc.topicCount
	d.HasLicense = rc.hasLicense
	d.Discussions = rc.vestige.DiscussionEnabled

	var hits int
	if rc.topicCount > 0 {
		hits++
	}
	if rc.hasLicense {
		hits++
	}
	if rc.vestige.DiscussionEnabled {
		hits++
	}
	ratio := float64(hits) / 3.0
	d.Level = levelFromRatio(ratio, 0.75, 0.4)
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "documentation", []string{"Topics", "License", "DiscussionEnabled"},
			factItem("has topic", rc.topicCount > 0),
			factItem("has license", rc.hasLicense),
			factItem("discussions enabled", rc.vestige.DiscussionEnabled),
		),
	}
	d.summary = fmt.Sprintf("Documentation is %s (%d of 3 signals present).", d.Level, hits)
	return d
}

// 9. Quality — what observable quality signals does this repository exhibit?
type QualityDimension struct {
	dimensionBase
	Level           Level
	Confidence      Confidence
	HasLicense      bool
	BranchProtected bool
	Collaborators   int
	Discussions     bool
}

func buildQuality(rc repoContext) QualityDimension {
	d := QualityDimension{}
	d.name = "quality"
	d.HasLicense = rc.hasLicense
	d.BranchProtected = rc.vestige.DefaultBranchProtected
	d.Collaborators = rc.vestige.CollaboratorCount
	d.Discussions = rc.vestige.DiscussionEnabled

	var hits int
	if rc.hasLicense {
		hits++
	}
	if rc.vestige.DefaultBranchProtected {
		hits++
	}
	if rc.vestige.CollaboratorCount > 0 {
		hits++
	}
	if rc.vestige.DiscussionEnabled {
		hits++
	}
	ratio := float64(hits) / 4.0
	d.Level = levelFromRatio(ratio, 0.75, 0.4)
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "quality", []string{"License", "DefaultBranchProtected", "CollaboratorCount", "DiscussionEnabled"},
			factItem("has license", rc.hasLicense),
			factItem("default branch protected", rc.vestige.DefaultBranchProtected),
			factItem("collaborators", rc.vestige.CollaboratorCount),
			factItem("discussions enabled", rc.vestige.DiscussionEnabled),
		),
	}
	d.summary = fmt.Sprintf("Quality is %s (%d of 4 observable signals present).", d.Level, hits)
	return d
}

// 10. Complexity — how large and complex is this repository?
type ComplexityDimension struct {
	dimensionBase
	Level         Level
	Confidence    Confidence
	Size          int
	LanguageCount int
}

func buildComplexity(rc repoContext) ComplexityDimension {
	d := ComplexityDimension{}
	d.name = "complexity"
	d.Size = rc.vestige.Size
	d.LanguageCount = rc.languageCount

	switch {
	case rc.vestige.Size == 0:
		d.Level = LevelEmpty
	case rc.vestige.Size < 50:
		d.Level = LevelLow
	case rc.vestige.Size < 1000:
		d.Level = LevelModerate
	default:
		d.Level = LevelHigh
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "complexity", []string{"Size", "LanguageDistribution"},
			factItem("size (KB)", rc.vestige.Size),
			factItem("language count", rc.languageCount),
		),
	}
	d.summary = fmt.Sprintf("Complexity is %s (%dKB across %d languages).", d.Level, rc.vestige.Size, rc.languageCount)
	return d
}

// 11. Lifecycle — what maturity stage is this repository in?
type LifecycleDimension struct {
	dimensionBase
	Level        Level
	Confidence   Confidence
	AgeDays      int
	Archived     bool
	ReleaseCount int
	Template     bool
}

func buildLifecycle(rc repoContext) LifecycleDimension {
	d := LifecycleDimension{}
	d.name = "lifecycle"
	d.AgeDays = rc.ageDays
	d.Archived = rc.vestige.Archived
	d.ReleaseCount = rc.vestige.ReleaseCount
	d.Template = rc.vestige.Template

	switch {
	case rc.vestige.Archived:
		d.Level = LevelInactive
	case rc.ageDays < 180:
		d.Level = LevelEarly
	case rc.ageDays >= 365 && rc.vestige.ReleaseCount > 0 && !rc.vestige.Template:
		d.Level = LevelMature
	default:
		d.Level = LevelModerate
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "lifecycle", []string{"CreatedAt", "Archived", "ReleaseCount", "Template"},
			factItem("age (days)", rc.ageDays),
			factItem("archived", rc.vestige.Archived),
			factItem("release count", rc.vestige.ReleaseCount),
			factItem("is template", rc.vestige.Template),
		),
	}
	d.summary = fmt.Sprintf("Lifecycle is %s (age %d days; archived=%t).", d.Level, rc.ageDays, rc.vestige.Archived)
	return d
}

// 12. Governance — how is this repository owned and protected?
type GovernanceDimension struct {
	dimensionBase
	Level           Level
	Confidence      Confidence
	Collaborators   int
	BranchProtected bool
	Visibility      string
	Fork            bool
}

func buildGovernance(rc repoContext) GovernanceDimension {
	d := GovernanceDimension{}
	d.name = "governance"
	d.Collaborators = rc.vestige.CollaboratorCount
	d.BranchProtected = rc.vestige.DefaultBranchProtected
	d.Visibility = rc.vestige.Visibility
	d.Fork = rc.vestige.Fork

	switch {
	case rc.vestige.DefaultBranchProtected && rc.vestige.CollaboratorCount > 0:
		d.Level = LevelGoverned
	case rc.vestige.Fork || rc.vestige.CollaboratorCount == 0:
		d.Level = LevelIndividual
	default:
		d.Level = LevelModerate
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "governance", []string{"CollaboratorCount", "DefaultBranchProtected", "Visibility", "Fork"},
			factItem("collaborators", rc.vestige.CollaboratorCount),
			factItem("default branch protected", rc.vestige.DefaultBranchProtected),
			factItem("visibility", rc.vestige.Visibility),
			factItem("is fork", rc.vestige.Fork),
		),
	}
	d.summary = fmt.Sprintf("Governance is %s (%d collaborators; protected=%t).", d.Level, rc.vestige.CollaboratorCount, rc.vestige.DefaultBranchProtected)
	return d
}

// 13. Risk — how fragile or at-risk is this repository?
type RiskDimension struct {
	dimensionBase
	Level       Level
	Confidence  Confidence
	Flags       int
	Archived    bool
	Stale       bool
	SingleOwner bool
	NoLicense   bool
	IsFork      bool
}

func buildRisk(rc repoContext) RiskDimension {
	d := RiskDimension{}
	d.name = "risk"
	d.Archived = rc.vestige.Archived
	d.Stale = rc.pushAgeDays > 365
	d.SingleOwner = rc.vestige.CollaboratorCount <= 1
	d.NoLicense = !rc.hasLicense
	d.IsFork = rc.vestige.Fork

	var flags int
	if d.Archived {
		flags++
	}
	if d.Stale {
		flags++
	}
	if d.SingleOwner {
		flags++
	}
	if d.NoLicense {
		flags++
	}
	if d.IsFork {
		flags++
	}
	d.Flags = flags

	switch {
	case flags >= 3:
		d.Level = LevelAtRisk
	case flags >= 1:
		d.Level = LevelModerate
	default:
		d.Level = LevelSound
	}
	d.Confidence = ConfidenceHigh
	d.evidence = []evidence.EvidenceGroup{
		groupFrom(rc, "risk", []string{"Archived", "PushedAt", "CollaboratorCount", "License", "Fork"},
			factItem("archived", d.Archived),
			factItem("stale (no push >365d)", d.Stale),
			factItem("single owner", d.SingleOwner),
			factItem("no license", d.NoLicense),
			factItem("is fork", d.IsFork),
		),
	}
	d.summary = fmt.Sprintf("Risk is %s (%d of 5 fragility flags).", d.Level, flags)
	return d
}
