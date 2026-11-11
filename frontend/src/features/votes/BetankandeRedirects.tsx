import { useQuery } from "@tanstack/react-query";
import { Navigate, useParams, useSearchParams } from "react-router-dom";
import { betankandeApi, betankandePath } from "./betankandeApi";

function Waiting({ failed }: { failed: boolean }) {
  return (
    <div className="sdt-page px-8 py-16">
      <p className="text-sm text-on-surface-variant" data-testid="redirect-status">
        {failed ? "Betänkandet finns inte i vårt register." : "Letar upp betänkandet…"}
      </p>
    </div>
  );
}

/** Old vote detail URL (/votes/:beteckning/:punkt) → the canonical page,
 *  anchored to that förslagspunkt in the votering's own riksmöte. */
export function VoteRedirect() {
  const { beteckning = "", punkt = "" } = useParams<{ beteckning: string; punkt: string }>();
  const [sp] = useSearchParams();
  const rm = sp.get("session") ?? sp.get("rm") ?? "";
  const q = useQuery({
    queryKey: ["resolve", beteckning, punkt, rm],
    queryFn: () => betankandeApi.resolve(beteckning, punkt, rm || undefined),
    retry: false,
  });
  if (q.data) return <Navigate to={betankandePath(q.data.riksmote, q.data.beteckning, punkt)} replace />;
  return <Waiting failed={q.isError} />;
}

/** Old betänkande URL (/beslut/:beteckning, optionally ?rm=) → the canonical
 *  page of the most recent riksmöte holding that beteckning. */
export function BeslutRedirect() {
  const { beteckning = "" } = useParams<{ beteckning: string }>();
  const [sp] = useSearchParams();
  const rm = sp.get("rm") ?? "";
  const q = useQuery({
    queryKey: ["resolve", beteckning, "", rm],
    queryFn: () => betankandeApi.resolve(beteckning, undefined, rm || undefined),
    retry: false,
  });
  if (q.data) return <Navigate to={betankandePath(q.data.riksmote, q.data.beteckning)} replace />;
  return <Waiting failed={q.isError} />;
}
