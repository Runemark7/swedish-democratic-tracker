# API seam tests

These tests build the real HTTP router against a real Postgres database and
recorded responses from data.riksdagen.se. They do not call the live API.

`api.NewRouter` is the constructor `cmd/api` uses. A test passes it a database
pool, an empty write token, and the replay server's URL. Ingestion is then run
by worker name, and the result is read back through the public API.

## Run

CI sets `TEST_DATABASE_URL` on the Test backend job (Postgres 18, migrations
applied to the shared database the older package tests use). The seam creates
its own database from that URL, migrates it with golang-migrate, and drops it
when the test ends. The role must be allowed to `CREATE DATABASE`.

Locally, export either variable:

```bash
export TEST_DATABASE_URL='postgres://riksdagskollen:localdev@127.0.0.1:5432/riksdagskollen?sslmode=disable'
# DATABASE_URL is used when TEST_DATABASE_URL is unset
cd backend && go test -count=1 ./internal/apitest/ ./internal/api/
```

Without either variable the database tests skip. The replay tests do not need a
database.

## Record a fixture

1. Find the URL the client builds. For a votering that is
   `GET /dokumentlista/?doktyp=votering&rm=2025/26&utformat=json&sz=200&p=1&sort=datum&sortorder=desc`
   and then
   `GET /voteringlista/?rm=2025/26&parti=&iid=&bet=UbU30&sz=10000&utformat=json`.
   Members are
   `GET /personlista/?rdlstatus=tjanstgorande&parti=&utformat=json`.
2. Save the live response:

```bash
curl -fsS \
  'https://data.riksdagen.se/voteringlista/?rm=2025/26&parti=&iid=&bet=UbU30&sz=10000&utformat=json' \
  -o voteringlista.body.json
```

3. Wrap it in a JSON file in the fixture directory. The replay key is the
   method, the path, and the query. Query order does not matter. Empty
   parameters the client sends (`parti=`, `iid=`) must be present.

```json
{
  "method": "GET",
  "path": "/voteringlista/",
  "query": {
    "rm": "2025/26",
    "parti": "",
    "iid": "",
    "bet": "UbU30",
    "sz": "10000",
    "utformat": "json"
  },
  "status": 200,
  "recordedFrom": "https://data.riksdagen.se/voteringlista/?rm=2025/26&bet=UbU30&sz=10000&utformat=json",
  "body": {}
}
```

Paste the curl body into `body`. `recordedFrom` is ignored by the server; it
records where the bytes came from.

4. A full riksmöte is too large to commit. Delete the rows the test does not
   need and keep the JSON shape the client decodes. `dokument` and `votering`
   and `person` must stay JSON arrays even when one row remains: Riksdagen
   sometimes returns a single object, and the client decodes an array.
   Fields the client does not read can be dropped. Do not invent values.

An unrecorded request fails the test and returns 501. The client is not
allowed to fall through to the live API.

## Write a test

Pin the clock to a day in the riksmöte you recorded. The votes worker asks
for `CurrentRiksmote` of that instant. 13 August 2026 is still riksmöte
2025/26.

```go
h := apitest.Start(t, fixtureDir, apitest.WithNow(time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)))
h.RunWorker(ctx, "fetch-politicians")
h.RunWorker(ctx, "fetch-votes")
if h.Served("/voteringlista/") == 0 {
    t.Fatal("ingestion did not read the recording")
}
rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/politicians/"+id+"/votes?session=2025%2F26", nil))
```

Assert on the HTTP response. Do not assert on tables.

Worker names include `fetch-politicians`, `fetch-votes`, `fetch-speeches`,
`enrich-speech-text`, `link-speeches`, `keyword-matcher`,
`refresh-scorecards`, `refresh-coverage`, `agency-intel`, `authorities`,
`authority-headcount`, `authority-expenditure`, `national-kpis`,
`region-budget`, `municipality-budget`, and `sync-proposals`.
Run only the workers whose requests you have recorded.

`fixtures/one-votering` is the worked example: one recorded votering
(`8BADA090-C009-495A-98D2-F967E799BA61`, 2025/26:UbU30 förslagspunkt 2)
and the members who voted in it, cut down from the live responses on
2026-11-01. `TestRecordedVoteringIsReadableThroughTheVotesAPI` ingests it
and reads Kenneth G Forslund's Nej back from `GET /api/politicians/{id}/votes`.
