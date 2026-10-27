-- Chamber seats after the 13 September 2026 election.
--
-- The row opened in 000013 is the 2022 Tidö coalition (M 68, KD 19, L 16,
-- SD 73 as support, and the 2022 opposition). That parliament dissolved.
-- Valmyndigheten fastställde mandatfördelningen 19 September 2026
-- (dnr VAL-735-2026): S 99, M 70, SD 62, V 30, C 25, KD 22, MP 22, L 19.
-- https://resultat.val.se/val2026/
-- https://www.riksdagen.se/sv/aktuellt/aktuelltnotiser/2026/sep/19/den-nya-riksdagen-efter-valet_cmsad780a37-f1ae-4d47-a4a7-b6ea2a69a21esv/
--
-- No new government has been appointed. Regeringen.se states that a caretaker
-- government (övergångsregering) led by Ulf Kristersson has been in office
-- since 17 September 2026, when the outgoing prime minister requested
-- resignation. The caretaker cabinet is the outgoing one: M, KD and L.
-- Parties outside that cabinet are stored as opposition because the column
-- only allows governing, support or opposition. That is not a claim that they
-- form one bloc, and it is not the 2022 Tidö support agreement — SD is not
-- marked support. The UI treats type_label 'Övergångsregering' as a seat list
-- without a styre/opposition split. When a government is appointed, close
-- this row and insert the new one; do not edit it in place.
UPDATE riksdag_government
SET valid_to = DATE '2026-09-17'
WHERE valid_to IS NULL;

INSERT INTO riksdag_government (type_label, valid_from)
VALUES ('Övergångsregering', DATE '2026-09-17');

INSERT INTO riksdag_government_parties (government_id, name, short, seats, color, role, sort_order)
SELECT g.id, v.name, v.short, v.seats, v.color, v.role, v.sort_order
FROM (
    SELECT id FROM riksdag_government
    WHERE type_label = 'Övergångsregering' AND valid_from = DATE '2026-09-17'
) g
CROSS JOIN (VALUES
    ('Moderaterna',         'M',  70,  '#1E88E5', 'governing',  1),
    ('Kristdemokraterna',   'KD', 22,  '#00558E', 'governing',  2),
    ('Liberalerna',         'L',  19,  '#006AB3', 'governing',  3),
    ('Socialdemokraterna',  'S',  99,  '#E8112D', 'opposition', 4),
    ('Sverigedemokraterna', 'SD', 62,  '#DDD014', 'opposition', 5),
    ('Vänsterpartiet',      'V',  30,  '#AF0000', 'opposition', 6),
    ('Centerpartiet',       'C',  25,  '#009933', 'opposition', 7),
    ('Miljöpartiet',        'MP', 22,  '#83CF39', 'opposition', 8)
) AS v(name, short, seats, color, role, sort_order);
