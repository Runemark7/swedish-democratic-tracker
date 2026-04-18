# RIKSDAGSKOLLEN - Project Specification

## What This Is
A civic tech tool that tracks Swedish politicians' stated promises and party goals against their actual voting records in Riksdagen. The core value proposition: show the gap between what politicians SAY and what they DO, with full context on who initiated each vote.

## Two Main Views
1. **Party Goals View**: Extract goals from valmanifest/partiprogram, match against collective party voting behavior, show alignment percentage per goal
2. **Individual Politician View**: Extract promises from debate speeches, cross-reference with their personal voting record, flag contradictions

## Critical UX Requirement
Every vote displayed MUST show:
- **Which party initiated the proposal** (proposed_by_party)
- **Proposal type**: Proposition (government bill) vs Motion (party initiative)
- **Context note**: Brief explanation of why a party may have voted against something that superficially aligns with their goal (e.g., "S voted NEJ because the M-proposal bundled housing investment with market rent conditions S opposes")

This prevents misinterpretation. A NEJ vote on a healthcare bill doesn't mean anti-healthcare -- the bill may have been proposed by an opposing party with unacceptable conditions attached.

---

## Tech Stack
- **Backend**: Go (REST API)
- **Database**: PostgreSQL
- **Ingestion/Orchestration**: n8n workflows (periodic scraping + processing)
- **AI Layer**: Claude API (promise extraction, goal extraction, vote-relevance scoring)
- **Frontend**: React

---

## Data Source: Riksdagen Open API

Base URL: `https://data.riksdagen.se`
Format: JSON (append `&utformat=json` to all requests)
Auth: None required. Free to use, cite "Sveriges riksdag" as source.

### Key Endpoints

**1. Members (ledamöter)**
```
GET /personlista/?rdlstatus=tjanstgorande&parti={S|M|SD|...}&utformat=json
```
Returns: intressent_id, tilltalsnamn, efternamn, parti, valkrets, bild_url_192
- `intressent_id` is the unique person ID used across all endpoints

**2. Debate Speeches (anföranden)**
```
GET /anforandelista/?rm=2024/25&parti=S&iid={intressent_id}&sz=20&anftyp=Akt&utformat=json
```
Returns: dok_id, anforande_nummer, talare, parti, intressent_id, dok_datum, avsnittsrubrik, anforandetext, rel_dok_id
- `anforandetext` = full speech text (HTML)
- `avsnittsrubrik` = topic heading
- `rel_dok_id` = links to the document being debated
- `anftyp=Akt` filters to actual speeches (not replies)
- Data available from riksmöte 1993/94 onward

**3. Voting Records (voteringar)**
```
GET /voteringlista/?rm=2024/25&parti=S&iid={intressent_id}&bet={AU1}&sz=500&utformat=json
```
Returns per individual vote: votering_id, intressent_id, namn, parti, rost (Ja|Nej|Avstår|Frånvarande), beteckning, forslagspunkt, rm, dok_id
- `beteckning` = committee report ID (e.g., "SoU12" = Socialutskottet report 12)
- `forslagspunkt` = which proposal point within that report
- Can filter by: rm (session), parti, iid (person), bet (committee report), valkrets
- Can group by: ledamot-id, ledamot-namn, parti, valkrets, votering_id

**4. Document Details (dokumentstatus)**
```
GET /dokumentstatus/{dok_id}.json
```
Returns full metadata including:
- `dokument`: title, type, date, status
- `dokintressent.intressent[]`: who authored it (party via `partibet` field)
- `dokreferens.referens[]`: related documents (trace betänkande -> underlying prop/mot)
  - `ref_dok_typ`: "prop" or "mot"
  - `ref_dok_id`: ID of the underlying proposal

**5. Document Search**
```
GET /dokumentlista/?rm=2024/25&doktyp=prop&sz=20&sort=datum&sortorder=desc&utformat=json
```
- `doktyp`: prop (propositions), mot (motions), bet (committee reports), ip (interpellations), fr (written questions)

**6. Datasets (bulk download)**
Available at riksdagen.se/sv/dokument-och-lagar/riksdagens-oppna-data/
- CSV/SQL dumps per riksmöte for anföranden and voteringar
- Useful for initial bulk load instead of paginated API calls

### Tracing Vote Origin (Critical Flow)
To find who proposed what a party voted on:
1. Start with a vote's `beteckning` (e.g., "SoU12") 
2. Construct dok_id: `{rm_without_slash}:{beteckning}` (e.g., "202425:SoU12")
3. Fetch `/dokumentstatus/{dok_id}.json`
4. Look at `dokreferens.referens[]` for the underlying proposals
5. If `ref_dok_typ` = "prop" → government proposition (proposed_by = "Regeringen")
6. If `ref_dok_typ` = "mot" → party motion → fetch that motion's `dokumentstatus` → read `dokintressent.intressent[0].partibet` to get the proposing party

### Topic-to-Committee Mapping
Riksdagen committees (utskott) and their beteckning prefixes:
- SoU = Socialutskottet (healthcare, elderly care)
- FöU = Försvarsutskottet (defense, NATO)
- SfU = Socialförsäkringsutskottet (migration, social insurance)
- SkU = Skatteutskottet (taxes)
- FiU = Finansutskottet (budget, finance)
- UbU = Utbildningsutskottet (education)
- MJU = Miljö- och jordbruksutskottet (climate, environment)
- NU = Näringsutskottet (energy, business)
- JuU = Justitieutskottet (crime, police, justice)
- CU = Civilutskottet (housing)
- AU = Arbetsmarknadsutskottet (labor market)
- TU = Trafikutskottet (transport)
- KU = Konstitutionsutskottet (constitutional matters)
- UU = Utrikesutskottet (foreign affairs)

---

## Database Schema (PostgreSQL)

```sql
-- Core entities
CREATE TABLE politicians (
  intressent_id TEXT PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  party TEXT NOT NULL,
  constituency TEXT,
  image_url TEXT,
  is_active BOOLEAN DEFAULT true,
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE speeches (
  id SERIAL PRIMARY KEY,
  dok_id TEXT NOT NULL,
  anforande_nummer TEXT,
  politician_id TEXT REFERENCES politicians(intressent_id),
  party TEXT NOT NULL,
  date DATE NOT NULL,
  topic_heading TEXT,
  speech_text TEXT,
  related_dok_id TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE votes (
  id SERIAL PRIMARY KEY,
  votering_id TEXT NOT NULL,
  politician_id TEXT REFERENCES politicians(intressent_id),
  party TEXT NOT NULL,
  vote_result TEXT NOT NULL, -- Ja, Nej, Avstår, Frånvarande
  beteckning TEXT NOT NULL,  -- committee report ID
  forslagspunkt TEXT,
  session TEXT NOT NULL,      -- riksmöte e.g. "2024/25"
  dok_id TEXT,
  -- Proposal origin (enriched)
  proposed_by_party TEXT,     -- which party initiated
  proposal_type TEXT,         -- prop, mot, bet
  proposal_dok_id TEXT,
  document_title TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE(votering_id, politician_id)
);

-- AI-extracted data
CREATE TABLE party_goals (
  id SERIAL PRIMARY KEY,
  party TEXT NOT NULL,
  goal_text TEXT NOT NULL,
  topic TEXT NOT NULL,
  specificity TEXT NOT NULL,  -- concrete, directional, rhetorical
  source_document TEXT,       -- "Valmanifest 2022", "Tidöavtalet"
  keywords TEXT[],
  relevant_committees TEXT[],
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE promises (
  id SERIAL PRIMARY KEY,
  politician_id TEXT REFERENCES politicians(intressent_id),
  speech_id INTEGER REFERENCES speeches(id),
  promise_text TEXT NOT NULL,
  topic TEXT NOT NULL,
  specificity TEXT NOT NULL,
  keywords TEXT[],
  extracted_at TIMESTAMPTZ DEFAULT now()
);

-- Cross-reference results
CREATE TABLE goal_vote_matches (
  id SERIAL PRIMARY KEY,
  goal_id INTEGER REFERENCES party_goals(id),
  beteckning TEXT NOT NULL,
  forslagspunkt TEXT,
  relevance_score FLOAT,      -- 0.0-1.0 from LLM
  aligned_direction TEXT,      -- which vote direction aligns with goal
  explanation TEXT,
  proposed_by_party TEXT,
  proposal_type TEXT,
  context_note TEXT,           -- why party may have voted against
  verified BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE promise_vote_matches (
  id SERIAL PRIMARY KEY,
  promise_id INTEGER REFERENCES promises(id),
  vote_id INTEGER REFERENCES votes(id),
  relevance_score FLOAT,
  alignment TEXT,              -- supports, contradicts, unclear
  explanation TEXT,
  proposed_by_party TEXT,
  proposal_type TEXT,
  verified BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Materialized views for scorecards
CREATE MATERIALIZED VIEW party_scorecards AS
SELECT 
  pg.party,
  pg.id as goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning) as relevant_votes,
  -- Calculate alignment from actual member votes
  ROUND(
    COUNT(CASE WHEN v.vote_result = gvm.aligned_direction THEN 1 END)::numeric / 
    NULLIF(COUNT(v.id), 0) * 100, 1
  ) as alignment_pct
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
LEFT JOIN votes v ON v.beteckning = gvm.beteckning 
  AND v.forslagspunkt = gvm.forslagspunkt 
  AND v.party = pg.party
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;
```

---

## Claude API Prompts

### 1. Extract Promises from Speech
```
System: You analyze Swedish parliament debate speeches. Extract concrete promises/commitments.

Return JSON array:
[{
  "promise_text": "Swedish text of the commitment",
  "topic": "sjukvard|forsvar|invandring|skatt|skola|klimat|brott|bostad|arbete|energi|other",
  "specificity": "concrete|directional|rhetorical",
  "keywords": ["swedish", "keywords", "for", "matching"]
}]

Input: Speaker: {name} ({party}), Date: {date}, Topic: {heading}
Speech: {text}
```

### 2. Extract Goals from Party Platform
```
System: You analyze Swedish political party manifestos. Extract stated goals.

Return JSON array:
[{
  "goal_text": "Swedish text of the goal",
  "topic": "same categories as above",
  "specificity": "concrete|directional|rhetorical", 
  "keywords": ["matching", "keywords"],
  "relevant_committees": ["SoU", "FiU"]
}]

Input: Party: {party}, Document: {title}
Text: {text}
```

### 3. Score Vote Relevance to Goal/Promise
```
System: Given a political goal and a parliamentary vote, determine relevance and alignment direction.

CRITICAL: Consider that voting NEJ on a proposal from another party may still align with the goal if the proposal had unacceptable conditions. Explain this context.

Return JSON:
{
  "relevant": true/false,
  "aligned_direction": "Ja|Nej|unclear",
  "confidence": 0.0-1.0,
  "context_note": "Brief Swedish explanation of political context for why the party voted this way"
}

Input: 
Party goal: {goal_text} (by {party})
Vote on: {document_title}
Proposed by: {proposed_by_party} ({proposal_type})
Committee: {committee}
```

---

## n8n Workflow Design

### Workflow 1: Daily Ingestion
Trigger: Cron (daily 06:00)
1. Fetch new anföranden since last run → store in `speeches`
2. Fetch new voteringar since last run → store in `votes`
3. For each new vote, enrich with proposal origin (trace betänkande → prop/mot → author party)
4. Queue new speeches for promise extraction

### Workflow 2: Promise Extraction  
Trigger: New speeches in queue
1. For each speech, call Claude API with promise extraction prompt
2. Store extracted promises in `promises` table
3. For each promise, find potentially relevant votes (by topic→committee mapping)
4. Call Claude API to score relevance and alignment
5. Store matches in `promise_vote_matches`

### Workflow 3: Party Goal Processing
Trigger: Manual (when new valmanifest/partiprogram is added)
1. Call Claude API to extract goals from party document
2. Store in `party_goals`
3. Find all relevant historical votes
4. Score each goal-vote pair via Claude API
5. Store in `goal_vote_matches` with context_note

### Workflow 4: Periodic Re-scoring
Trigger: Weekly
1. For all party goals, check for new relevant votes since last run
2. Score new vote-goal pairs
3. Refresh materialized views

---

## Go API Endpoints

```
GET /api/parties                           -- list all parties with summary scores
GET /api/parties/{party}/goals             -- party goals with alignment data
GET /api/parties/{party}/goals/{id}/votes  -- detailed vote breakdown for a goal
GET /api/politicians                       -- list politicians with summary stats  
GET /api/politicians/{id}                  -- full politician profile
GET /api/politicians/{id}/promises         -- promises with vote matches
GET /api/politicians/{id}/votes            -- voting record with proposal origins
GET /api/votes/{beteckning}/{punkt}        -- full vote detail with origin context
```

---

## Implementation Order

1. **Database + Ingestion**: Set up PostgreSQL, write Go ingestion service that pulls from Riksdagen API, stores politicians/speeches/votes, enriches votes with proposal origin
2. **Party Goals (manual)**: Manually extract 5-10 goals per party from current valmanifest as seed data, store in party_goals
3. **Vote Matching**: Build the topic→committee mapping, find relevant votes per goal, calculate alignment percentages
4. **Claude Integration**: Wire up promise extraction from speeches and vote-relevance scoring
5. **API Layer**: Go REST API serving scorecards
6. **Frontend**: React app with party goals view and individual politician view
7. **n8n Automation**: Set up periodic ingestion and re-scoring workflows

Start with parties S, M, SD (largest, most data) and the current session 2024/25.

---

## Existing Projects to Reference
- github.com/OAndell/Riksdagskollen -- Android app using same API
- github.com/Reicher/RiksdagenPythonAPI -- Python wrapper
- hack23.com/cia -- Citizen Intelligence Agency, tracks politician metrics
- github.com/welfare-state-analytics/riksdagen-corpus -- Full historical corpus

None of these do promise-to-vote cross-referencing with proposal origin context. That is the differentiator.
