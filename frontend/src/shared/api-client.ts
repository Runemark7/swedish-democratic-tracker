const BASE = "/api";

export class ApiError extends Error {
  status: number;
  /** Seconds the server asked us to wait (Retry-After), when it said so. */
  retryAfter?: number;

  constructor(status: number, message: string, retryAfter?: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.retryAfter = retryAfter;
  }
}

async function request<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) {
    // The rate limiter answers in plain text, so the body may not be JSON.
    const body = await res.json().catch(() => ({ error: res.statusText }));
    const wait = Number(res.headers.get("Retry-After"));
    throw new ApiError(
      res.status,
      body.error ?? res.statusText,
      Number.isFinite(wait) && wait > 0 ? wait : undefined,
    );
  }
  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
};
