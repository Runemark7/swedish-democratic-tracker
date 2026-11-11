-- Small-table indexes for the slow-endpoint pass. Plain CREATE INDEX: every
-- table here is under ~100k rows (ballots is deliberately not touched).

-- ListBeslut party/involved filters and the motions list probe by betänkande.
CREATE INDEX IF NOT EXISTS idx_yfp_betankande
    ON yrkande_forslagspunkt (riksmote, beteckning, motion_dok_id);

-- beslut details and the Involved filter join voteringar on (session, beteckning).
CREATE INDEX IF NOT EXISTS idx_voteringar_session_beteckning
    ON voteringar (session, beteckning);

-- Case-insensitive lookups: WHERE upper(v.votering_id) = upper($1).
CREATE INDEX IF NOT EXISTS idx_voteringar_votering_id_upper
    ON voteringar (upper(votering_id));

-- Undecided half of the motions list: newest filing date first, no sort.
CREATE INDEX IF NOT EXISTS idx_motions_list_order
    ON motions ((COALESCE(motion_date::timestamp, 'epoch'::timestamp)) DESC, dok_id DESC);
