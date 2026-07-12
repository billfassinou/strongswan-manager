// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useState } from "react";
import { login } from "../api";
import { useAuth } from "../auth";

export function Login() {
  const { refresh } = useAuth();
  const [identity, setIdentity] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setBusy(true);
    try {
      await login(identity, password);
      await refresh();
    } catch (ex: any) {
      setErr(ex.message || "Échec de connexion");
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
        <h1>Connexion à la console</h1>
        <p className="login-sub">Authentifiez-vous pour accéder au tableau de bord.</p>
        {err && <div className="err">{err}</div>}
        <label className="flabel">Identifiant</label>
        <input className="field" autoFocus autoComplete="username" value={identity} onChange={(e) => setIdentity(e.target.value)} />
        <label className="flabel" style={{ marginTop: 12 }}>
          Mot de passe
        </label>
        <input
          className="field"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button className="btn pri" disabled={busy} style={{ width: "100%", justifyContent: "center", marginTop: 16 }}>
          {busy ? "Connexion…" : "Se connecter"}
        </button>
      </form>
    </div>
  );
}
