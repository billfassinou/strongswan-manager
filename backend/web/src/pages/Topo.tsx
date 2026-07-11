import { useEffect, useState } from "react";
import { api, Gateway, Tunnel } from "../api";
import { statusClass } from "../ui";

// Topologie calculée à partir des données réelles (/gateways + /tunnels), sans backend
// dédié : les passerelles sont des nœuds, les tunnels des arêtes colorées par état.
export function Topo() {
  const [gws, setGws] = useState<Gateway[]>([]);
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  useEffect(() => {
    Promise.all([api.get("/gateways"), api.get("/tunnels")]).then(([g, t]) => {
      setGws(g.items || []);
      setTunnels(t.items || []);
    });
  }, []);

  const W = 760,
    H = 420,
    cx = W / 2,
    cy = H / 2,
    R = Math.min(W, H) / 2 - 90;
  const pos: Record<string, { x: number; y: number }> = {};
  gws.forEach((g, i) => {
    const a = (i / Math.max(1, gws.length)) * Math.PI * 2 - Math.PI / 2;
    pos[g.id] = { x: cx + R * Math.cos(a), y: cy + R * Math.sin(a) };
  });

  const edges = tunnels.map((t, i) => {
    const a = pos[t.gateway_id];
    let b = t.peer_gateway_id ? pos[t.peer_gateway_id] : undefined;
    if (!b) {
      // extrémité distante non gérée : satellite près de la passerelle
      const base = a || { x: cx, y: cy };
      const ang = (i / Math.max(1, tunnels.length)) * Math.PI * 2;
      b = { x: base.x + 70 * Math.cos(ang), y: base.y + 70 * Math.sin(ang) };
    }
    return { t, a: a || { x: cx, y: cy }, b, managed: !!t.peer_gateway_id };
  });

  const col = (s: string) => `var(--${statusClass(s) === "u" ? "up" : statusClass(s) === "n" ? "warn" : statusClass(s) === "d" ? "down" : "text-faint"})`;

  return (
    <div>
      <div className="card">
        <div className="card-head">
          <h2>Vue logique du réseau VPN</h2>
          <div className="act muted" style={{ fontSize: 12 }}>
            {gws.length} passerelles · {tunnels.length} tunnels
          </div>
        </div>
        <div style={{ padding: 10 }}>
          <svg viewBox={`0 0 ${W} ${H}`} width="100%" style={{ maxHeight: 440 }}>
            {edges.map((e, i) => (
              <line
                key={i}
                x1={e.a.x}
                y1={e.a.y}
                x2={e.b.x}
                y2={e.b.y}
                stroke={col(e.t.status)}
                strokeWidth={2.5}
                strokeDasharray={statusClass(e.t.status) === "n" ? "7 6" : statusClass(e.t.status) === "d" ? "3 7" : ""}
              />
            ))}
            {edges
              .filter((e) => !e.managed)
              .map((e, i) => (
                <circle key={"s" + i} cx={e.b.x} cy={e.b.y} r={6} fill="var(--surface-2)" stroke={col(e.t.status)} strokeWidth={1.5} />
              ))}
            {gws.map((g) => (
              <g key={g.id}>
                <circle cx={pos[g.id].x} cy={pos[g.id].y} r={34} fill="var(--accent-soft)" stroke={col(g.status)} strokeWidth={2.5} />
                <text x={pos[g.id].x} y={pos[g.id].y - 2} textAnchor="middle" fontSize={13} fontWeight={600} fill="var(--text)">
                  {g.name}
                </text>
                <text x={pos[g.id].x} y={pos[g.id].y + 14} textAnchor="middle" fontSize={10} fill="var(--text-dim)">
                  {tunnels.filter((t) => t.gateway_id === g.id).length} tunnels
                </text>
              </g>
            ))}
          </svg>
        </div>
      </div>
      <div className="grid cols-4 mt">
        <Stat lbl="Passerelles" val={gws.length} />
        <Stat lbl="Liens sains" val={tunnels.filter((t) => t.status === "up").length} cls="up" />
        <Stat lbl="En négociation" val={tunnels.filter((t) => t.status === "negotiating").length} cls="warn" />
        <Stat lbl="Down" val={tunnels.filter((t) => t.status === "down").length} cls="down" />
      </div>
    </div>
  );
}
function Stat({ lbl, val, cls }: { lbl: string; val: any; cls?: string }) {
  return (
    <div className="card stat">
      <div className="lbl">{lbl}</div>
      <div className={"val " + (cls || "")}>{val}</div>
    </div>
  );
}
