DELETE FROM riksdag_government_parties
WHERE government_id IN (
    SELECT id FROM riksdag_government
    WHERE type_label = 'Övergångsregering' AND valid_from = DATE '2026-09-17'
);

DELETE FROM riksdag_government
WHERE type_label = 'Övergångsregering' AND valid_from = DATE '2026-09-17';

UPDATE riksdag_government
SET valid_to = NULL
WHERE valid_from = DATE '2022-10-18' AND valid_to = DATE '2026-09-17';
