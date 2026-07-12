// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useState } from "react";
import { changePassword } from "../api";
import { useAuth } from "../auth";

const MIN = 12;

// Écran bloquant présenté tant que le compte utilise le mot de passe posé à l'installation.
// Ce n'est pas qu'une politesse d'interface : le backend refuse (403) tout autre appel
// d'API dans cet état. Il n'y a donc rien à contourner ici.
export function ChangePassword() {
  const { me, refresh, logout } = useAuth();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  const tooShort = next.length > 0 && next.length < MIN;
  const mismatch = confirm.length > 0 && next !== confirm;
  const ok = current && next.length >= MIN && next === confirm;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setBusy(true);
    try {
      await changePassword(current, next);
      await refresh();
    } catch (ex: any) {
      const details = ex?.body?.details?.map((d: any) => d.issue).join(", ");
      setErr(details || ex.message || "Changement refusé");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="login-wrap">
      <form className="login-card" onSubmit={submit}>
        <div className="brand" style={{ padding: 0, marginBottom: 18 }}>
          🛡️ StrongSwan <span>Manager</span>
        </div>
        <h1>Choisissez un mot de passe</h1>
        <p className="login-sub">
          Le compte <b>{me?.identity}</b> utilise encore le mot de passe créé à l'installation. Il
          figure en clair dans le fichier de configuration du serveur : changez-le pour ouvrir la
          console.
        </p>
        {err && <div className="err">{err}</div>}

        <label className="flabel">Mot de passe actuel</label>
        <input
          className="field"
          type="password"
          autoFocus
          value={current}
          onChange={(e) => setCurrent(e.target.value)}
        />

        <label className="flabel" style={{ marginTop: 12 }}>
          Nouveau mot de passe
        </label>
        <input
          className="field"
          type="password"
          value={next}
          onChange={(e) => setNext(e.target.value)}
        />
        {tooShort && <div className="err">Au moins {MIN} caractères.</div>}

        <label className="flabel" style={{ marginTop: 12 }}>
          Confirmation
        </label>
        <input
          className="field"
          type="password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
        />
        {mismatch && <div className="err">Les deux saisies diffèrent.</div>}

        <button
          className="btn pri"
          disabled={busy || !ok}
          style={{ width: "100%", justifyContent: "center", marginTop: 16 }}
        >
          {busy ? "Enregistrement…" : "Changer le mot de passe"}
        </button>

        <p className="muted" style={{ fontSize: 11.5, textAlign: "center", marginTop: 14 }}>
          Les autres comptes livrés (operator, auditor, viewer) partagent ce mot de passe : pensez à
          les traiter aussi, ou à les désactiver.{" "}
          <a href="#" onClick={(e) => (e.preventDefault(), logout())}>
            Se déconnecter
          </a>
        </p>
      </form>
    </div>
  );
}
