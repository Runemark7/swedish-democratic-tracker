import { test } from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { resolveDocHref } from "../../src/features/data/docLinks.ts";

test("relative .md links map to /data/<id>", () => {
  assert.equal(resolveDocHref("./voteringar.md"), "/data/voteringar");
  assert.equal(resolveDocHref("party-manifestos.md"), "/data/party-manifestos");
  assert.equal(resolveDocHref("./kolada.md#x"), "/data/kolada#x");
});

test("other links are left alone", () => {
  assert.equal(resolveDocHref("https://example.com/a.md"), null);
  assert.equal(resolveDocHref("/data/kolada"), null);
  assert.equal(resolveDocHref(undefined), null);
});

test("every ./x.md link in data-source docs resolves to an existing doc", () => {
  const dir = path.resolve(import.meta.dirname, "../../../docs/data-sources");
  for (const f of fs.readdirSync(dir).filter((f) => f.endsWith(".md"))) {
    for (const m of fs.readFileSync(path.join(dir, f), "utf8").matchAll(/\]\((\.\/[^)\s]+\.md[^)]*)\)/g)) {
      const to = resolveDocHref(m[1]);
      assert.ok(to, `${f}: ${m[1]}`);
      const id = to.slice("/data/".length).split("#")[0];
      assert.ok(fs.existsSync(path.join(dir, `${id}.md`)), `${f} links to missing ${id}`);
    }
  }
});
