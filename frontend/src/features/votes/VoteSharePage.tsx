import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { votesApi } from "./api";
import { VoteShareCard } from "./VoteShareCard";

export function VoteSharePage() {
  const { beteckning = "", punkt = "" } = useParams();

  const { data, isLoading, isError } = useQuery({
    queryKey: ["vote-detail", beteckning, punkt],
    queryFn: () => votesApi.getDetail(beteckning, punkt),
    enabled: !!beteckning && !!punkt,
    staleTime: 5 * 60 * 1000,
  });

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
        href={`/votes/${beteckning}/${punkt}`}
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
