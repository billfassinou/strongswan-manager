import { useEffect, useRef, useState } from "react";
import { getToken } from "./api";

export interface TunnelStatusEvent {
  type: string;
  id: string;
  name: string;
  status: string;
}

// useLiveStatus ouvre le WebSocket /api/v1/ws et renvoie la dernière mise à jour d'état
// de tunnel reçue, plus l'état de connexion du flux.
export function useLiveStatus(): { last: TunnelStatusEvent | null; connected: boolean } {
  const [last, setLast] = useState<TunnelStatusEvent | null>(null);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let closed = false;
    let retry: ReturnType<typeof setTimeout>;

    function connect() {
      const proto = location.protocol === "https:" ? "wss" : "ws";
      const url = `${proto}://${location.host}/api/v1/ws?token=${encodeURIComponent(getToken())}`;
      const ws = new WebSocket(url);
      wsRef.current = ws;
      ws.onopen = () => setConnected(true);
      ws.onclose = () => {
        setConnected(false);
        if (!closed) retry = setTimeout(connect, 3000);
      };
      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data);
          if (msg.type === "tunnel_status") setLast(msg);
        } catch {
          /* ignore */
        }
      };
    }
    connect();
    return () => {
      closed = true;
      clearTimeout(retry);
      wsRef.current?.close();
    };
  }, []);

  return { last, connected };
}
