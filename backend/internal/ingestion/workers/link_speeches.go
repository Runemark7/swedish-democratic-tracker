package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/speeches"
)

// LinkSpeechesWorker links stored anföranden to the betänkande they debate.
// It is a cheap, idempotent pass over the speeches table: re-running it
// rewrites only the links that changed, so it doubles as the relink of
// speeches stored before linking existed, and picks up new speeches and newly
// read debate dates.
type LinkSpeechesWorker struct {
	svc *speeches.Service
}

func NewLinkSpeechesWorker(svc *speeches.Service) LinkSpeechesWorker {
	return LinkSpeechesWorker{svc: svc}
}

func (w *LinkSpeechesWorker) Name() string { return "link-speeches" }

func (w *LinkSpeechesWorker) Run(ctx context.Context) error {
	st, err := w.svc.RelinkDebates(ctx)
	slog.Info("link-speeches done", "seen", st.Seen, "linked", st.Linked, "unlinked", st.Unlinked, "changed", st.Changed)
	return err
}
