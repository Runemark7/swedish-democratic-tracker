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
