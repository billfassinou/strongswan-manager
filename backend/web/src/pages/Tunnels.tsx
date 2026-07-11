// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useEffect, useState } from "react";
import { api, Tunnel, Gateway } from "../api";
import { Pill, useToast } from "../ui";
import { scoreColor } from "../score";
import { useAuth } from "../auth";
import { useNav } from "../nav";
import { useLiveStatus } from "../ws";

export function Tunnels() {
  const { me } = useAuth();
  const { go } = useNav();
  const toast = useToast();
  const canWrite = !!me?.can_write;
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const { last } = useLiveStatus();

  async function load() {
    const [t, g] = await Promise.all([api.get("/tunnels"), api.get("/gateways")]);
    setTunnels(t.items || []);
    setGateways(g.items || []);
  }
  useEffect(() => {
    load();
  }, []);
  useEffect(() => {
    if (last) setTunnels((ts) => ts.map((t) => (t.id === last.id ? { ...t, status: last.status } : t)));
  }, [last]);

  async function action(t: Tunnel, kind: "initiate" | "terminate" | "rekey") {
    try {
      await api.post(`/tunnels/${t.id}/${kind}`);
      toast(`${t.name} : ${kind}`, "ok");
      setTimeout(load, 1500);
    } catch (e: any) {
      toast(e.message, "err");
    }
  }
  async function del(t: Tunnel) {
    if (!confirm(`Supprimer le tunnel ${t.name} ?`)) return;
    try {
      await api.del(`/tunnels/${t.id}`);
      toast("Tunnel supprimé", "ok");
      load();
    } catch (e: any) {
      toast(e.message, "err");
    }
  }

  return (
    <div className="card">
      <div className="card-head">
        <h2>Connexions</h2>
        <div className="act">
          <button className="btn sm pri" disabled={!canWrite} onClick={() => go("editor")}>
            + Nouveau tunnel
          </button>
        </div>
      </div>
      {tunnels.length === 0 ? (
        <div className="empty">Aucun tunnel. {canWrite && "Créez-en un avec « Nouveau tunnel »."}</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Nom</th>
              <th>Type</th>
              <th>Passerelle</th>
              <th>Auth</th>
              <th>État</th>
              <th className="num">Score</th>
              <th></th>
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
                <td className="mono muted">{gwName(gateways, t.gateway_id)}</td>
                <td className="muted">{t.auth_method}</td>
                <td>
                  <Pill status={t.status} />
                </td>
                <td className="num">
                  <b className={scoreColor(t.security_score)}>{t.security_score}</b>
                </td>
                <td className="rowact">
                  {canWrite && (
                    <>
                      <button className="btn xs ghost" onClick={() => action(t, "initiate")}>
                        Monter
                      </button>
                      <button className="btn xs ghost" onClick={() => action(t, "terminate")}>
                        Couper
                      </button>
                      <button className="btn xs ghost" onClick={() => go("editor", { tunnel: t })}>
                        Éditer
                      </button>
                      <button className="btn xs ghost down" onClick={() => del(t)}>
                        ✕
                      </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

function gwName(gws: Gateway[], id: string): string {
  return gws.find((g) => g.id === id)?.name || id.slice(0, 8);
}
