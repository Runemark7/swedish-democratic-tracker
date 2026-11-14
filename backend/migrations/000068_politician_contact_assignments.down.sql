ALTER TABLE politicians
    DROP COLUMN IF EXISTS committees,
    DROP COLUMN IF EXISTS member_since,
    DROP COLUMN IF EXISTS email;
