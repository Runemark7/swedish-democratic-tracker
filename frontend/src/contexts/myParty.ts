import { createContext, useContext } from "react";
import { PARTIES } from "@/shared/design";

/** localStorage key. The choice never leaves the browser: no API call reads or sends it. */
export const MY_PARTY_KEY = "sdt-my-party";

export interface MyPartyContextValue {
  myParty: string | null;
  setMyParty: (party: string | null) => void;
}

export const MyPartyContext = createContext<MyPartyContextValue>({
  myParty: null,
  setMyParty: () => {},
});

export function readStoredParty(): string | null {
  try {
    const v = localStorage.getItem(MY_PARTY_KEY);
    return v && (PARTIES as string[]).includes(v) ? v : null;
  } catch {
    return null;
  }
}

export function useMyParty() {
  return useContext(MyPartyContext);
}

/** Highlight only: a ring and a marker. Callers must never filter or reorder by it. */
export const MY_PARTY_RING = "0 0 0 2px var(--color-accent)";
