package provenance

// Source names the acquisition backend that produced an observation. It is
// transport-agnostic: it records that an observation originated from GitHub
// without coupling any downstream layer to GitHub's API surface.
type Source string

const (
	SourceGitHub  Source = "github"
	SourceDataset Source = "dataset"
	SourceUnknown Source = ""
)

// ObservationKind enumerates the canonical observation families Atlas records.
type ObservationKind string

const (
	ObservationRepository ObservationKind = "repository"
	ObservationActivity   ObservationKind = "activity"
)

// ObservationRef is a deterministic reference to an observation and, optionally,
// the specific observed field that mattered to a downstream conclusion. An empty
// ID denotes the whole observation family of the given Kind (used by portfolio
// facts that aggregate every repository observation).
type ObservationRef struct {
	Kind   ObservationKind `json:"kind"`
	ID     string          `json:"id,omitempty"`
	Field  string          `json:"field,omitempty"`
	Source Source          `json:"source,omitempty"`
}

// FactRef references a derived fact by its canonical name.
type FactRef struct {
	Name string `json:"name"`
}

// IndicatorRef references a normalized indicator (signal) by its canonical name.
type IndicatorRef struct {
	Signal string `json:"signal"`
}

// RepositoryRef references a repository that contributed to a portfolio-level
// conclusion, optionally naming the repository intelligence dimension consumed.
type RepositoryRef struct {
	Repository string `json:"repository"`
	Dimension  string `json:"dimension,omitempty"`
}

// Chain is the complete deterministic provenance for a single conclusion. It
// binds the pipeline layers together — repositories, indicators, facts,
// observations — without duplicating the underlying data. Each slice records
// only what the conclusion truly consumed; empty slices are omitted from JSON.
type Chain struct {
	Repositories []RepositoryRef  `json:"repositories,omitempty"`
	Indicators   []IndicatorRef   `json:"indicators,omitempty"`
	Facts        []FactRef        `json:"facts,omitempty"`
	Observations []ObservationRef `json:"observations,omitempty"`
}

// IsEmpty reports whether the chain references nothing.
func (c Chain) IsEmpty() bool {
	return len(c.Repositories) == 0 &&
		len(c.Indicators) == 0 &&
		len(c.Facts) == 0 &&
		len(c.Observations) == 0
}

// RepositoryField constructs an observation reference to a single repository's
// observed field. Source is GitHub because every repository observation is
// acquired from GitHub.
func RepositoryField(id, field string) ObservationRef {
	return ObservationRef{Kind: ObservationRepository, ID: id, Field: field, Source: SourceGitHub}
}

// RepositoryFamilyField references a field across the entire repository
// observation family (ID omitted). Portfolio facts that aggregate every
// repository use this form.
func RepositoryFamilyField(field string) ObservationRef {
	return ObservationRef{Kind: ObservationRepository, Field: field, Source: SourceGitHub}
}

// Facts constructs fact references from canonical fact names.
func Facts(names ...string) []FactRef {
	if len(names) == 0 {
		return nil
	}
	refs := make([]FactRef, 0, len(names))
	for _, n := range names {
		refs = append(refs, FactRef{Name: n})
	}
	return refs
}

// Indicators constructs indicator references from canonical signal names.
func Indicators(signals ...string) []IndicatorRef {
	if len(signals) == 0 {
		return nil
	}
	refs := make([]IndicatorRef, 0, len(signals))
	for _, s := range signals {
		refs = append(refs, IndicatorRef{Signal: s})
	}
	return refs
}

// RepositoryObservations builds observation references to a single repository's
// fields, preserving field order for deterministic output.
func RepositoryObservations(id string, fields ...string) []ObservationRef {
	if len(fields) == 0 {
		return nil
	}
	refs := make([]ObservationRef, 0, len(fields))
	for _, f := range fields {
		refs = append(refs, RepositoryField(id, f))
	}
	return refs
}

// Merge combines chains deterministically, concatenating each layer in argument
// order. Callers are responsible for passing chains in a stable order.
func Merge(chains ...Chain) Chain {
	var out Chain
	for _, c := range chains {
		out.Repositories = append(out.Repositories, c.Repositories...)
		out.Indicators = append(out.Indicators, c.Indicators...)
		out.Facts = append(out.Facts, c.Facts...)
		out.Observations = append(out.Observations, c.Observations...)
	}
	return out
}
