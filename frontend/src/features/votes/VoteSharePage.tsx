import { useParams, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { votesApi } from "./api";
import { VoteShareCard } from "./VoteShareCard";

export function VoteSharePage() {
  const { beteckning = "", punkt = "" } = useParams();
  // The riksmöte names which omröstning this is: beteckning + punkt alone
  // recur every riksmöte. Old addresses without it still open, and the API
  // then answers with the latest riksmöte holding that punkt.
  const [sp] = useSearchParams();
  const session = sp.get("session") ?? sp.get("rm") ?? "";

  const { data, isLoading, isError } = useQuery({
    queryKey: ["vote-detail", beteckning, punkt, session],
    queryFn: () => votesApi.getDetail(beteckning, punkt, session || undefined),
    enabled: !!beteckning && !!punkt,
    staleTime: 5 * 60 * 1000,
  });

  // Back to the same riksmöte the card shows, not whichever is latest.
  const backSession = session || data?.session || "";
  const backHref =
    `/votes/${encodeURIComponent(beteckning)}/${encodeURIComponent(punkt)}` +
    (backSession ? `?session=${encodeURIComponent(backSession)}` : "");

  return (
    <div
      style={{
        minHeight: "100vh",
        background: "var(--color-share-page)",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        padding: "48px 20px 64px",
      }}
    >
      {isLoading && (
        <p
          style={{
            fontFamily: "monospace",
            fontSize: 12,
            color: "var(--color-share-faint)",
            letterSpacing: "0.1em",
          }}
        >
          Laddar omröstning...
        </p>
      )}

      {isError && (
        <p
          style={{
            fontFamily: "monospace",
            fontSize: 12,
            color: "#ef4444",
            letterSpacing: "0.1em",
          }}
        >
          Kunde inte hämta omröstningsdata.
        </p>
      )}

      {data && <VoteShareCard data={data} />}

      <a
        href={backHref}
        style={{
          marginTop: 28,
          fontFamily: "monospace",
          fontSize: 11,
          color: "var(--color-share-faint)",
          textDecoration: "none",
          letterSpacing: "0.12em",
          textTransform: "uppercase",
        }}
      >
        ← Tillbaka till omröstningen
      </a>
    </div>
  );
}
