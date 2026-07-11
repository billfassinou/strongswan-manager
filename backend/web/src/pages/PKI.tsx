import { useEffect, useState } from "react";
import { api, Certificate } from "../api";
import { Pill, Modal, useToast } from "../ui";
import { useAuth } from "../auth";

export function PKI() {
  const { me } = useAuth();
  const toast = useToast();
  const canWrite = !!me?.can_write;
  const [ca, setCa] = useState<any>(null);
  const [certs, setCerts] = useState<Certificate[]>([]);
  const [gen, setGen] = useState(false);

  async function load() {
    const [c, l] = await Promise.all([api.get("/ca").catch(() => null), api.get("/certificates")]);
    setCa(c);
    setCerts(l.items || []);
  }
  useEffect(() => {
    load();
  }, []);

  async function revoke(c: Certificate) {
    if (!confirm(`Révoquer le certificat ${c.name} ? La CRL sera régénérée.`)) return;
    try {
      await api.post(`/certificates/${c.id}/revoke`);
      toast("Certificat révoqué · CRL régénérée", "warn");
      load();
    } catch (e: any) {
      toast(e.message, "err");
    }
  }
  async function publishCRL() {
    try {
      const r = await api.post("/crl/publish");
      toast(`CRL publiée (n°${r.crl_number})`, "ok");
    } catch (e: any) {
      toast(e.message, "err");
    }
  }

  const expiring = certs.filter((c) => c.status === "valid" && daysLeft(c.not_after) < 90).length;

  return (
    <div>
      <div className="grid cols-4 mb">
        <Stat lbl="Certificats" val={certs.length} />
        <Stat lbl="Valides" val={certs.filter((c) => c.status === "valid").length} cls="up" />
        <Stat lbl="Révoqués" val={certs.filter((c) => c.status === "revoked").length} cls="down" />
        <Stat lbl="Expirent < 90 j" val={expiring} cls="warn" />
      </div>

      <div className="grid" style={{ gridTemplateColumns: "1fr 1fr", gap: 16 }}>
        <div className="card">
          <div className="card-head">
            <h2>Certificats émis</h2>
            <div className="act">
              <button className="btn sm pri" disabled={!canWrite} onClick={() => setGen(true)}>
                Générer
              </button>
            </div>
          </div>
          {certs.length === 0 ? (
            <div className="empty">Aucun certificat.</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Nom</th>
                  <th>CN</th>
                  <th>État</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {certs.map((c) => (
                  <tr key={c.id}>
                    <td>
                      <b>{c.name}</b>
                      <div className="muted mono" style={{ fontSize: 11 }}>
                        {c.kind} · {c.serial.slice(0, 12)}…
                      </div>
                    </td>
                    <td className="mono muted">{c.cn}</td>
                    <td>
                      {c.status === "revoked" ? (
                        <Pill status="down" text="Révoqué" />
                      ) : daysLeft(c.not_after) < 90 ? (
                        <Pill status="negotiating" text={`J-${daysLeft(c.not_after)}`} />
                      ) : (
                        <Pill status="up" text="Valide" />
                      )}
                    </td>
                    <td className="rowact">
                      {canWrite && c.status !== "revoked" && (
                        <button className="btn xs ghost down" onClick={() => revoke(c)}>
                          Révoquer
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <div className="card" style={{ padding: 18 }}>
            <h3 style={{ marginBottom: 10 }}>Autorité de certification interne</h3>
            <div className="row" style={{ gap: 12, padding: 12, background: "var(--surface-2)", borderRadius: 9 }}>
              <span style={{ fontSize: 20 }}>🛡️</span>
              <div style={{ flex: 1 }}>
                <b>{ca?.name || "—"}</b>
                <div className="muted mono" style={{ fontSize: 11 }}>
                  ECDSA · générée au démarrage
                </div>
              </div>
              <Pill status="up" text="active" />
            </div>
            <div className="row mt" style={{ gap: 9 }}>
              <button className="btn sm" disabled={!canWrite} onClick={publishCRL}>
                Publier la CRL
              </button>
              <a className="btn sm ghost" href="/crl.der" target="_blank" rel="noreferrer">
                Télécharger la CRL (.der)
              </a>
            </div>
          </div>
          <div className="card" style={{ padding: 16, fontSize: 13, color: "var(--text-dim)" }}>
            La révocation passe par le <b>CRL Distribution Point</b> inscrit dans les certificats (récupéré par les
            passerelles via le plugin curl de charon).
          </div>
        </div>
      </div>

      {gen && (
        <GenCert
          onClose={() => setGen(false)}
          onDone={() => {
            setGen(false);
            load();
          }}
        />
      )}
    </div>
  );
}

function GenCert({ onClose, onDone }: { onClose: () => void; onDone: () => void }) {
  const toast = useToast();
  const [name, setName] = useState("gw-a-cert");
  const [cn, setCn] = useState("gw-a");
  const [kind, setKind] = useState("server");
  const [san, setSan] = useState("203.0.113.10");
  const [busy, setBusy] = useState(false);
  async function submit() {
    setBusy(true);
    try {
      await api.post("/certificates", { name, cn, kind, sans: san.split(",").map((s) => s.trim()).filter(Boolean) });
      toast("Certificat généré", "ok");
      onDone();
    } catch (e: any) {
      toast(e.message, "err");
      setBusy(false);
    }
  }
  return (
    <Modal
      title="Générer un certificat"
      onClose={onClose}
      footer={
        <>
          <button className="btn ghost" onClick={onClose}>
            Annuler
          </button>
          <button className="btn pri" disabled={busy} onClick={submit}>
            Générer
          </button>
        </>
      }
    >
      <L l="Nom">
        <input className="field" value={name} onChange={(e) => setName(e.target.value)} />
      </L>
      <L l="Common Name (CN)">
        <input className="field" value={cn} onChange={(e) => setCn(e.target.value)} />
      </L>
      <L l="Usage">
        <select className="field" value={kind} onChange={(e) => setKind(e.target.value)}>
          <option value="server">Serveur</option>
          <option value="client">Client</option>
        </select>
      </L>
      <L l="SAN (IP/DNS, séparés par des virgules)">
        <input className="field mono" value={san} onChange={(e) => setSan(e.target.value)} />
      </L>
    </Modal>
  );
}

function L({ l, children }: { l: string; children: any }) {
  return (
    <div>
      <label className="flabel">{l}</label>
      {children}
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
function daysLeft(iso: string): number {
  return Math.round((new Date(iso).getTime() - Date.now()) / 86400000);
}
