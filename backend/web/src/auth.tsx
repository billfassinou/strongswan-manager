import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { api, getToken, setToken, Me } from "./api";

interface AuthCtx {
  me: Me | null;
  ready: boolean;
  refresh: () => Promise<void>;
  logout: () => void;
}

const Ctx = createContext<AuthCtx>({ me: null, ready: false, refresh: async () => {}, logout: () => {} });

export function AuthProvider({ children }: { children: ReactNode }) {
  const [me, setMe] = useState<Me | null>(null);
  const [ready, setReady] = useState(false);

  async function refresh() {
    if (!getToken()) {
      setMe(null);
      setReady(true);
      return;
    }
    try {
      setMe(await api.get("/me"));
    } catch {
      setToken("");
      setMe(null);
    } finally {
      setReady(true);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  function logout() {
    setToken("");
    setMe(null);
  }

  return <Ctx.Provider value={{ me, ready, refresh, logout }}>{children}</Ctx.Provider>;
}

export const useAuth = () => useContext(Ctx);
