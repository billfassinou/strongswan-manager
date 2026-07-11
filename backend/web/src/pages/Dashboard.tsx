import { useEffect, useState } from "react";
import { api, Tunnel, Gateway } from "../api";
import { Pill, scoreColor } from "../ui";
import { useLiveStatus } from "../ws";

export function Dashboard() {
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const { last, connected } = useLiveStatus();

  async function load() {
    const [t, g] = await Promise.all([api.get("/tunnels"), api.get("/gateways")]);
    setTunnels(t.items || []);
    setGateways(g.items || []);
  }
  useEffect(() => {
    load();
  }, []);

  // applique les mises à jour d'état temps réel reçues par WebSocket
  useEffect(() => {
    if (!last) return;
    setTunnels((ts) => ts.map((t) => (t.id === last.id ? { ...t, status: last.status } : t)));
  }, [last]);

  const up = tunnels.filter((t) => t.status === "up").length;
  const neg = tunnels.filter((t) => t.status === "negotiating" || t.status === "installing").length;
  const down = tunnels.filter((t) => t.status === "down").length;

  return (
    <div>
      <div className="grid cols-4 mb">
        <Stat lbl="Tunnels actifs" val={`${up} / ${tunnels.length}`} dot="u" />
        <Stat lbl="En négociation" val={neg} cls="warn" dot="n" />
        <Stat lbl="Tunnels down" val={down} cls="down" dot="d" />
        <Stat lbl="Passerelles" val={gateways.length} />
      </div>

      <div className="grid cols-2 mb">
        <div className="card">
          <div className="card-head">
            <h2>Tunnels</h2>
            <div className="act muted" style={{ fontSize: 12 }}>
              <span className="livedot" style={{ background: connected ? "var(--up)" : "var(--text-faint)" }}></span>{" "}
              {connected ? "temps réel" : "hors ligne"}
            </div>
          </div>
          {tunnels.length === 0 ? (
            <div className="empty">Aucun tunnel.</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Nom</th>
                  <th>Type</th>
                  <th>État</th>
                  <th className="num">Score</th>
                </tr>
              </thead>
              <tbody>
                {tunnels.map((t) => (
                  <tr key={t.id}>
                    <td>
                      <b>{t.name}</b>
                    </td>
                    <td className="muted">
                      {t.type} · IKEv{t.ike_version}
                    </td>
                    <td>
                      <Pill status={t.status} />
                    </td>
                    <td className="num">
                      <b className={scoreColor(t.security_score)}>{t.security_score}</b>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        <div className="card">
          <div className="card-head">
            <h2>Passerelles</h2>
          </div>
          <table>
            <thead>
              <tr>
                <th>Nom</th>
                <th>Version</th>
                <th>État</th>
              </tr>
            </thead>
            <tbody>
              {gateways.map((g) => (
                <tr key={g.id}>
                  <td>
                    <b>{g.name}</b> <span className="muted mono">{g.endpoint}</span>
                  </td>
                  <td className="mono muted">{g.version || "—"}</td>
                  <td>
                    <Pill status={g.status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function Stat({ lbl, val, cls, dot }: { lbl: string; val: any; cls?: string; dot?: string }) {
  return (
    <div className="card stat">
      <div className="lbl">
        {dot && <span className={"dot " + dot}></span>}
        {lbl}
      </div>
      <div className={"val " + (cls || "")}>{val}</div>
    </div>
  );
}
