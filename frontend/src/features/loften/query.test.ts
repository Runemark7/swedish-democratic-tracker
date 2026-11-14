import assert from "node:assert/strict";
import test from "node:test";
import { loftenCsvHref, loftenQuery } from "./query.ts";

test("no filters, no query string", () => {
  assert.equal(loftenQuery({}), "");
  assert.equal(loftenCsvHref({}), "/api/loften.csv");
});

test("the download carries every active filter", () => {
  assert.equal(
    loftenCsvHref({ party: "S", utskott: "FöU", status: "inte-provat", betankande: "2026/27:SkU2" }),
    "/api/loften.csv?party=S&utskott=F%C3%B6U&status=inte-provat&betankande=2026%2F27%3ASkU2",
  );
});

test("empty filters are left out", () => {
  assert.equal(loftenCsvHref({ party: "", utskott: "SkU", status: "" }), "/api/loften.csv?utskott=SkU");
});

test("the download carries the mandate-period filter", () => {
  assert.equal(loftenCsvHref({ mandatperiod: "2026-2030" }), "/api/loften.csv?mandatperiod=2026-2030");
  assert.equal(
    loftenCsvHref({ party: "S", mandatperiod: "2022-2026" }),
    "/api/loften.csv?party=S&mandatperiod=2022-2026",
  );
});
