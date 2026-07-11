// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useEffect, useState } from "react";
import { api, AuditEntry } from "../api";

export function Audit() {
  const [items, setItems] = useState<AuditEntry[]>([]);
  async function load() {
    setItems((await api.get("/audit?limit=100")).items || []);
  }
  useEffect(() => {
    load();
    const t = setInterval(load, 5000);
    return () => clearInterval(t);
  }, []);

  return (
    <div className="card">
      <div className="card-head">
        <h2>Journal d'audit</h2>
        <div className="act muted" style={{ fontSize: 12 }}>
          append-only · chaîné (intégrité)
        </div>
      </div>
      {items.length === 0 ? (
        <div className="empty">Aucune entrée.</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Horodatage</th>
              <th>Action</th>
              <th>Cible</th>
            </tr>
          </thead>
          <tbody>
            {items.map((e) => (
              <tr key={e.id}>
                <td className="mono muted">{new Date(e.timestamp).toLocaleString("fr-FR")}</td>
                <td>
                  <span className="pill a">{e.action}</span>
                </td>
                <td className="mono">{e.target || "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
