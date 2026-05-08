package workers

import (
	"context"
	"log/slog"

	"riksdagskollen/internal/speeches"
)

// EnrichSpeechesWorker fetches the prose body for speeches that only
// have metadata. The /anforandelista endpoint returns 0-length
// anforandetext so we have to follow up per speech against
// /anforande/{dok}-{nr}.json. Runs daily and processes a batch each
// pass.
type EnrichSpeechesWorker struct {
	svc       *speeches.Service
	batchSize int
}

func NewEnrichSpeechesWorker(svc *speeches.Service, batchSize int) EnrichSpeechesWorker {
	if batchSize <= 0 {
		batchSize = 100
	}
	return EnrichSpeechesWorker{svc: svc, batchSize: batchSize}
}

func (w *EnrichSpeechesWorker) Name() string { return "enrich-speech-text" }

func (w *EnrichSpeechesWorker) Run(ctx context.Context) error {
	enriched, skipped, err := w.svc.EnrichMissingText(ctx, w.batchSize)
	if err != nil {
		return err
	}
	slog.Info("enrich-speech-text done",
		"enriched", enriched,
		"skipped", skipped,
		"batch", w.batchSize,
	)
	return nil
}
