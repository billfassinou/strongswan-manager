import { useEffect, useRef, useState } from "react";
import { api, Tunnel } from "../api";

// Détection d'anomalies dérivée de l'état réel des tunnels + assistant de diagnostic à
// base de règles (opère sur les données réelles ; pas de LLM — honnête sur sa nature).
export function AI() {
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  const [chat, setChat] = useState<{ who: "ai" | "me"; text: string }[]>([
    { who: "ai", text: "Assistant de diagnostic (règles). Demandez par ex. « pourquoi <tunnel> est down ? » ou « quels tunnels sont à durcir ? »." },
  ]);
  const [input, setInput] = useState("");
  const chatRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    api.get("/tunnels").then((r) => setTunnels(r.items || []));
  }, []);
  useEffect(() => {
    chatRef.current?.scrollTo(0, chatRef.current.scrollHeight);
  }, [chat]);

  const anomalies = [
    ...tunnels.filter((t) => t.status === "down").map((t) => ({ t, sev: "d", msg: "Tunnel down — pair injoignable ou négociation échouée" })),
    ...tunnels.filter((t) => t.status === "negotiating").map((t) => ({ t, sev: "n", msg: "Négociation prolongée (rekey / auth)" })),
    ...tunnels.filter((t) => t.security_score < 65).map((t) => ({ t, sev: "n", msg: `Configuration faible (score ${t.security_score})` })),
  ];

  function answer(q: string): string {
    const low = q.toLowerCase();
    const named = tunnels.find((t) => low.includes(t.name.toLowerCase()));
    if (named) {
      if (named.status === "down") return `Le tunnel « ${named.name} » est down : vérifiez la joignabilité du pair (UDP 500/4500), l'horloge NTP et la correspondance des propositions/PSK. Score de sécurité ${named.security_score}/100.`;
      if (named.status === "negotiating") return `« ${named.name} » est en négociation : un échange IKE/rekey est en cours. Transitoire ; s'il persiste, contrôlez les propositions cryptographiques et l'authentification.`;
      return `« ${named.name} » est actif (${named.status}), score ${named.security_score}/100.`;
    }
    if (/faible|durcir|score|algo/.test(low)) {
      const weak = tunnels.filter((t) => t.security_score < 65).map((t) => t.name);
      return weak.length ? `Tunnels à durcir : ${weak.join(", ")}. Migrez vers IKEv2 + AES-GCM + PFS ecp384 et ajoutez ML-KEM.` : "Aucun tunnel critique — parc conforme.";
    }
    if (/down|panne|probl/.test(low)) {
      const d = tunnels.filter((t) => t.status === "down").map((t) => t.name);
      return d.length ? `Tunnels down : ${d.join(", ")}.` : "Aucun tunnel down actuellement.";
    }
    return "Je peux diagnostiquer un tunnel (nommez-le), lister les tunnels down ou ceux à durcir. Reformulez avec un nom de tunnel.";
  }

  function send(text?: string) {
    const q = (text ?? input).trim();
    if (!q) return;
    setChat((c) => [...c, { who: "me", text: q }]);
    setInput("");
    setTimeout(() => setChat((c) => [...c, { who: "ai", text: answer(q) }]), 300);
  }

  return (
    <div className="grid" style={{ gridTemplateColumns: "1fr 1fr", gap: 16 }}>
      <div className="card">
        <div className="card-head">
          <h2>Détection d'anomalies</h2>
          <div className="act muted" style={{ fontSize: 12 }}>
            dérivée de l'état réel
          </div>
        </div>
        {anomalies.length === 0 ? (
          <div className="empty">Aucune anomalie détectée 🎉</div>
        ) : (
          anomalies.map((a, i) => (
            <div key={i} className="row" style={{ padding: "12px 18px", borderBottom: "1px solid var(--border)" }}>
              <span className={"dot " + a.sev}></span>
              <div style={{ flex: 1 }}>
                <b>{a.t.name}</b>
                <div className="muted" style={{ fontSize: 12 }}>
                  {a.msg}
                </div>
              </div>
              <span className={"pill " + a.sev}>{a.sev === "d" ? "critique" : "à surveiller"}</span>
            </div>
          ))
        )}
      </div>

      <div className="card" style={{ display: "flex", flexDirection: "column" }}>
        <div className="card-head">
          <h2>Assistant de diagnostic</h2>
        </div>
        <div ref={chatRef} style={{ padding: 18, display: "flex", flexDirection: "column", gap: 12, maxHeight: 380, overflowY: "auto" }}>
          {chat.map((m, i) => (
            <div
              key={i}
              style={{
                maxWidth: "80%",
                padding: "11px 14px",
                borderRadius: 14,
                fontSize: 13.5,
                alignSelf: m.who === "me" ? "flex-end" : "flex-start",
                background: m.who === "me" ? "var(--accent)" : "var(--surface-2)",
                color: m.who === "me" ? "#fff" : "var(--text)",
                border: m.who === "me" ? "none" : "1px solid var(--border)",
              }}
            >
              {m.text}
            </div>
          ))}
        </div>
        <div className="row" style={{ padding: "14px 18px", borderTop: "1px solid var(--border)", gap: 10 }}>
          <input className="field" value={input} placeholder="Ex : pourquoi paris-dakar est down ?" onChange={(e) => setInput(e.target.value)} onKeyDown={(e) => e.key === "Enter" && send()} />
          <button className="btn pri" onClick={() => send()}>
            Envoyer
          </button>
        </div>
      </div>
    </div>
  );
}
