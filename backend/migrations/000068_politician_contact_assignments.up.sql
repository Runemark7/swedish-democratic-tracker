-- Contact and assignment data from data.riksdagen.se/personlista, written by
-- the existing politicians sync. Additive: existing rows keep NULL until the
-- next daily sync fills them.
ALTER TABLE politicians
    ADD COLUMN IF NOT EXISTS email        text,
    ADD COLUMN IF NOT EXISTS member_since integer,
    ADD COLUMN IF NOT EXISTS committees   jsonb;
