# Research: which utgiftsområden does each utskott bereda?

**Date:** 2026-07-30
**Resolves:** wayfinder ticket "Which utgiftsområden does each utskott bereda?" (#84)
**Status:** findings — not yet a data-source doc. Phase 2 turns this into
`docs/data-sources/utskott-utgiftsomrade.md` per CLAUDE.md rule 11.

## Where the answer lives

Not on riksdagen.se's committee pages — all 15 were checked and contain no
`utgiftsområde` string. The mapping is **statutory**: the **Bilaga
(tilläggsbestämmelse 7.5.1) till riksdagsordningen (2014:801)**, "Utskottens
ämnesområden". Each committee's final lettered item reads
`anslag inom utgiftsområde(na) N …`.

Chain of authority:

- RO **7 kap. 5 §** — the Riksdag prescribes by tilläggsbestämmelse how ärenden
  are distributed between committees.
- RO **tb 7.5.1** — points at the Bilaga.
- **Bilagan** — the 15-item list, each ending with its UO allocation.
- RO **tb 9.5.3** — canonical list of the 27 utgiftsområden and their names.

## The mapping (riksmöte 2025/26)

Source: Bilaga (tb 7.5.1) RO 2014:801, lydelse per **Lag (2023:507)**, in force
2023-09-01.

| UO | Official name | Utskott |
|---|---|---|
| 1 | Rikets styrelse | KU |
| 2 | Samhällsekonomi och finansförvaltning | FiU |
| 3 | Skatt, tull och exekution | SkU |
| 4 | Rättsväsendet | JuU |
| 5 | Internationell samverkan | UU |
| 6 | Försvar och samhällets krisberedskap | FöU |
| 7 | Internationellt bistånd | UU |
| 8 | Migration | SfU |
| 9 | Hälsovård, sjukvård och social omsorg | SoU |
| 10 | Ekonomisk trygghet vid sjukdom och funktionsnedsättning | SfU |
| 11 | Ekonomisk trygghet vid ålderdom | SfU |
| 12 | Ekonomisk trygghet för familjer och barn | SfU |
| 13 | Integration och jämställdhet | AU |
| 14 | Arbetsmarknad och arbetsliv | AU |
| 15 | Studiestöd | UbU |
| 16 | Utbildning och universitetsforskning | UbU |
| 17 | Kultur, medier, trossamfund och fritid | KrU |
| 18 | Samhällsplanering, bostadsförsörjning och byggande samt konsumentpolitik | CU |
| 19 | Regional utveckling | NU |
| 20 | Klimat, miljö och natur | MJU |
| 21 | Energi | NU |
| 22 | Kommunikationer | TU |
| 23 | Areella näringar, landsbygd och livsmedel | MJU |
| 24 | Näringsliv | NU |
| 25 | Allmänna bidrag till kommuner | FiU |
| 26 | Statsskuldsräntor m.m. | FiU |
| 27 | Avgiften till Europeiska unionen | FiU |

Inverted:

```
KU  → 1            SoU → 9
FiU → 2,25,26,27   KrU → 17
SkU → 3            UbU → 15,16
JuU → 4            TU  → 22
CU  → 18           MJU → 20,23
UU  → 5,7          NU  → 19,21,24
FöU → 6            AU  → 13,14
SfU → 8,10,11,12
```

## Cardinality: a strict partition

**Every UO has exactly one committee; every committee has at least one UO.** No
UO is shared, split, or unassigned. No orphans in either direction.

Verified twice — statute text, and independently against practice: querying
Riksdagen's open data for `doktyp=bet, rm=2025/26` titled "Utgiftsområde …"
returns exactly **27 documents, one per UO**, and the issuing `organ` matches
all 27 rows.

Legal escape hatches exist but are not exercised for UO allocation: RO 7 kap.
6 § authorises splitting the *budget bill* between committees (which is how the
partition is executed, not a violation of it), and RO 7 kap. 11 § permits
case-by-case deviation — no instance found for a UO.

### Betänkande suffix is NOT the UO number

Critical for implementation. `FiU2 → UO2` but `NU2 → UO19`:

```
KU1→1   FiU2→2  SkU1→3  JuU1→4  UU1→5   FöU1→6  UU2→7
SfU4→8  SoU1→9  SfU1→10 SfU2→11 SfU3→12 AU1→13  AU2→14
UbU2→15 UbU1→16 KrU1→17 CU1→18  NU2→19  MJU1→20 NU3→21
TU1→22  MJU2→23 NU1→24  FiU3→25 FiU4→26 FiU5→27
```

## FiU has two distinct roles — model them separately

Conflating these produces a wrong UI.

**(a) Ordinary beredande utskott** — bilaga 2 j: UO **2, 25, 26, 27**. Same in
kind as JuU's UO 4. Products `FiU2, FiU3, FiU4, FiU5`.

**(b) Rambeslut / overall budget responsibility** — RO 7 kap. 9 §, *not* in the
bilaga. Under RO 11 kap. 18 § the budget is decided in two steps:

1. **Rambeslut** — one vote fixing the income estimate plus an `utgiftsram` per
   UO. Prepared by FiU alone (`2025/26:FiU1`). Sector committees only submit
   **yttranden** (RO 10 kap. 7 §).
2. **Anslagsbeslut** — per UO, by the committee in the bilaga, bound by the ram.
3. **Compilation** — FiU assembles the whole budget (`2025/26:FiU10`).

So FiU *bereder* 4 UOs but *ramsätter* all 27 and compiles all 27.

## SkU: the opposite misrepresentation risk

SkU bereder only UO 3 (`Skatt, tull och exekution` — the *administration*). The
**tax revenue side is not a UO at all**: RO 7 kap. 10 § gives SkU "statliga och
kommunala skatter", and revenue enters via the income estimate in step 1, on
which SkU yttrar sig. A UI implying "SkU = 1 UO = small" badly misrepresents it.

## Stability: fixed in law, not restated per riksmöte

The Bilaga has been amended **6 times in 12 years**. There is no annual
restatement — the allocation is not re-decided each year.

| SFS | In force | Omfattning |
|---|---|---|
| 2015:382 | 2015-09-01 | ändr. tb 9.5.3, bil. |
| 2019:26 | 2019-04-01 | ändr. bil. (Riksrevisionen) |
| 2020:104 | 2020-04-01 | ändr. bil. tb 7.5.1 |
| 2020:608 | 2020-09-01 | ändr. tb 9.5.3, bil. |
| **2023:507** | **2023-09-01** | ändr. tb 9.5.3, bil. — *current* |
| **2026:1349** | **2026-09-01** | ändr. bil. (tb 7.5.1) — *imminent* |

**Version by validity interval (in-force date), not by riksmöte.**

Diffs of the two most recent:

- **2023:507** — UO *names* only. UO 13 → `Integration och jämställdhet`;
  UO 20 → `Klimat, miljö och natur`. Assignments unchanged.
- **2026:1349** — **no UO reassignment at all.** Subject-area changes only:
  `immaterialrätt` NU→CU, `upphandling i allmänhet` FiU→NU, KrU's
  `trossamfunden` loses its KU carve-out. The `anslag inom utgiftsområde…`
  lines are byte-identical between lydelser.

### ⚠️ UO *contents* do shift annually

RO tb 9.5.3 second paragraph: decisions on which ändamål och verksamheter fall
inside a UO are taken with the ekonomiska vårpropositionen. Prepared by **KU**
each spring — e.g. `2025/26:KU42 "Indelning i utgiftsområden"` (2026-06-10),
which moved activities between UO 2, 5, 6, 20, 22, 24.

Framst. 2025/26:RS5 states that transferring an anslag-financed activity between
UOs *"kan innebära att beredningsansvaret flyttas från ett utskott till ett
annat"*. So an agency or anslag can change committee **without the bilaga
changing**. A committee↔UO mapping needs no annual versioning; a
committee↔anslag or committee↔myndighet mapping would.

## Primary sources

| What | URL |
|---|---|
| RO 2014:801 consolidated, incl. Bilaga (best single source) | `https://rkrattsbaser.gov.se/sfst?bet=2014:801` |
| RO 2014:801 amendment register | `https://rkrattsbaser.gov.se/sfsr?bet=2014:801` |
| SFS 2023:507 (current lydelse) | `https://svenskforfattningssamling.se/sites/default/files/sfs/2023-06/SFS2023-507.pdf` |
| SFS 2020:608 (previous) | `https://svenskforfattningssamling.se/sites/default/files/sfs/2020-06/SFS2020-608.pdf` |
| Cross-check: 27 UO betänkanden | `https://data.riksdagen.se/dokumentlista/?doktyp=bet&rm=2025%2F26&sok=utgiftsomr%C3%A5de&utformat=json&sz=200` |

⚠️ riksdagen.se's HTML rendering of RO **truncates the Bilaga**; lagen.nu also
failed. Use `rkrattsbaser.gov.se/sfst?bet=2014:801`.

## Reproduction

```bash
curl -sSL "https://rkrattsbaser.gov.se/sfst?bet=2014:801" -o ro.html
# strip tags to ro.txt, then:
grep -n "anslag inom utgiftsområde" ro.txt      # 15 hits per lydelse
grep -A20 "Tilläggsbestämmelse 9.5.3" ro.txt    # canonical UO names

# cross-check against practice, any riksmöte:
curl -sSL "https://data.riksdagen.se/dokumentlista/?doktyp=bet&rm=2025%2F26&sok=utgiftsomr%C3%A5de&utformat=json&sz=200" \
 | jq -r '.dokumentlista.dokument[] | select(.titel|startswith("Utgiftsområde")) | [.organ,.beteckning,.titel]|@tsv' | sort
```

Gotchas: the document contains **two full copies of the Bilaga**, marked
`/Upphör att gälla U:2026-09-01/` and `/Träder i kraft I:2026-09-01/` — parse the
marker, not the first match. rkrattsbaser mangles some en-dashes; harmless here.

## Gaps — deliberately unverified

1. **What SFS 2015:382 changed in the bilaga.** Register entry confirmed;
   act text not retrieved (svenskforfattningssamling.se carries SFS from 2018
   only). Route: `bet. 2014/15:KU17`.
2. **The 2014 original bilaga lydelse** — not retrieved.
3. **KU's and JuU's yttranden on the 2025/26 utgiftsramar** did not surface in
   the API query. This is a search-coverage artifact — do **not** assert they
   were absent.
4. **riksdagen.se committee landing pages state no UO** — verified negative for
   all 15. Never cite them as the source; cite the bilaga.

## Tooling note

`mgrep` is not installed on this machine (`command not found`; the plugin ships
`SKILL.md` but no `bin/`). Findings were gathered with `curl` against primary
sources plus WebFetch/WebSearch.
