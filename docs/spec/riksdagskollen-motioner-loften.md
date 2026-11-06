# Spec: Riksdagskollen — Motioner first (Phase 1), Löften after (Phase 2)

> Draft v3, 2026-10-28. It replaces the 2026-10-27 draft ("Beslut first, with a Löfte engine"). On 2026-10-28, Alexander (owner and Redaktör) decided:
> 1. **Motioner first, Löften after.**
> 2. **Top nav: Beslut · Motioner · Ledamöter · Partier · Regeringen.** Regeringen holds the budget, the Riksdag overview and the agencies. Lokalt stays as a side menu. Om sajten and Data are secondary.
> 3. **Phase 1 links only what is certain:** a cited motion beteckning, identical yrkande wording, or a reference in data.riksdagen.se. There is no manual linking in Phase 1.
> 4. **Löften live as a section under Partier**, not as a top-menu item.
> 5. **Speeches are backfilled as far back as the voteringar go (2002/03).**
> 6. **If all of a party's Bekräftade kopplingar for a Löfte have Partilinje Avstår or Delad, its Löftesstatus is Inte prövat än.**
>
> All the Löfte decisions from 2026-10-27 stand unchanged and form Phase 2: Löftesstatus is vote-based, curation happens in chat through a write API with an audit log (Ändringslogg), and suggestions come in a weekly Veckobatch. Manual curation of motion links also starts in Phase 2. Vocabulary follows [`CONTEXT.md`](../../CONTEXT.md) (the glossary); decisions are in [`adr/`](../../adr/).
>
> Prerequisite: the separate quick-wins PR (clickable front-page feed, the Riksdagskollen name, the committee lookup fix for 2026–2030), and the CI fix that runs the Go tests.
>
> Every data.riksdagen.se endpoint and field named here was checked against the live API on 2026-10-28.

## Problem Statement

A voter who wants to know "what happened to the proposal my party (or I) cared about" cannot find out on Riksdagskollen. The site only knows **voteringar**. It does not know **motioner**, which are the proposals members and parties actually file. So it cannot follow a demand from the motion, through the utskott and the chamber, to the government's answer.

**Worked example: moms on CS:GO skins.**

- **The motion.** In 2017 the Moderaterna filed a kommittémotion, *2017/18:3293* by Maria Malmer Stenergard m.fl. (M). It explained that trade in virtual goods such as Counter-Strike skins first qualified for vinstmarginalbeskattning (VMB), the moms rule for second-hand goods that taxes only the dealer's margin. Skatteverket then reclassified skin sales as electronic services. That means full 25 % moms on the whole sale price. The motion's yrkande 1 asked for an inquiry into the tax treatment of trade in virtual goods.
- **The utskott.** Skatteutskottet handled it in betänkande *2017/18:SkU11 Mervärdesskatt*, förslagspunkt 6 "Övriga frågor". The utskott proposed avslag and wrote that it "gör ingen annan bedömning om virtuella varor än Skatteverket". M (reservation 8) and KD (reservation 9) reserved themselves.
- **The chamber.** It decided on 2018-03-28 by **acklamation**. There was no votering, so no party-by-party vote exists for the original motion.
- **The demand returned.**
  - M filed the identical demand again as 2018/19:2900 yrkande 26, 2019/20:2718 yrkande 26 and 2021/22:3638 yrkande 9. Each was avslagen.
  - In 2021/22 (SkU16 p. 6), M's own reservation was finally put to a vote. M voted Nej (for its reservation). S, SD, V, L and MP voted Ja (for the utskott's avslag), and C and KD abstained.
  - Rickard Nordin (C) then filed enskilda motioner for VMB on digital products in 2022/23, 2023/24 and 2024/25. All were avslagna, with only C voting against avslag in the first two.
- **The government's answer.** Nordin asked skriftlig fråga 2022/23:339, which explicitly cites "en kommittémotion från Moderaterna redan 2017". Finansminister Elisabeth Svantesson (M) answered on 2023-02-22: the rules are largely EU-harmonised, and "det finns inget beslut om att tillsätta en utredning".

Today the site shows none of this chain:

- There is no motion, yrkande, betänkande, förslagspunkt or reservation data in the database. Only voteringar and ballots exist.
- The decision page cannot say whose proposal a votering was about. The step that fills in each votering's origin writes by beteckning and förslagspunkt **without riksmöte**. Every "SkU16 punkt 6" since 2011/12 now carries the same 2021/22 origin, and other origins are wrong or empty.
- **Acklamation decisions are invisible.** No votering exists for them, and the site has no other record of them. That is exactly how 2017/18:3293 was decided.
- Speeches cover only 2025-05..2025-09 and 2026-08..2026-10. 610 have no related document, and none link to the betänkanden above.
- Skriftliga frågor, svar and interpellationer are not ingested at all.
- The top navigation mixes Riksdag, Regering, Budget, Manifest, kommuner and regioner as equals, so the core question is buried.

The earlier problems still stand and are solved in Phase 2: every party promise on the site is from 2022, 96 % of the 8,453 keyword matches between goals and votes are "unclear" and none is verified, and Mål/Manifest/Löfte overlap.

## Solution

Riksdagskollen answers one question for an ordinary voter, in two phases: **"Vad beslutar riksdagen — och håller de vad de lovar?"**

### Phase 1: Motioner

1. **Motions in Postgres.** Every motion from data.riksdagen.se is stored with:
   - its yrkanden;
   - its authors and their party;
   - its kind: **Kommittémotion**, **Partimotion** or **Enskild motion**, which are Riksdagen's own subtyp values. **Flerpartimotion** is derived by us: an enskild motion whose signers come from more than one party. The API has no such subtyp.
2. **Every yrkande resolved to its decision.** Each yrkande is linked to its betänkande and förslagspunkt. From there it gets either the specific votering or **"Avgjord utan omröstning (acklamation)"**. Every förslagspunkt stores:
   - its reservations, and which one was put to the vote;
   - the utskott's reasoning text.
3. **The motion page shows the whole chain**, in this order:
   1. the motion and its yrkanden;
   2. the utskott and why it proposed what it did;
   3. reservations;
   4. the vote per party and per member, or the acklamation;
   5. the debate (anföranden);
   6. follow-up skriftliga frågor, interpellationer and the government's answers;
   7. **Återkommande krav**, meaning the same demand filed again in other years. In Phase 1 these are grouped only when the wording is identical.
4. **The Motioner list.** Filters for party, utskott, outcome, riksmöte and subtyp. Each motion shows a plain-language outcome: *Avslagen*, *Bifallen*, *Delvis bifallen*, *Besvarad*, or *Ej avgjord än*.
5. **Fixed origins.** The broken origin step is replaced. A votering's "what was this about" now comes from the yrkanden handled in its förslagspunkt, resolved per riksmöte. The corrupted rows are recomputed.
6. **Speeches backfilled** for every riksmöte the site has voteringar for, back to 2002/03, and linked to betänkanden.
7. **Questions and answers.** Skriftliga frågor with their svar, and interpellationer, are ingested. They are linked to motions only automatically: when they cite a motion's beteckning, or when Riksdagen's data references the motion. A citation in prose only, such as "en kommittémotion från Moderaterna redan 2017" in fråga 2022/23:339, is not linked in Phase 1. Such links wait for Phase 2 curation.
8. **New navigation:**
   - Top menu: **Beslut · Motioner · Ledamöter · Partier · Regeringen**.
   - **Regeringen** holds who governs, the ministers, the budget (formerly Budget), the Riksdag overview (formerly Riksdag) and the agencies (myndigheter).
   - **Lokalt** (kommun, region, compare) is a visually separate side menu.
   - **Om sajten** and **Data** are secondary links.
   - Old URLs redirect.
9. **Front page variant A**, adjusted for motions, in this order:
   - the question and the optional **Mitt parti** picker;
   - the 5 latest decided betänkanden with each party's Partilinje;
   - **"Senaste motionsutfall"**, the latest decided motions with party, subtyp and outcome;
   - the 3 next votes coming up.

   Every row links somewhere.
10. **Cross-links.** Beslut pages list the motions each förslagspunkt decided. Ledamöter pages list the member's motions (with outcomes), frågor and interpellationer.

### Phase 2: Löften (unchanged from the 2026-10-27 decisions)

- **Löften from the 2026 valmanifest.**
  - Each of the 8 Riksdag parties gets 15–25 concrete, testable Löften.
  - Every Löfte shows the exact quote and a link to the source, with a page or anchor.
  - The 2022 promises stay available as a closed archive.
- **Löftesstatus, from votes only.** Each Löfte shows one of Röstat för, Röstat emot, Blandat or Inte prövat än. The status is derived only from the votes the Redaktör has confirmed as testing that Löfte (Bekräftade kopplingar), judged by the party's Partilinje. If all of them have Partilinje Avstår or Delad, the status is Inte prövat än. The site never says a promise was kept or broken.
- **Löften live under Partier.** Each party page gets a Löften section, and there is a Löften overview across parties reached from Partier. Löften are not a top-menu item.
- **Curation without an admin UI.**
  - The bot extracts candidate Löften and suggests promise↔vote links (Föreslagna kopplingar), dropping weak ones.
  - It also suggests the motion links Phase 1 could not make automatically: questions that cite a motion only in prose, and Återkommande krav with different wording.
  - Once a week it presents a numbered Veckobatch of at most about 20 in the bot chat. The Redaktör answers yes or no there.
  - The bot writes the answers through the token-protected write API.
  - All data lives in Postgres, and every change is recorded in an audit log that can be undone.
- **Additions to the front page:** "berör N löften" tags on decisions, and 8 party tiles summarising Löftesstatus.
- **Additions to the motion page:** a motion page shows the Löften its förslagspunkt's votering is confirmed to test, plus the follow-up questions and Återkommande krav the Redaktör has confirmed.

## User Stories

## Phase 1: Motioner

### Väljare: a motion's whole chain (worked example: 2017/18:3293)

1. As a väljare, I want one canonical page per motion, identified by riksmöte and beteckning (for example 2017/18:3293), so that I can find and share it.
2. As a väljare, I want the motion page to show the title, date, kind (Kommittémotion, Partimotion, Enskild motion, or Flerpartimotion when an enskild motion has signers from several parties), the authors with their party, and the utskott it went to, so that I know who proposed what and where it went.
3. As a väljare, I want the motion's subtyp explained in one plain sentence (for example "a kommittémotion is filed by a party's members in an utskott and usually states the party's line"), so that I understand how much weight it carries.
4. As a väljare, I want every yrkande listed with its exact wording, so that I see each concrete demand, not just the title.
5. As a väljare, I want each yrkande to show its plain-language outcome (Avslagen, Bifallen, Delvis bifallen, Besvarad or Ej avgjord än), so that I understand the result without knowing riksdag procedure.
6. As a väljare, I want each yrkande to link to the betänkande and förslagspunkt that decided it (for example 2017/18:SkU11 punkt 6 "Övriga frågor"), so that I can see where it was handled.
7. As a väljare, I want to read the utskott's reasoning for that förslagspunkt (for example "Utskottet gör ingen annan bedömning om virtuella varor än Skatteverket"), so that I understand why it was rejected.
8. As a väljare, I want to see the reservations on that förslagspunkt, with the parties and members behind each and their reasoning, so that I see who disagreed in the utskott and why.
9. As a väljare, I want to see how each party and each member voted on the förslagspunkt when there was a votering, so that I can see who stood where.
10. As a väljare, I want to be told which alternative the chamber actually voted on (for example "utskottets förslag mot reservation 2 (SD)"), so that I do not assume a vote was on my party's reservation when it was not.
11. As a väljare, I want a decision taken by acklamation shown as "Avgjord utan omröstning (acklamation)", with the date and the reservations as the only record of party positions, so that a missing vote is explained rather than shown as missing data.
12. As a väljare, I want to see the chamber debate on that betänkande (speakers, party, an excerpt and a link to the full anförande), so that I can hear the arguments.
13. As a väljare, I want to see skriftliga frågor and interpellationer that followed up the motion, with the minister's answer, when the question cites the motion's beteckning or Riksdagen's data references it, so that I see what the government said about it. A question that names the motion only in prose is not linked in Phase 1. For example, skriftlig fråga 2022/23:339 and Svantesson's answer of 2023-02-22 reach the motion page only in Phase 2.
14. As a väljare, I want an "Återkommande krav" section listing every other year the same demand was filed, by whom and with what outcome, so that I see how persistent the demand was and how each party's stance changed (for example M 2018/19:2900 yrkande 26, 2019/20:2718 yrkande 26 and 2021/22:3638 yrkande 9, grouped automatically because the wording is identical).
15. As a väljare, I want a timeline at the top of the motion page (filed, in the utskott, justerat, debated, decided, follow-up questions), so that I see the whole path at a glance.
16. As a väljare, I want every fact on the motion page to link to its source at Riksdagen, so that I can verify it.
17. As a väljare, I want the motion page to state facts only and never to judge whether the decision was good or what it cost, so that I draw my own conclusions.
18. As a väljare, I want a motion that is not yet decided shown as "Ej avgjord än", with the betänkande it is waiting for if known, so that I can follow it.

### Väljare: Motioner list

19. As a väljare, I want a Motioner list sorted by most recently decided, so that I see what just happened.
20. As a väljare, I want to filter motions by party, utskott, outcome, riksmöte and subtyp, so that I can find, for example, all of M's avslagna kommittémotioner in Skatteutskottet.
21. As a väljare, I want to search motions by words in the title and yrkanden (for example "virtuella varor"), so that I can find a motion I only half remember.
22. As a väljare, I want each list row to show the title, the party or parties, the subtyp, the date and the outcome, so that I can scan quickly.
23. As a väljare, I want a motion with several yrkanden and mixed outcomes shown as "Delvis bifallen", with the counts, so that it is neither over- nor understated.
24. As a väljare, I want Mitt parti to pre-highlight my party's motions in the list (not filter them), so that I can find them without hiding the others.

### Väljare: Beslut

25. As a väljare, I want one Beslut list of betänkanden sorted by date, so that I can browse decisions in order.
26. As a väljare, I want to filter Beslut by decided versus coming up, by utskott and by party, so that I can narrow down to what interests me.
27. As a väljare, I want one canonical page per betänkande, identified by riksmöte and beteckning, so that two betänkanden with the same beteckning in different years are never confused.
28. As a väljare, I want the betänkande page to list all its förslagspunkter with each party's Partilinje, so that I see the complete decision, not just point 1.
29. As a väljare, I want each förslagspunkt to show its Avvikare separately (members who voted against their Partilinje), so that I can see dissent without it changing the party's record.
30. As a väljare, I want each förslagspunkt to list the motions and yrkanden it decided, each linked to its motion page, so that I can see whose proposals were on the table.
31. As a väljare, I want each förslagspunkt to show its reservations and whether it was decided by votering or acklamation, so that the record is complete.
32. As a väljare, I want the betänkande page to link to the debate and speeches about it, so that I can hear the arguments.
33. As a väljare, I want every fact on a decision page to link to its source at Riksdagen, so that I can verify it.
34. As a väljare, I want the share page for a vote to keep working, so that I can still share decisions.
35. As a väljare, I want a votering's "proposed by" tag to always be correct for its own riksmöte, so that a 2025/26 vote is never labelled with a 2021/22 motion.

### Väljare: Ledamöter and Partier

36. As a väljare, I want a member page that lists the member's motions with their outcomes, so that I see what my representative tried to change.
37. As a väljare, I want the member page to list the skriftliga frågor and interpellationer the member asked (or, for a minister, answered), so that I can follow their follow-up work (for example 2022/23:339 on both Rickard Nordin's and Elisabeth Svantesson's pages).
38. As a väljare, I want the member page to show how the member voted, including where they were an Avvikare, so that I can follow my representative.
39. As a väljare, I want the member page not to show an empty Löften tab, so that the site does not imply members have no promises.
40. As a väljare, I want a party page that shows the party's recent decisions with its Partilinje, and its kommittémotioner and their outcomes, then its members and ministers, so that I see what the party did and demanded.

### Väljare: front page

41. As a väljare, I want to see the question "Vad beslutar riksdagen — och håller de vad de lovar?" at the top of the front page, so that I immediately understand what the site is for.
42. As a väljare, I want to pick Mitt parti once, so that the front page highlights my party without me creating an account.
43. As a väljare, I want my Mitt parti choice to stay only in my browser, so that no personal data about my political preference is stored by the site.
44. As a väljare, I want to change or clear Mitt parti at any time, so that I am not locked into a choice.
45. As a väljare, I want the site to work fully without picking Mitt parti, so that personalisation is never a gate.
46. As a väljare, I want to see the 5 most recently decided betänkanden on the front page, so that I know what the Riksdag decided lately.
47. As a väljare, I want each decided betänkande to show its date, a plain title, its utskott and the outcome, so that I understand it without opening Riksdagen's documents.
48. As a väljare, I want each decided betänkande to show how each of the 8 parties voted (Partilinje), so that I can see where my party stood.
49. As a väljare, I want a "Senaste motionsutfall" block showing recently decided motions with party, subtyp and outcome (Mitt parti's highlighted), so that I see whose demands were accepted or rejected.
50. As a väljare, I want to click a decided betänkande or motion and land on its page, so that I can read the details.
51. As a väljare, I want to click a party's vote mark on a decision row and land on that party, so that I can explore the party from there.
52. As a väljare, I want to see the 3 next betänkanden coming up for a vote, with their planned beslutsdag, so that I know what is coming.
53. As a väljare, I want upcoming items to be clearly separated from decided ones, so that I never mistake a planned vote for a decision.
54. As a väljare, I want "Alla beslut" and "Alla motioner" links from the front page, so that I can see more.
55. As a väljare, I want the front page to say plainly which dates the decisions cover when nothing new has been decided, so that I am not misled about recency.

### Väljare: navigation, Regeringen and Lokalt

56. As a väljare, I want a top menu with only Beslut, Motioner, Ledamöter, Partier and Regeringen, so that I can find the core parts quickly.
57. As a väljare, I want one Regeringen page that shows who governs right now (currently a caretaker government), the ministers, the state budget, the Riksdag overview and the agencies, so that government and budget information is in one place instead of three menu items.
58. As a väljare, I want the budget area pages to keep working inside Regeringen, so that the budget details I used are not lost.
59. As a väljare, I want Lokalt (kommun, region, compare) to look clearly separate from the core menu, so that I understand it is a side section.
60. As a väljare, I want Om sajten and Data as secondary links, so that method and sources are always one click away without crowding the menu.
61. As a väljare, I want the site to be called Riksdagskollen everywhere, so that I am not confused by two names.
62. As a väljare who bookmarked an old page (such as votes, budget, riksdag, regering, manifestos or a party goal's votes), I want it to redirect to its new place, so that my links still work.
63. As a väljare, I want every motion, decision, party, member, debate, question and betänkande page to link to the others it relates to, so that I can follow any thread without searching.
64. As a väljare, I want Lokalt to keep working exactly as before, so that the kommun and region information I use is not lost.

### Journalist (Phase 1)

65. As a journalist, I want the public read API to expose motions, yrkanden, their decision links, reservations, frågor and svar, so that I can analyse them myself.
66. As a journalist, I want to query all motions by a party in an utskott over several riksmöten with their outcomes, so that I can report on what a party demanded and how it went.
67. As a journalist, I want Återkommande krav and motion links exposed as data, with the rule that made each link (Riksdagen reference, cited beteckning or identical wording), so that I can judge it.
68. As a journalist, I want Avvikare per votering available on the vote page and in the API, so that I can report on dissent.
69. As a journalist, I want the Data section to describe each Riksdagen source and its coverage (riksmöten, last update), so that I can assess the dataset's limits.

### Automatic linking (Phase 1)

70. As the operator, I want links made only when they are certain (a structured reference in Riksdagen's data, an exact motion beteckning in the text, or identical normalised yrkande wording), so that Phase 1 needs no manual curation and nothing on the site is guessed.
71. As the operator, I want possible links that fail these rules to be neither shown nor stored as links, so that Phase 1 has no hidden curation queue and Phase 2 starts its suggestions from a clean slate.

### Operator (Phase 1)

72. As the operator, I want motion, betänkande, question and speech ingestion to be idempotent and resumable per riksmöte, with coverage recorded, so that a backfill can be stopped and restarted safely.
73. As the operator, I want ingestion to respect a polite request rate to data.riksdagen.se, so that the site is not blocked.
74. As the operator, I want the corrupted origin rows recomputed by the new linking, not patched by hand, so that the fix is reproducible.
75. As the operator, I want speeches backfilled newest first back to 2002/03, the oldest riksmöte with voteringar, with coverage per riksmöte, so that every votering's debate is available.

## Phase 2: Löften

### Väljare: front page (additions)

76. As a väljare, I want each decided betänkande to show "berör N löften" when it has Bekräftade kopplingar, so that I can see which decisions matter for promises.
77. As a väljare, I want to click "berör N löften" and see exactly those Löften, so that I can check the link myself.
78. As a väljare, I want to see 8 party tiles summarising how many Löften are Röstat för, Röstat emot, Blandat and Inte prövat än, so that I get an overview of each party's promise record.
79. As a väljare, I want the party tiles in a fixed, neutral order (alphabetical, with Mitt parti only highlighted, not ranked), so that the site does not imply which party is best.
80. As a väljare, I want each party tile to link to the party's Löften, so that I can go from the summary to the details.

### Väljare: Löften

81. As a väljare, I want a Löften section under Partier (on each party page, plus an overview across parties) listing each party's 2026 Löften, so that I can see what each party promised for this mandate period without a separate top-menu item.
82. As a väljare, I want each Löfte to show the exact quote from the Valmanifest and a link to the original document (with page or anchor), so that I can check that the party really said it.
83. As a väljare, I want each Löfte to show its Löftesstatus (Röstat för, Röstat emot, Blandat or Inte prövat än), so that I see what the party did about it.
84. As a väljare, I want each Löftesstatus to link to the votes behind it, so that I can judge for myself.
85. As a väljare, I want each vote behind a Löftesstatus to explain in plain Swedish why voting Ja or Nej on that förslagspunkt counts as for or against the Löfte, so that the link between promise and vote is understandable.
86. As a väljare, I want to see "Inte prövat än" worded as a fact about the voting record so far, never as a failing of the party, so that untested promises are not read as broken ones.
87. As a väljare, I want a Löfte whose Bekräftade kopplingar all have the party's Partilinje Avstår or Delad shown as "Inte prövat än", with those votes listed on the Löfte page, so that abstentions are visible without being read as a vote for or against.
88. As a väljare, I want the site never to say "hållet" or "brutet", so that I get facts and draw my own conclusion.
89. As a väljare, I want a Metod page explaining how a Löfte is chosen, how a Bekräftad koppling is confirmed, how Partilinje is computed and how Löftesstatus follows, so that I can trust the method.
90. As a väljare, I want the Metod page to say plainly what the site cannot do (for example, that opposition parties cannot deliver on their own, and that only confirmed votes count), so that I do not over-read the statuses.
91. As a väljare, I want to see when a Löfte's status last changed, and why, so that I can follow developments.
92. As a väljare, I want to filter Löften by party, utskott and Löftesstatus, so that I can find the promises I care about.
93. As a väljare, I want to see each party's Valmanifest listed as a source (title, date, link, and whether it is a draft or final), so that I know where the Löften come from.
94. As a väljare, I want the 2022 promises available as a clearly labelled archive for 2022–2026, so that history is kept without being mixed with the current period.
95. As a väljare, I want the archive not to show a Löftesstatus that our data cannot support, so that gaps in our curation are not presented as facts about the parties.
96. As a väljare, I want the betänkande page and the motion page to show which Löften the decision touches, through Bekräftade kopplingar, so that I can connect decisions and motions to promises.
97. As a väljare, I want the party page to add its Löften summary and Valmanifest, so that the party page answers the whole question for that party.

### Väljare: curated motion links (moved from Phase 1)

98. As a väljare, I want Återkommande krav to include other parties' motions with the same demand when the Redaktör has confirmed them (for example Rickard Nordin's C motions 2022/23–2024/25), so that the whole history of a demand is in one place.
99. As a väljare, I want a skriftlig fråga or interpellation that names a motion only in prose linked to that motion once the Redaktör has confirmed it (for example fråga 2022/23:339 to 2017/18:3293), so that the government's answer appears in the motion's chain.

### Journalist (Phase 2)

100. As a journalist, I want every Löfte, Bekräftad koppling and Löftesstatus to have a stable URL, so that I can cite it.
101. As a journalist, I want the public read API to expose Löften, their sources, Bekräftade kopplingar and Löftesstatus as data, so that I can analyse it myself.
102. As a journalist, I want to see when each Bekräftad koppling was confirmed, so that I can understand how the record evolved.
103. As a journalist, I want the Data section to describe the Valmanifest source and the curation method, so that I can assess the dataset's limits.

### Redaktör (Phase 2)

104. As the Redaktör, I want to receive the proposed Löften for one party at a time in the bot chat, each with its quote and source, so that I can review a manageable list.
105. As the Redaktör, I want to strike, edit or approve proposed Löften by replying in chat, so that only Löften I accept go public.
106. As the Redaktör, I want only about 15–25 testable Löften per party, with vague statements already left out, so that review takes minutes, not hours.
107. As the Redaktör, I want a Veckobatch of at most about 20 Föreslagna kopplingar every Sunday after the weekly jobs, so that curation is a predictable small habit.
108. As the Redaktör, I want each batch item to be a numbered line with the Löfte, the votering, the proposed direction (whether Ja or Nej counts as for the Löfte) and the bot's reasoning, so that I can decide in about 10 seconds.
109. As the Redaktör, I want to answer in short form like "1,3,7 ja, resten nej", so that answering is fast.
110. As the Redaktör, I want to be able to say an item is right but the direction is reversed, so that a good link with a wrong direction is not lost.
111. As the Redaktör, I want items I did not answer to roll over to the next Veckobatch, so that nothing is lost when I skip a week.
112. As the Redaktör, I want weak suggestions never shown to me but logged, so that my time goes only to likely links.
113. As the Redaktör, I want to ask the bot to undo any earlier decision, such as an approval or a confirmed koppling, so that mistakes can be fixed.
114. As the Redaktör, I want to ask the bot what changed and when, so that I can review the history of my decisions.
115. As the Redaktör, I want nothing to become public until I have said yes, so that the site's credibility depends only on confirmed data.
116. As the Redaktör, I want there to be no admin page to maintain or secure, so that curation lives where I already work: the chat.

### Redaktör and bot: motion links (moved from Phase 1)

117. As the Redaktör, I want the bot to suggest links between a skriftlig fråga or interpellation and a motion when the question cites the motion only in prose, so that the follow-up chain is complete.
118. As the Redaktör, I want the bot to suggest Återkommande krav that are not identical in wording (for example another party's demand for the same change), so that the history of a demand is complete.
119. As the Redaktör, I want to answer these suggestions yes or no in chat as items in the same numbered Veckobatch as Föreslagna kopplingar, so that curation is one habit.
120. As the Redaktör, I want nothing curated to become public until I have said yes, so that the site's credibility depends only on confirmed data.
121. As the bot, I want to write the Redaktör's link decisions through the token-protected write API with an audit entry, so that every curated link can be traced and undone, using the same Ändringslogg as the Löfte writes.

### The bot (Phase 2)

122. As the bot, I want a write API protected by a token only I hold, so that I can record the Redaktör's decisions without touching the database by hand.
123. As the bot, I want every write to be recorded in an audit log with who, when, the old value and the new value, so that every change can be traced and undone.
124. As the bot, I want to record which chat answer each write came from, so that every change can be traced back to the Redaktör's words.
125. As the bot, I want an endpoint listing candidate voteringar that are new since the last Veckobatch, together with the approved Löften they might test, so that I can generate suggestions without scanning the whole record.
126. As the bot, I want to submit all my suggestions, including dropped low-confidence ones with their confidence and reasoning, so that dropped suggestions are logged for tuning.
127. As the bot, I want to create a Veckobatch from the strongest open suggestions and get back numbered items, so that the numbers I show in chat match what I write back.
128. As the bot, I want to submit the Redaktör's answers for a batch in one request, so that a whole answer is applied atomically.
129. As the bot, I want writes to be idempotent (repeating the same request has no extra effect), so that a retried request cannot create duplicates.
130. As the bot, I want the API to reject a koppling between a 2026 Löfte and a votering outside the 2026–2030 mandate period, so that a promise is never tested against votes from before it was made.
131. As the bot, I want to submit proposed Löften for a party's Valmanifest, with the exact quote and locator, and have them stored as not public, so that I can present them for approval.
132. As the bot, I want the API to reject a Löfte without a quote or source link, so that every public Löfte is checkable.
133. As the bot, I want to undo an audit-log entry through the API, which itself writes a new audit entry, so that undoing a change is just as traceable as making one.
134. As the bot, I want clear error responses (unauthorised, validation error, conflict), so that I can report problems to the Redaktör in chat.
135. As the bot, I want the write API to be unavailable when no token is configured, so that a misconfigured deployment fails closed.

### Operator (Phase 2)

136. As the operator, I want the write token stored only as a Kubernetes Secret and never logged, so that it cannot leak through config or logs.
137. As the operator, I want the retired keyword matcher and scorecard jobs removed from the daily chain, so that the scheduler does no work the site no longer uses.

## Implementation Decisions

### Phase 1: Motioner

#### Data source (verified against the live API on 2026-10-28)

**Motions.**

- **The motion list.** The document list endpoint with document type `mot`, filtered by riksmöte.
  - It also accepts a `subtyp` filter. For 2017/18 it returned 4,203 motions: 750 Kommittémotion, 68 Partimotion and 3,361 Enskild motion. The remaining 24 have none of these subtyps and are shown as Riksdagen gives them.
  - The value `Flerpartimotion` returns 0 hits for both 2017/18 and 2023/24. Multi-party motions are typed Enskild motion. For example, 2017/18:4202 is signed by C, KD and M.
  - Each row carries id, riksmöte, beteckning, subtyp, organ, title and subtitle, and also the signers (`dokintressent`) and a status line with the decision count (e.g. "1 avslag").
- **A motion's document status** returns these blocks: `dokument`, `dokforslag`, `dokintressent`, `dokreferens`, `dokaktivitet`, `dokuppgift` and `dokbilaga`.
  - **Document fields:** `subtyp` (Kommittémotion / Partimotion / Enskild motion), `organ` (the utskott), datum, titel, and the full HTML.
  - **Yrkanden** are in `dokforslag`, with these fields:
    - `nummer` and `lydelse`/`lydelse2` (the wording);
    - `utskottet` and `kammaren` (the utskott's proposal and the chamber's decision, e.g. "Avslag");
    - `behandlas_i` (e.g. "2017/18:SkU11");
    - `behandlas_i_punkt`.
  - **Authors** are in `dokintressent` (intressent_id, namn, partibet, ordning, roll).
  - **The betänkande link** is in `dokreferens` (referenstyp `behandlas_i`, ref_dok_id, ref_dok_rm, ref_dok_bet).
- **Gaps observed:**
  - `behandlas_i_punkt` was **empty** on every yrkande checked (2017/18:3293, 2018/19:2900, 2019/20:2718, 2021/22:3638, 2022/23:1675, 2024/25:989).
  - For 2018/19:2900 yrkande 26, `utskottet` and `kammaren` were also empty.

**Betänkanden.** A betänkande's document status returns these blocks:

- **`dokutskottsforslag`, one item per förslagspunkt**, with these fields:
  - punkt and rubrik;
  - `forslag`: HTML that names every motion and yrkande handled, e.g. "2021/22:3638 av Niklas Wykman m.fl. (M) yrkande 9";
  - `beslutstyp`: "röstning" or "acklamation";
  - `votering_id`;
  - `motforslag_nummer` and `motforslag_partier`: the reservation put to the vote, 0 if none;
  - `votering_sammanfattning_html`, with Ja/Nej/Avstår/Frånvarande per party;
  - `vinnare`.
- **`dokmotforslag`, one item per reservation or särskilt yttrande**: nummer, partier, typ ("reservation") and utskottsforslag_punkt.
- **`dokaktivitet`**: activity codes and dates, including beslut.
- The reservation text (the "ställningstagande" with its signers) and **the utskott's reasoning ("Utskottets ställningstagande") exist only in the document HTML**. They must be parsed by section heading.

**Voteringar.** The votering list endpoint, filtered by riksmöte, beteckning and punkt, returns per-party totals. The existing per-member ballot ingestion is kept.

**Skriftliga frågor and interpellationer.**

- **Document types:** `fr` (fråga), `frs` (svar på fråga) and `ip` (interpellation). For 2022/23 the API lists 979 fr, 977 frs and 432 ip.
- **The question and its answer** are linked both ways in `dokreferens` (referenstyp `frågesvar` from fr to frs, and `fråga` from frs to fr).
- **People** come from `dokintressent` with roles `fragestallare`, `stalldtill` and `besvaradav`. For example, HA12339's answerer is Finansminister Elisabeth Svantesson (M), and the svar is dated 2023-02-22.
- **No structured link from a question to a motion.** Fråga 2022/23:339 cites the 2017 M motion only in prose ("en kommittémotion från Moderaterna redan 2017"), without a beteckning.

**Speeches.** The anförande list endpoint by riksmöte returns these fields:

- dok_id (the protocol), dok_rm, dok_datum;
- avsnittsrubrik, kammaraktivitet;
- anforande_id, anforande_nummer;
- talare, parti, intressent_id;
- rel_dok_id, replik;
- anforande_url_html, protokoll_url_www.

For ärendedebatt, `rel_dok_id` is the betänkande's dok_id. For example, the 2022-09-21 debate speeches carry H901MJU33. `anforandetext` is empty in the list, so the text is fetched per anförande.

#### Modules

- **Motioner (a new deep module, inside or next to the votes module).** It owns motions, yrkanden, motion authors, betänkanden, förslagspunkter, reservations, the yrkande→förslagspunkt→votering link, frågor/svar/interpellationer, motion links and Återkommande krav.
  - **Public interface:**
    - list motions with filters;
    - one motion with its whole chain;
    - motions decided in a förslagspunkt;
    - motions, frågor and interpellationer by member;
    - Återkommande krav for a yrkande.
  - **Ingestion interface:** fetch and upsert per riksmöte. It is idempotent, keyed on dok_id, and records coverage.
- **Linker (a pure function inside Motioner).** It turns a yrkande's motion beteckning, yrkande number and `behandlas_i` betänkande into the förslagspunkt(s) that decided it.
  - It parses the förslagspunkt's `forslag` text for "rm:nr av … yrkande N", including the forms "yrkandena 1–3 och 5" and "yrkande 2 delvis".
  - It uses `behandlas_i_punkt` when present, and falls back to text parsing when it is empty, as observed.
  - It is a pure function with no I/O, tested with real recorded text.
- **Outcome (a pure function inside Motioner).**
  - **Per yrkande**, it maps the chamber decision to the plain-language outcome: Avslag → Avslagen, Bifall → Bifallen, partial bifall → Delvis bifallen, Besvarad → Besvarad, none → Ej avgjord än.
  - When `kammaren` is empty, it falls back to the förslagspunkt's decision ("avslår"/"bifaller" in the förslag text plus the decided status).
  - **Per motion**, it aggregates the yrkanden: all the same gives that outcome, mixed gives Delvis bifallen with counts.
  - The full value vocabulary of `kammaren` is collected during the backfill. Any unknown value is shown as Riksdagen's own word, never guessed.
- **Decision mode.** For each förslagspunkt, the decision is either a votering, using `votering_id` from Riksdagen joined to our votering by its internal id, or Avgjord utan omröstning (acklamation).
  - The votered alternative ("utskottets förslag mot reservation N (party)") is stored, so the motion page never implies that a party's own reservation was voted on when another was.
- **Återkommande krav.**
  - **Automatic grouping:** yrkanden whose normalised wording (whitespace, punctuation and case removed) is identical across riksmöten are grouped automatically. For example, M's 2017/18, 2018/19, 2019/20 and 2021/22 yrkanden have word-for-word the same lydelse.
  - **No curated grouping in Phase 1.** Non-identical demands, such as Nordin's C motions for "vinstmarginalbeskattning på digitala tjänster och produkter", are not grouped. Phase 2 adds Redaktör-confirmed grouping.
  - Each member records the rule that grouped it: identical wording.
- **Motion links (fråga/interpellation → motion), automatic only.**
  - A link is made when Riksdagen's `dokreferens` points from the question to the motion, or when the question's text contains a motion beteckning ("2017/18:3293", or "motion 3293" with the riksmöte known from the context).
  - Each link records the rule that made it.
  - Anything else is not linked in Phase 1. Fråga 2022/23:339 is therefore not linked to 2017/18:3293 until Phase 2 curation. It still appears on the asker's and the answering minister's member pages.
- **Origin repair (votes module).**
  - **The bug.** Today's origin step resolves one origin per betänkande from its first proposition or motion reference. It then writes that origin to every votering matching beteckning and förslagspunkt, ignoring riksmöte. It also attributes the whole point to the first author of the first motion.
  - **The fix.** It is replaced by a read of the Motioner link:
    - a votering's "proposed by" is the set of motions and yrkanden its förslagspunkt decided, plus the proposition if any;
    - the alternative put to the vote is the reservation in `motforslag_*`;
    - every lookup and write is keyed on riksmöte + beteckning + förslagspunkt.
  - The stored origin columns are dropped or recomputed from the link for every riksmöte, and the old single-motion tag is retired.
- **Speeches (existing ingestion, extended).**
  - **Backfill** every riksmöte for which voteringar exist, newest first, back to 2002/03, recording coverage per riksmöte so that the run can be stopped and resumed.
  - **Linking to a betänkande:** by `rel_dok_id` when it names a betänkande. Otherwise by matching `avsnittsrubrik` and the protocol date to the betänkande's debate date and title. If neither matches, the speech is left unlinked rather than guessed.
- **Questions (new ingestion).** fr, frs and ip per riksmöte, with question↔answer pairing from `dokreferens` and people from `dokintressent`. Interpellation debates are found among speeches by the interpellation's title and date.
- **Partilinje and Avvikare (votes module).** One shared definition, unchanged from Phase 2 below, used by Beslut, motion pages and the front page.
- **Beslut read models (votes module, extended).**
  - The list of betänkanden, decided or upcoming, with filters.
  - The canonical betänkande page keyed by riksmöte and beteckning: every förslagspunkt with Partilinje per party, Avvikare, decision mode, reservations, the utskott's reasoning, and the motions decided.
  - A front-page feed.
  - The decision date comes from Riksdagen's beslut activity and decided status, never from the record's system timestamp (which is unreliable).
- **No write API in Phase 1.** Phase 1 has no curation, so the write API, the Ändringslogg and manual confirmation all stay in Phase 2 as planned on 2026-10-27.
- **Frontend.**
  - Motion page, Motioner list, and an extended betänkande page.
  - Extended member and party pages.
  - Regeringen page (fifth top-menu item): the current government and ministers, plus the budget, the Riksdag overview and the agencies as sections. The budget area detail stays as a sub-page of Regeringen.
  - New nav, front page variant A with "Senaste motionsutfall", and Mitt parti in localStorage.

#### Schema changes, Phase 1 (Postgres only)

- **Motion:**
  - dok_id (primary key), riksmöte, beteckning (unique with riksmöte), titel, subtyp (as given), derived kind (adds Flerpartimotion), utskott, datum, status, source URL;
  - motivering text, stored for search and excerpt;
  - fetched-at.
- **Motion author:** motion, intressent_id, name, party at the time, ordning, roll.
- **Yrkande:**
  - motion, nummer, lydelse;
  - normalised-wording hash (for Återkommande krav);
  - the utskott's proposal and the chamber's decision, both as given by Riksdagen;
  - behandlas_i (riksmöte + beteckning);
  - outcome (derived, stored for filtering, recomputed on ingest).
- **Betänkande:** dok_id, riksmöte, beteckning, titel, utskott, debate date, beslut date, status, source URL.
- **Förslagspunkt:**
  - betänkande, punkt, rubrik, förslag text;
  - beslutstyp (votering / acklamation);
  - votering (FK to the votering's internal id, nullable);
  - votered alternative (reservation nummer and parties);
  - utskottets ställningstagande text.
- **Reservation:** betänkande, nummer, förslagspunkt, typ (reservation / särskilt yttrande), parties, signers, text.
- **Yrkande decision:**
  - yrkande, förslagspunkt;
  - link origin: behandlas_i_punkt or parsed text.
  - Unique per (yrkande, förslagspunkt).
- **Fråga, svar, interpellation:**
  - dok_id, typ, riksmöte, nummer, titel;
  - the asker, and the minister asked or answering;
  - datum, text;
  - the paired answer.
- **Motion link:** source document, motion, rule (riksdagen-reference / cited-beteckning), created-at. Phase 2 extends the rule with suggested / confirmed / rejected.
- **Återkommande krav:** krav group id, member yrkande, rule (identical-wording). Phase 2 extends it the same way.
- **Ingestion coverage:** extended with the new sources per riksmöte.

#### Redirects, Phase 1

These are client-side, following the existing goals-redirect pattern. Share URLs keep working.

| Old URL | New place |
|---|---|
| Votes list | Beslut list |
| A votes detail page (beteckning + förslagspunkt) | The canonical betänkande page, anchored to that förslagspunkt, resolved to the votering's riksmöte |
| The old betänkande URL without riksmöte | The canonical page for the most recent riksmöte with that beteckning |
| Budget overview and budget area pages | The Regeringen page's budget section and its area sub-pages |
| Riksdag overview | The Regeringen page's Riksdag section |
| Regering | Regeringen |

- The vote share URL is kept.
- Agency pages keep their URLs and are a section of the Regeringen page.
- Kommun and region pages keep their URLs.
- Manifestos stays reachable (not in the nav) until Phase 2.

#### Cross-links, Phase 1

- Each motion links to its authors, their party, its betänkande and förslagspunkt, its votering, its debate speeches, the frågor and interpellationer about it, and its Återkommande krav.
- Each votering links to its betänkande, its motions, and each party and member.
- Each betänkande links to its debate and speeches, and to each motion it decided.
- Each member links to their ballots, speeches, motions, frågor and interpellationer.
- Each party links to its votes and its motions.

#### Public read API, Phase 1 (additions to the OpenAPI contract, which stays the single source of truth)

- **List motions.** Filters: party, utskott, outcome, riksmöte, subtyp and free text.
- **One motion** by riksmöte + beteckning, with the whole chain.
- **Återkommande krav** for a yrkande.
- **Motions decided in a förslagspunkt.**
- **Frågor and interpellationer**, by member or by motion.
- **Beslut list and betänkande page**, now including decision mode, reservations and motions.
- **Front-page feed:** 5 recent decided betänkanden, recent motion outcomes, and 3 upcoming.

#### Data-source discipline, Phase 1

Motions, betänkanden, frågor/svar/interpellationer and the speech backfill are registered data sources, with a data-source page, source markers and an architecture diagram update, following the repo's existing rule for data sources.

### Phase 2: Löften

The Löfte decisions of 2026-10-27 are kept as they were, including the write API, the Ändringslogg and the Veckobatch. The only edits mark what Phase 1 now delivers first (Partilinje, the Beslut read models, the navigation, the front page and most redirects), place Löften under Partier, and add curation of the motion links Phase 1 could not make automatically.

#### Modules

- **Löften (new deep module, replaces the goals and matching features).** It owns Valmanifest, Löfte, Koppling, Veckobatch and the audit log. Its editorial interface also decides motion links and Återkommande krav memberships in the Motioner module, through the same audited write path.
  - Public interface: read Löften (filtered by party, election, utskott and status), read one Löfte with its Bekräftade kopplingar, read party summaries, and read Löften touched by a votering.
  - Editorial interface: one set of write operations, used only by the write API.
  - Every write goes through one internal path that applies the change and writes the audit entry in the same transaction, so a change without an audit entry is impossible by construction.
- **Löftesstatus (pure domain function inside Löften).**
  - Input: a list of (koppling direction, Partilinje) pairs.
  - Output: one of the four statuses plus the counts behind it.
  - It has no I/O and is the only place the rule lives. Read queries compute status on read; there is no materialised status column and no refresh job.
- **Partilinje (votes module; built in Phase 1, reused here).**
  - It is one shared definition of a party's position in a votering, used by the decision pages, the front page, Avvikare and Löftesstatus.
  - It replaces today's separate "party position" calculations, so the site cannot state two different party positions for the same votering.
- **Beslut (votes module; built in Phase 1, Phase 2 adds the Löfte count).**
  - Read models: the list of betänkanden (decided or upcoming, with filters), the canonical betänkande page keyed by riksmöte and beteckning (all förslagspunkter, Partilinje per party, Avvikare, the number of Löften touched), and a front-page "recent decided" and "upcoming" feed.
  - The decision date comes from Riksdagen's beslutsdag and decided status, never from the record's system timestamp.
- **Redaktion write API (new HTTP adapter in front of Löften's editorial interface).** It is mounted under its own path prefix with its own authentication middleware.
- **Frontend.**
  - Löften pages (list, Löfte detail, sources and the 2022 archive), a Metod page, and the Löften additions to the front page (variant A, built in Phase 1), betänkande, motion and party pages.
  - The navigation, Mitt parti picker and most redirects ship in Phase 1. Phase 2 adds Löften as a section under Partier (no top-menu item) and the Löfte redirects.
  - One shared renderer for a Löfte, replacing the goal card, used on every surface that shows a Löfte.
- **Retired.**
  - The keyword-matcher worker, the refresh-scorecards worker and the party scorecards materialised view.
  - The goal↔vote match table: its 8,453 unverified rows are not migrated as kopplingar.
  - The speech-extracted per-politician "promises" feature and its tables, which are empty and collide with the word Löfte.
  - The stub AI port in the backend. The backend never calls an AI provider; the bot is the AI component.

#### Schema changes, Phase 2 (Postgres only)

- **Valmanifest:**
  - id, party (FK parties), election year, title, source URL, publication date, state (draft / published), retrieved-at.
  - This replaces the hard-coded manifesto list in the frontend, so that every Valmanifest is data in Postgres.
- **Löfte:**
  - id, party, Valmanifest (FK), election year, short text (the testable claim, phrased close to the source), exact quote, source locator (page number or anchor), and source URL. The quote and source URL are required for any approved Löfte.
  - Relevant utskott codes. These are used only to narrow candidate voteringar and are never displayed as a ranking.
  - State: proposed / approved / struck / archived. Created-at and approved-at.
  - Only approved Löften (and archived 2022 ones) are public.
  - The 2022 party goals are migrated into this table with state "archived" and **their existing ids kept**, so that old goal URLs can redirect one-to-one. After that the old goals table is dropped.
- **Koppling:**
  - id, Löfte (FK), votering (FK to voteringar by its internal id, **not** beteckning + förslagspunkt, because two voteringar can share those).
  - Direction: which ballot, Ja or Nej, is in the Löfte's direction. This is the decisive field: the same Löfte can be tested by a förslagspunkt where Ja supports it or one where Nej supports it, for example a vote to reject a motion.
  - State: föreslagen / bekräftad / avvisad / bortsorterad (dropped for low confidence).
  - Confidence (0–1), the bot's reasoning, origin (who suggested it), Veckobatch (FK, nullable), decided-at.
  - Unique per (Löfte, votering).
  - Only bekräftad kopplingar are exposed by the public API.
- **Veckobatch:** id, created-at, closed-at, and ordered items (batch, item number, koppling). An item that gets no answer is released when the batch closes and is eligible for the next one.
- **Ändringslogg (audit):**
  - id, timestamp, actor (the token identity, currently the bot), on-behalf-of (the Redaktör), reference (for example the chat message the decision came from), entity type, entity id, action, old value and new value (both as JSON), and reverts (FK to the entry it undoes, nullable).
  - It is append-only. The application never updates or deletes rows, and the database role used by the app is granted insert and select only on this table.
- **Motion link and Återkommande krav:** the Phase 1 tables gain the states suggested / confirmed / rejected, plus reasoning, Veckobatch (FK, nullable) and decided-at. Only automatic and confirmed rows are public.
- Migrations follow the existing numbering. Down-migrations exist but are not expected to be used in production.

#### Löftesstatus computation

- **Partilinje for a party in a votering:**
  - Count the party's ballots cast (Ja, Nej, Avstår), with Frånvarande excluded, and take the strict plurality.
  - A tie gives "Delad".
  - This matches what Om sajten already promises ("majority of cast ballots, absences excluded").
- **Direction of a Bekräftad koppling for a party:**
  - If Partilinje equals the koppling's direction, the koppling counts as **för**.
  - If it is the opposite of Ja / Nej, it counts as **emot**.
  - If Partilinje is Avstår or Delad, the koppling has no direction for that party.
- **Löftesstatus over the Löfte's party's directional Bekräftade kopplingar:**
  - All för: **Röstat för**.
  - All emot: **Röstat emot**.
  - Both: **Blandat**.
  - None, including when every Bekräftad koppling has Partilinje Avstår or Delad: **Inte prövat än**. Those votes are still listed on the Löfte page.
- **Avvikare:** a member whose cast ballot differs from a Ja or Nej Partilinje. Avvikare are shown on the vote page and never affect Löftesstatus.
- **Time window:** only voteringar inside the Löfte's mandate period count, which is 2026–2030 for 2026 Löften.
- **Archive:** archived 2022 Löften are shown with their source but **without** a Löftesstatus, because they have no Bekräftade kopplingar. "Inte prövat än" would wrongly state our curation gap as a fact about the party.

#### Public read API, Phase 2 (additions to the OpenAPI contract)

- **List Löften.** Filters: party, election (default 2026), utskott and status. Each item has the Löfte, its source fields, its Löftesstatus with counts, and when the status last changed.
- **One Löfte.** The fields above plus every Bekräftad koppling: the votering, its betänkande, its förslagspunkt, the direction and a plain-language explanation, the Partilinje of the Löfte's party, the confirmed-at date, and the source link.
- **Party summary.** Counts per Löftesstatus for one or all parties, always returned in alphabetical party order.
- **Valmanifest sources:** a list per party.
- **Beslut list, betänkande page and motion page.** Phase 2 adds the Löften touched (only through Bekräftade kopplingar) to the Phase 1 responses.
- **Front-page feed:** Phase 2 adds the number of Löften touched per decided betänkande and the party Löftesstatus summaries.
- **Retired endpoints:** the party-goals endpoints and the goal↔vote endpoint go away once the frontend no longer uses them. Each answers with a pointer to its replacement for one release.

#### Write API (Redaktion)

- **Authentication:**
  - A bearer token compared in constant time against a value from a Kubernetes Secret.
  - If no token is configured, every write route answers 404, so a misconfigured deployment fails closed.
  - These routes are not covered by the browser CORS allowlist, have their own rate limit, and never log the token or request bodies.
- **Every write request:**
  - Carries an on-behalf-of field and a reference, which go into the audit entry.
  - Accepts an idempotency key, so a retry returns the original result.
  - Is atomic: the change and its audit entries land together, or neither does.
- **Operations:**
  - **Upsert Valmanifest:** party, year, title, URL, state.
  - **Propose Löften:** a batch for one Valmanifest, stored as proposed. Validation rejects any Löfte without a quote, locator or URL.
  - **Edit, approve or strike a Löfte:** one or many by id. Approving makes the Löfte public.
  - **Candidates:**
    - Returns decided voteringar new since the last closed Veckobatch, inside the current mandate period, each paired with the approved Löften whose utskott matches the votering's utskott.
    - Includes the betänkande title, förslagspunkt text and the motions/yrkanden decided there, so the bot can judge without other calls.
    - This is a read endpoint, but it sits behind the same token.
  - **Submit suggestions:** a list of (Löfte, votering, direction, confidence, reasoning). Suggestions below the confidence threshold are stored as bortsorterad, and the threshold is part of the request so it can be tuned.
  - **Create Veckobatch:** picks up to N (default 20) open föreslagna kopplingar, highest confidence first, and returns them numbered.
  - **Decide a Veckobatch:** a list of (item number, ja / nej / ja-reversed). "Ja" makes the koppling bekräftad, "nej" makes it avvisad, and "ja-reversed" makes it bekräftad with the direction flipped. Items not listed stay open and roll over.
  - **Submit and decide motion-link and Återkommande krav suggestions:** the bot submits suggestions with reasoning; they join the same Veckobatch as numbered items and are decided ja / nej.
  - **Read the Ändringslogg:** filter by entity and time.
  - **Undo an entry:** applies that entry's old value as a new change, with a new audit entry that points back to the one it undoes. Undo is refused if the entity has changed since, which returns a conflict.
- **Error responses:** 401 for a missing or wrong token, 422 for a validation error, 409 for a conflict, and 404 for an unknown entity.

#### Extraction and suggestion pipeline

- **Löfte extraction (once per party, then whenever a Valmanifest changes):**
  1. The bot fetches the Valmanifest and extracts concrete, testable statements. A statement is testable if some riksdag vote could in principle support or oppose it.
  2. It keeps 15–25 per party and skips vague statements.
  3. It checks that each quote appears word for word in the source, and stores the page or anchor.
  4. It submits them as proposed.
  5. It presents the list to the Redaktör in chat, one party at a time.
  6. It applies strikes, edits and approvals through the write API.
- **Weekly suggestions (Sunday, after the weekly jobs in the daily chain):**
  1. A routine of the bot calls the candidates endpoint.
  2. For each candidate pair it judges relevance, direction and confidence and writes a one-line reasoning.
  3. It submits everything, and weak suggestions are stored as bortsorterad.
  4. It also suggests motion links for questions that name a motion only in prose, and Återkommande krav with different wording.
  5. It creates a Veckobatch of 20 or fewer and posts the numbered lines in chat.
  6. It parses the Redaktör's answer and submits the decisions.
  7. It confirms back in chat what was written.
- **Backend AI:** the backend contains no AI call and no AI key. This keeps the repo rule that the stub AI service is not wired to real calls without an explicit ask, and it keeps the only credential on the bot side.

#### Redirects, Phase 2

The Phase 1 redirects stay; Phase 2 adds:

| Old URL | New place |
|---|---|
| Manifestos | Löften sources |
| A party goal's votes page | The Löfte detail page (same id, archived) |
| A party goals page | That party's Löften section under Partier |


#### Cross-links, Phase 2

- Each party links to its Löften.
- Each betänkande and each motion links to the Löften it touches.
- Each Löfte links to its Valmanifest and to every Bekräftad koppling.

#### Data-source discipline, Phase 2

The Valmanifest catalogue becomes a registered data source with a data-source page, source markers and an architecture diagram update, following the repo's existing rule for data sources.

## Testing Decisions

- **What makes a good test here:**
  - It exercises external behaviour: HTTP requests in, and JSON or pages out.
  - It runs against real Postgres with real migrations.
  - It never asserts on internal structure, SQL text or how calls are made between modules.
  - Test names use the glossary terms.
- **Seam 1 (primary, and ideally the only backend seam): the HTTP API against a real migrated Postgres.**
  - The router is built by a constructor that tests can call with a database pool, a token, and the Riksdagen base URL. Today the router is assembled inline in the API's main function, so this is the one new seam needed.
  - **Ingestion is tested through the same seam.** The Riksdagen client already takes a configurable base URL, and its existing tests use an HTTP test server.
    - The ingestion tests point it at an HTTP test server that replays responses **recorded from the live API**.
    - They run ingestion and then assert through the public read API, never on tables.
  - **Recorded fixtures (the worked example):**
    - motions 2017/18:3293, 2018/19:2900, 2019/20:2718, 2021/22:3638, 2022/23:1675 and Nordin's other two C motions;
    - their betänkanden (2017/18:SkU11, 2018/19:SkU11, 2019/20:SkU21, 2021/22:SkU16, 2022/23:SkU15, and SkU14 for 2023/24 and 2024/25);
    - the matching voteringar;
    - fråga and svar 2022/23:339;
    - one debate's anföranden.
  - **What Phase 1 covers:**
    - **The chain for 2017/18:3293 y1:** SkU11 förslagspunkt 6, Avgjord utan omröstning, reservations 8 (M) and 9 (KD), and the outcome Avslagen.
    - **2021/22:3638 y9:** decided by votering at SkU16 p6, where the alternative put to the vote is reservation 12.
    - **2018/19:2900 y26:** the motion page shows that the votered alternative was another party's reservation, not M's own.
    - **Linker forms:** "yrkande N", "yrkandena 1–3 och 5", "yrkande N delvis", and an empty `behandlas_i_punkt` falling back to the text. A motion missing from every förslagspunkt yields "Ej avgjord än", not a guess.
    - **Outcome mapping** for each chamber value seen, with an unknown value shown verbatim.
    - **Återkommande krav:** the four identical M yrkanden are grouped automatically, and Nordin's C yrkanden are not.
    - **Motion links:** a question citing a beteckning is linked; fråga 2022/23:339, which cites no beteckning, is not linked to 2017/18:3293 but appears on Nordin's and Svantesson's member pages.
    - **Origin regression:** SkU16 p6 in 2021/22 and in 2025/26 get different, correct motions. No votering shows a motion from another riksmöte.
    - **Idempotency:** ingesting the same riksmöte twice changes nothing, and coverage records each source.
    - **Motioner list filters:** party, utskott, outcome, riksmöte and kind, including derived Flerpartimotion.
    - **Speech linking:** rel_dok_id, then the fallback, then unlinked, including a 2002/03 debate.
  - **What Phase 2 covers (as decided 2026-10-27):**
    - **Löftesstatus:** all four statuses, Avstår / Delad Partilinje (a Löfte whose Bekräftade kopplingar are all Avstår or Delad gives Inte prövat än), Frånvarande excluded, the time window, the reversed direction, and the archive showing no status.
    - **Partilinje and Avvikare.**
    - **Write API:**
      - authentication: no token, wrong token, and an unconfigured token giving 404;
      - validation: missing quote or URL, and a votering outside the mandate period;
      - idempotency, batch numbering and rollover;
      - atomicity: no change without an audit entry;
      - undo, including the conflict case.
    - **Visibility:** proposed Löften and föreslagna kopplingar never appear in the public responses.
    - **Curated motion links:** after a confirmed link, fråga 2022/23:339 appears on the 2017/18:3293 page; Nordin's C yrkanden join M's Återkommande krav only after confirmation and leave again after undo; rejected or suggested links never appear.
- **Seam 2 (only for what the API cannot show): the Playwright end-to-end suite against a running stack.** It covers:
  - every redirect, including Budget, Riksdag and Regering into Regeringen, and votes into Beslut;
  - the five-item top nav, Lokalt as a side menu, and Om sajten and Data in the secondary place;
  - in Phase 2, Löften reached under Partier, with no top-menu item;
  - front-page clickability: each row, dot, tag and tile has a link target;
  - the motion page's chain sections rendering, with each step linking out;
  - Mitt parti persisting in localStorage across reloads and never being sent to the backend;
  - light and dark rendering of the new pages.
- **Prior art:**
  - **Postgres adapter integration tests** gated on the test database URL environment variable (goals, committees, votes and the scorecard view).
  - **HTTP handler tests** using an HTTP test server and hand-written fakes, with no mocking library (committee votes handler, ingestion handler).
  - **The Riksdagen, SCB and Kolada client tests** using an HTTP test server with a configurable base URL.
  - **The existing end-to-end specs.**
  - The new seam combines these: a real database behind real handlers, fed by recorded Riksdagen responses.
- **CI (prerequisite).** CI currently only builds and does not run the Go tests. The separate CI fix (audit fix 7) adds a Postgres service and runs the Go tests, including seam 1, on every PR. The end-to-end suite runs against the dev stack.
- **Pure functions** (Linker, Outcome, Löftesstatus) may get table-driven unit tests as documentation of the rules. Seam 1 remains the source of truth.

## Out of Scope

### Phase 1

- **Claims about consequences.** The site does not say that a decision "cost tax revenue", "hurt a company" or "made the wrong call". It shows the demand, who decided, how, and what was said, linking to the sources. Commentary of that kind belongs to readers and journalists.
- **Propositioner, SOU reports and remiss documents** beyond what is needed to show a förslagspunkt's own text.
- **Följdmotioner as their own concept.** They are motions like any other.
- **Success-rate statistics or rankings** per party or member (for example, "M got 3% of its motions bifallna").
- **Manual or similarity-based links.** Phase 1 makes no link that is not certain (Riksdagen reference, cited beteckning or identical wording). Everything else waits for Phase 2 curation.
- **The write API, Ändringslogg and any curation.** These are Phase 2.
- **Full-text search** across motion bodies. Free-text search covers title and yrkande wording only.
- **Lokalt features.** Lokalt gets maintenance only.

### Phase 2 (as decided 2026-10-27)

- Löften from a government agreement or a government statement. There is only a caretaker government now, so this comes once a government is formed.
- Email or RSS alerts, and any account system.
- The promise-first front page (variant B). It is revisited once about 50 2026 Löften have Bekräftade kopplingar.
- Any admin UI, moderation UI or web login.
- Data files in the repo for Löften, kopplingar, motion links or Återkommande krav. All of it lives in Postgres.
- Verdicts such as "hållet / brutet / uppfyllt", percentages, scores or rankings across parties.
- Per-member promises extracted from speeches.
- Backfilling Löftesstatus for the 2022 archive.

### Both phases

- The quick-wins PR: the clickable feed, nav promotion, the brand name and the committee lookup fix. That is a prerequisite and is specified separately.

## Further Notes

- **This draft supersedes the 2026-10-27 draft** on the following points:
  - **Order:** Motioner first, then Löften.
  - **Top nav:** Beslut · Motioner · Ledamöter · Partier · Regeringen. In Phase 2, Löften become a section under Partier, not a top-menu item.
  - **One Regeringen page** absorbs Budget, Riksdag, Regering and the agencies.
  - **Lokalt** is a side menu, and Om sajten and Data are secondary.
  - **Redirects:** Riksdag overview → Regeringen, not Beslut. Agency pages are a section of Regeringen, not Lokalt.
  - **Front page:** the Löfte elements ("berör N löften" and the party Löftesstatus tiles) move to Phase 2. Phase 1 shows motion outcomes instead.
  - **Speech backfill** goes back to 2002/03, the oldest riksmöte with voteringar.
  - **Löftesstatus** is explicitly Inte prövat än when every Bekräftad koppling has Partilinje Avstår or Delad.
  - **Unchanged:** the write API, the Ändringslogg and all manual confirmation stay in Phase 2. Phase 1 links only automatically, so ADR-003 stands as written.
- **Supersedes ADR-001 (Concern A).** ADR-001 put party goals in SQL migrations. From now on, editorial data is written to Postgres through the write API. Migrations only create the schema and move the 2022 goals into the archive once.
- **Issue #82 (the utskott spine Wayfinder map) is closed as not planned.** This spec replaces its direction: utskott become a filter on Beslut and on Motioner, and Motioner and Beslut lead.
- **Issues that need a fresh decision against this spec** (left open and untouched):
  - **#100, no vote count beside a goal.** Phase 2 adds "berör N löften" beside a decision and status counts on party tiles. Both count only Bekräftade kopplingar, never keyword matches. Decide whether #100 still holds in that form.
  - **#101, the second navigation tier (Kommun · Region · Regering · Riksdag).** This spec replaces it with the five-item top menu and Lokalt as a side menu. Decide whether to close or rewrite #101.
  - **#89, "RÖSTAT does not judge whether a vote kept a promise".** This spec agrees: Löftesstatus reports votes, never kept or broken. Phase 1 also fixes the origin bug that made the retired alignment percentage wrong. Decide whether #89 is now satisfied or needs new wording.
- **No derived comparison across parties.** The Phase 1 Motioner list shows outcomes per motion, never a rate per party. The Phase 2 tiles keep alphabetical order, no percentages and no sorting by status.
- **Framing check.** The Problem section speaks of what decisions mean for people (the moms example). The site itself still states only what was voted and decided, never a verdict; the consequence lives in the motion's and betänkande's own text, which is quoted and linked.
- **Audit reference for the origin bug, read at commit 29b65f7.**
  - **The bug is in the votes service's origin step and the repository's origin write:**
    - `internal/votes/service.go:94-139` resolves one origin per betänkande from the first reference;
    - `internal/votes/repository.go:182-190` updates every votering with that beteckning and förslagspunkt, with no riksmöte filter.
  - **Production evidence:**
    - every SkU16 p6 votering, across riksmöten, carries the same motion H9021211;
    - 2024/25 SkU14 p3 carries an unrelated SD "Punktskatt och tull" origin;
    - 2022/23 SkU15 p4 has no origin at all.
- **Speech coverage today:** 2025-05 to 2025-09 and 2026-08 to 2026-10 only. The worked example's debates (2018-03-28 and later) are therefore missing until the backfill, which now runs back to 2002/03.
- **Open questions for Alexander** (resolved on 2026-10-28: Phase 1 curation, Regeringen's place, Löften's place, backfill depth, and Avstår/Delad; resolved on 2026-10-30: the betänkande URL format):
  1. **Outcome words.** For acklamation: "Avslagen (utan omröstning)" or "Avslagen" plus a separate "Avgjord utan omröstning" tag? For Besvarad, which means the utskott held the demand already met or under way, is "Besvarad" plain enough, or should it be explained?
  2. **Phase 2 carry-over:** the Veckobatch size of 20 and the confidence threshold of 0.6.
  3. **The canonical betänkande URL format** (resolved 2026-10-30): the riksmöte is written "2026/27", as Riksdagen writes it, never "2026-27". Where it must sit in a single path segment or a query parameter it is URL-encoded ("2026%2F27").
- **Repo documentation that contradicts the decisions.** The docs PR for this spec fixes CLAUDE.md's "promises kept or broken", its database description (Neon / Postgres 17, while production is self-hosted Postgres 18 on the VPS) and its navigation. Still to update in the build:
  - CLAUDE.md lists "keyword-matcher" and "refresh-scorecards" as scheduled workers.
  - CLAUDE.md describes the stub AI service as the future home of AI calls.
  - Om sajten's sections on goals.
- **The repo's domain-doc config** (docs/agents/domain.md) names CONTEXT.md at the root and a docs/adr folder. The glossary is therefore added as CONTEXT.md. The ADRs follow the existing root adr folder (numbered 001); domain.md's docs/adr path is a known mismatch to fix separately.
- **File paths and line references** appear only in these notes, as audit references. The spec body stays path-free, following the to-spec convention.
