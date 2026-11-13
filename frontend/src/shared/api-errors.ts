import { ApiError } from "./api-client.ts";

// Only the API saying "no such record" means a record is missing. A 429, a
// 5xx or a dropped connection is a temporary failure: the record may well be
// there, so pages must not say it is not.

const MAX_RETRIES = 2;
const MAX_DELAY_MS = 5000;

/** The API answered that the record does not exist (404) or that the id is
 *  not one it can hold (400). */
export function isNotFound(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 404 || error.status === 400);
}

/** Rate limiting, server errors and network failures can pass on their own. */
function isTemporary(error: unknown): boolean {
  if (error instanceof ApiError) return error.status === 429 || error.status === 408 || error.status >= 500;
  return error != null;
}

/** TanStack Query `retry`: two more attempts for temporary failures, none for
 *  answers that will not change. */
export function shouldRetry(failureCount: number, error: unknown): boolean {
  return failureCount < MAX_RETRIES && isTemporary(error);
}

/** TanStack Query `retryDelay`: wait as long as the server asks (Retry-After),
 *  otherwise 1 s, 2 s, 4 s; never more than five seconds. */
export function retryDelayMs(failureCount: number, error: unknown): number {
  const asked = error instanceof ApiError ? error.retryAfter : undefined;
  const ms = asked !== undefined ? asked * 1000 : 1000 * 2 ** failureCount;
  return Math.min(ms, MAX_DELAY_MS);
}
