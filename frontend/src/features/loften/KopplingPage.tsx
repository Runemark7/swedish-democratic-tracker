import { useQuery } from "@tanstack/react-query";
import { Navigate, useParams } from "react-router-dom";
import { loftenApi } from "./api";
import { isNotFound } from "@/shared/api-errors";
import { TemporaryError } from "@/shared/components";

/** Stable URL of a Bekräftad koppling: it lives on its Löfte's page, anchored. */
export function KopplingPage() {
  const { id = "" } = useParams<{ id: string }>();
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["koppling", id],
    queryFn: () => loftenApi.koppling(Number(id)),
  });
  if (isLoading) return <p className="text-sm p-8">Laddar...</p>;
  if (error && !isNotFound(error)) {
    return <div className="p-8"><TemporaryError what="Kopplingen" onRetry={() => void refetch()} /></div>;
  }
  if (error || !data) return <p className="text-sm p-8">Kopplingen finns inte eller är inte bekräftad.</p>;
  return <Navigate to={`/loften/${data.loefteId}#koppling-${data.id}`} replace />;
}
