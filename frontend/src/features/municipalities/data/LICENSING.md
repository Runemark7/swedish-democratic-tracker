# Kommuner SVG — license + attribution

`sweden-kommuner.svg` and the generated `sweden-kommuner-svg.ts` are derived from:

- **File**: SWE-Map Kommuner.svg
- **Source**: https://commons.wikimedia.org/wiki/File:SWE-Map_Kommuner.svg
- **Author**: Lokal_Profil (Wikimedia Commons)
- **License**: Creative Commons Attribution-ShareAlike 2.5 (CC-BY-SA 2.5)
- **Original data**: Statistics Sweden (SCB)

## Attribution requirement

Anywhere this map is rendered in production, the UI must include a visible
attribution that satisfies CC-BY-SA. Recommended footer text:

> Kommunkarta: Lokal_Profil / Wikimedia Commons (CC BY-SA 2.5),
> grunddata från Statistiska centralbyrån (SCB).

## ShareAlike implication

CC-BY-SA is a copyleft license. If we modify the SVG, the modified version
must be made available under the same license. Embedding it in our app is
fine; redistributing the SVG itself or a modified version requires
re-publishing under CC-BY-SA.

## Path identifiers

Each `<path>` carries an `id` attribute matching the official 4-digit SCB
kommunkod (e.g. `id="0180"` for Stockholms stad). The first two digits of
each code are the län (region) code and are reused as `regionCode`.
