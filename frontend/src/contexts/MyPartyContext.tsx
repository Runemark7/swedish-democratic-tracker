import { useState } from "react";
import { MY_PARTY_KEY, MyPartyContext, readStoredParty } from "./myParty";

export function MyPartyProvider({ children }: { children: React.ReactNode }) {
  const [myParty, setState] = useState<string | null>(readStoredParty);

  function setMyParty(party: string | null) {
    setState(party);
    try {
      if (party) localStorage.setItem(MY_PARTY_KEY, party);
      else localStorage.removeItem(MY_PARTY_KEY);
    } catch {
      /* storage unavailable: the choice lasts for this page view only */
    }
  }

  return <MyPartyContext.Provider value={{ myParty, setMyParty }}>{children}</MyPartyContext.Provider>;
}
