package domain

// AgendaItem is one whole government document, not a point extracted from one.
//
// It carries no description and no status. A description would be our
// paraphrase of the document's words, and a status ('active', 'in_progress')
// would be our claim about how the government is progressing — unsourced, and
// stale from the moment it is written. Title, issuer, date and link let the
// reader read the primary source and judge for themselves.
//
// The list of documents is ours: a missing document means we have not
// registered it, never that it does not exist.
type AgendaItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	// Source is the document's short designation, e.g. "Prop. 2025/26:1".
	Source string `json:"source"`
	// Issuer is who published it — a ministry, or the parties to an agreement.
	Issuer string `json:"issuer"`
	// URL points at the document itself, in full.
	URL string `json:"url"`
	// Published is the publication date as YYYY-MM-DD, nil when we do not hold one.
	Published *string `json:"published"`
}
