import { useEffect, useState } from "react";
import { api, Tunnel } from "../api";
import { scoreTunnel, scoreColor } from "../score";
import { useNav } from "../nav";
import { useAuth } from "../auth";

export function Security() {
  const { go } = useNav();
  const { me } = useAuth();
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  useEffect(() => {
    api.get("/tunnels").then((r) => setTunnels(r.items || []));
  }, []);

  const scored = tunnels.map((t) => ({ t, ...scoreTunnel(t) }));
  const avg = scored.length ? Math.round(scored.reduce((a, x) => a + x.score, 0) / scored.length) : 100;
  const col = scoreColor(avg);
  const off = 314 - (314 * avg) / 100;
  const weakRows = scored.flatMap((x) =>
    x.notes.filter((n) => /faible|obsolète|MD5|DH|3DES|quantique/i.test(n)).map((n) => ({ t: x.t, note: n }))
  );

  return (
    <div>
      <div className="grid" style={{ gridTemplateColumns: "1fr 1fr", gap: 16, marginBottom: 16 }}>
        <div className="card" style={{ padding: 18, display: "flex", alignItems: "center", gap: 22 }}>
          <div style={{ position: "relative", width: 112, height: 112, flex: "none" }}>
            <svg width="112" height="112" style={{ transform: "rotate(-90deg)" }}>
              <circle cx="56" cy="56" r="50" fill="none" stroke="var(--surface-3)" strokeWidth="10" />
              <circle cx="56" cy="56" r="50" fill="none" stroke={`var(--${col})`} strokeWidth="10" strokeLinecap="round" strokeDasharray="314" strokeDashoffset={off} />
            </svg>
            <div style={{ position: "absolute", inset: 0, display: "grid", placeItems: "center" }}>
              <b className={col} style={{ fontSize: 29, fontWeight: 700 }}>
                {avg}
              </b>
            </div>
          </div>
          <div>
            <h3 style={{ marginBottom: 6 }}>Posture de sécurité du parc</h3>
            <p className="muted" style={{ margin: "0 0 10px" }}>
              Moyenne sur {scored.length} tunnels (score calculé côté client, aligné sur le backend).
            </p>
            <div className="row wrap">
              <span className="pill u">{scored.filter((x) => x.score >= 85).length} conformes</span>
              <span className="pill n">{scored.filter((x) => x.score >= 65 && x.score < 85).length} à durcir</span>
              <span className="pill d">{scored.filter((x) => x.score < 65).length} critiques</span>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="card-head">
            <h2>Conformité</h2>
          </div>
          <div style={{ padding: 18, color: "var(--text-dim)", fontSize: 13 }}>
            Score par tunnel (édition Community). Les rapports ANSSI / ISO 27001 / PCI-DSS et l'audit consolidé relèvent
            de l'édition Premium.
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-head">
          <h2>Algorithmes faibles détectés</h2>
          <div className="act muted" style={{ fontSize: 12 }}>
            {weakRows.length} constats
          </div>
        </div>
        {weakRows.length === 0 ? (
          <div className="empty">Aucun algorithme faible 🎉</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Tunnel</th>
                <th>Constat</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {weakRows.map((r, i) => (
                <tr key={i}>
                  <td>
                    <b>{r.t.name}</b>
                  </td>
                  <td className={/quantique|PFS/.test(r.note) ? "warn" : "down"}>{r.note}</td>
                  <td className="rowact">
                    {me?.can_write && (
                      <button className="btn xs pri" onClick={() => go("editor", { tunnel: r.t })}>
                        Corriger
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
