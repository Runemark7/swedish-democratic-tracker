package registret

import "testing"

const sampleFragment = `
<table id="ResultTable"><thead><tr><th>Namn</th><th>Organisationsnr</th><th>SFS</th><th>WebbAdress</th></tr></thead>
<tbody>
<tr><td></td><td></td><td></td><td></td></tr>
<tr><td>Alkoholsortimentsn&#228;mnden</td><td>202100-5943</td><td>2007:1216</td><td>www.kammarkollegiet.se/alkoholsortimentsnamnden</td></tr>
<tr><td>Arbetsf&#246;rmedlingen</td><td>202100-2114</td><td>2007:1030</td><td>www.arbetsformedlingen.se</td></tr>
</tbody></table>`

func TestParseAgencyTable(t *testing.T) {
	g := group{typ: "Förvaltningsmyndighet", principal: "Regeringen", underGov: true}
	got := parseAgencyTable(sampleFragment, g)

	if len(got) != 2 {
		t.Fatalf("expected 2 agencies, got %d", len(got))
	}
	first := got[0]
	if first.Name != "Alkoholsortimentsnämnden" {
		t.Errorf("name: got %q", first.Name)
	}
	if first.OrgNumber != "202100-5943" {
		t.Errorf("org: got %q", first.OrgNumber)
	}
	if first.SFS != "2007:1216" {
		t.Errorf("sfs: got %q", first.SFS)
	}
	if first.Website != "www.kammarkollegiet.se/alkoholsortimentsnamnden" {
		t.Errorf("website: got %q", first.Website)
	}
	if first.Slug != "alkoholsortimentsnamnden" {
		t.Errorf("slug: got %q", first.Slug)
	}
	if !first.UnderGovernment || first.Type != "Förvaltningsmyndighet" {
		t.Errorf("group mapping wrong: %+v", first)
	}
}

const courtFragment = `
<table id="ResultTable"><thead><tr><th>Namn</th><th>CfarNr</th><th>Postadress</th><th>Postnr</th><th>Postort</th><th>WebbAdress</th></tr></thead>
<tbody>
<tr><td></td><td></td><td></td><td></td><td></td><td></td></tr>
<tr><td>ARBETSDOMSTOLEN</td><td>19075852</td><td>BOX 2018</td><td>103 11</td><td>STOCKHOLM</td><td>www.arbetsdomstolen.se</td></tr>
<tr><td>DOMSTOLSVERKET</td><td>20738233</td><td></td><td>551 81</td><td>J&#214;NK&#214;PING</td><td>www.domstol.se</td></tr>
</tbody></table>`

func TestParseAgencyTable_Court(t *testing.T) {
	g := group{typ: "Domstol", principal: "Riksdagen", underGov: false, schema: schemaCourt}
	got := parseAgencyTable(courtFragment, g)

	if len(got) != 2 {
		t.Fatalf("expected 2 courts, got %d", len(got))
	}
	if got[0].OrgNumber != "cfar:19075852" {
		t.Errorf("synth org: got %q", got[0].OrgNumber)
	}
	if got[0].Name != "ARBETSDOMSTOLEN" {
		t.Errorf("name: got %q", got[0].Name)
	}
	if got[0].Website != "www.arbetsdomstolen.se" {
		t.Errorf("website: got %q", got[0].Website)
	}
	if got[0].Type != "Domstol" || got[0].UnderGovernment {
		t.Errorf("classification: got %+v", got[0])
	}
}

const foreignFragment = `
<table id="ResultTable"><thead><tr><th>Land</th><th>LopNr</th><th>Namn</th><th>Ambassad&#246;r</th><th>WebbAdress</th></tr></thead>
<tbody>
<tr><td></td><td>201</td><td></td><td></td><td></td></tr>
<tr><td>Afghanistan</td><td>210</td><td>Sveriges ambassad Kabul</td><td>Torkel Stiernl&#246;f</td><td>www.swedenabroad.se/sv/utlandsmyndigheter/afghanistan-kabul</td></tr>
</tbody></table>`

func TestParseAgencyTable_Foreign(t *testing.T) {
	g := group{typ: "Utlandsmyndighet", principal: "Regeringen", underGov: true, schema: schemaForeign}
	got := parseAgencyTable(foreignFragment, g)

	if len(got) != 1 {
		t.Fatalf("expected 1 foreign mission (empty name row skipped), got %d", len(got))
	}
	if got[0].OrgNumber != "utland:210" {
		t.Errorf("synth org: got %q", got[0].OrgNumber)
	}
	if got[0].Name != "Sveriges ambassad Kabul" {
		t.Errorf("name: got %q", got[0].Name)
	}
	if got[0].Website != "www.swedenabroad.se/sv/utlandsmyndigheter/afghanistan-kabul" {
		t.Errorf("website: got %q", got[0].Website)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Åklagarmyndigheten": "aklagarmyndigheten",
		"Sveriges Domstolar": "sveriges-domstolar",
		"Försäkringskassan":  "forsakringskassan",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
