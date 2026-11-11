import { useQuery } from "@tanstack/react-query";
import { Navigate, useParams } from "react-router-dom";
import { loftenApi } from "./api";

/** Stable URL of a Bekräftad koppling: it lives on its Löfte's page, anchored. */
export function KopplingPage() {
  const { id = "" } = useParams<{ id: string }>();
  const { data, isLoading, isError } = useQuery({
    queryKey: ["koppling", id],
    queryFn: () => loftenApi.koppling(Number(id)),
    retry: false,
  });
  if (isLoading) return <p className="text-sm p-8">Laddar...</p>;
  if (isError || !data) return <p className="text-sm p-8">Kopplingen finns inte eller är inte bekräftad.</p>;
  return <Navigate to={`/loften/${data.loefteId}#koppling-${data.id}`} replace />;
}
