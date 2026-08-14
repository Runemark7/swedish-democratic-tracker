import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { committeesApi } from "@/features/committees/api";
import { useRecordCoverage } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { SourceMarker } from "@/components/sources/SourceMarker";

/**
 * Every committee the record shows for the current mandate period,
 * alphabetical by Swedish name (never by vote count — that would rank the
 * committees, an editorial act this site does not perform, #100).
 *
 * Each row shows the committee's name alone. We hold no plain-language
 * remit for any committee — the API carries only code, name, voteringar and
 * expenditureAreas — so writing one here would be inventing what the Riksdag
 * itself has not published (finding 5, phase 3 pre-flight).
 */
export function CommitteeIndex() {
  const isMobile = useMediaQuery("(max-width: 640px)");

  const {
    data: coverage,
    isLoading: loadingCoverage,
    isError: coverageError,
  } = useRecordCoverage();
  const period = coverage?.mandate.code;

  const {
    data,
    isLoading: loadingCommittees,
    isError: committeesError,
  } = useQuery({
    queryKey: ["committees", period],
    queryFn: () => committeesApi.listCommittees(period!),
    enabled: !!period,
    staleTime: 5 * 60_000,
  });

  const loading = loadingCoverage || (!!period && loadingCommittees);
  // Gated on having nothing to show rather than on isError: TanStack keeps
  // `data` across a failed background refetch, so keying off the error alone
  // would hide 16 committees that are already in cache and correct.
  const failed = (coverageError || committeesError) && !data;

  // Alphabetical by name, not by the code order the API happens to return
  // (which sorts by code — "UFöU" would land near "UU" instead of beside
  // "Sammansatta..." where its actual Swedish name puts it).
  const committees = [...(data ?? [])].sort((a, b) =>
    a.name.localeCompare(b.name, "sv"),
  );

  const textStyle: React.CSSProperties = {
    fontFamily: "var(--font-body)",
    fontSize: 13,
    color: "var(--color-fg-muted)",
    margin: 0,
  };

  return (
    <section style={{ margin: isMobile ? "20px 14px 6px" : "32px 32px 6px" }}>
      <header
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "0.15em",
          color: "var(--color-fg-muted)",
          textTransform: "uppercase",
          marginBottom: 10,
        }}
      >
        UTSKOTTEN
      </header>

      {loading && <p style={textStyle}>Laddar utskotten...</p>}

      {!loading && failed && (
        <p style={textStyle}>
          Vi kunde inte hämta utskotten just nu. Försök igen.
        </p>
      )}

      {!loading && !failed && committees.length === 0 && (
        <p style={{ ...textStyle, fontStyle: "italic" }}>
          Vi har inga utskott inlästa för den här mandatperioden.
          <SourceMarker sourceId="riksdagen" />
        </p>
      )}

      {!loading && !failed && committees.length > 0 && (
        <>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: isMobile ? "1fr" : "repeat(2, minmax(0, 1fr))",
              gap: 1,
              background: "var(--color-border)",
              border: "1px solid var(--color-border)",
            }}
          >
            {committees.map((c) => (
              <Link
                key={c.code}
                to={`/committees/${encodeURIComponent(c.code)}`}
                style={{
                  background: "var(--color-sdt-surface)",
                  padding: isMobile ? "12px 14px" : "12px 16px",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "baseline",
                  gap: 10,
                  textDecoration: "none",
                  color: "inherit",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                >
                  {c.name}
                </span>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                    whiteSpace: "nowrap",
                    flexShrink: 0,
                  }}
                >
                  {c.code} · {c.voteringar.toLocaleString("sv-SE")} voteringar
                </span>
              </Link>
            ))}
          </div>
          {/* The same two qualifiers the committee page carries on this number.
              Without the period the count reads as all-time, and without the
              shortfall it reads as complete — CommitteePage states both, and a
              front page that states neither publishes the same figure as a
              firmer claim than the detail page is willing to make. */}
          <p style={{ ...textStyle, fontSize: 11, marginTop: 10 }}>
            Antalet voteringar avser {coverage?.mandate.label ?? "mandatperioden"}
            {coverage?.mandate.ended ? " (avslutad)" : ""} och är en räkning av
            registret, inte ett mått på utskottets betydelse. Listan är
            alfabetisk.
            {coverage && coverage.expected > 0 && coverage.ingested < coverage.expected && (
              <>
                {" "}Vi saknar ett fåtal av periodens omröstningar.{" "}
                <Link to="/data" style={{ color: "inherit" }}>
                  Vad saknas?
                </Link>
              </>
            )}
            <SourceMarker sourceId="riksdagen" />
          </p>
        </>
      )}
    </section>
  );
}
