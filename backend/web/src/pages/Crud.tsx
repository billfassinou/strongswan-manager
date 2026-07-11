// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useEffect, useState } from "react";
import { api } from "../api";
import { Modal, Pill, useToast } from "../ui";
import { useAuth } from "../auth";
import { Schema, Field } from "../schemas";

interface Item {
  id: string;
  name: string;
  data: Record<string, any>;
}

// Page CRUD générique pilotée par un schéma : liste, création/édition (modale), suppression.
// Chaque entité est stockée côté backend dans config_items (kind + JSON).
export function Crud({ schema }: { schema: Schema }) {
  const { me } = useAuth();
  const toast = useToast();
  const canWrite = !!me?.can_write;
  const [items, setItems] = useState<Item[]>([]);
  const [edit, setEdit] = useState<Item | "new" | null>(null);

  async function load() {
    const r = await api.get(`/config/${schema.kind}`);
    setItems((r.items || []).map((it: any) => ({ id: it.id, name: it.name, data: it.data || {} })));
  }
  useEffect(() => {
    load();
  }, [schema.kind]);

  async function del(it: Item) {
    if (!confirm(`Supprimer « ${it.name} » ?`)) return;
    try {
      await api.del(`/config/${schema.kind}/${it.id}`);
      toast("Supprimé", "ok");
      load();
    } catch (e: any) {
      toast(e.message, "err");
    }
  }

  return (
    <div className="card">
      <div className="card-head">
        <h2>{schema.title}</h2>
        <div className="act muted" style={{ fontSize: 12, marginRight: "auto", marginLeft: 10 }}>
          {schema.sub}
        </div>
        <div className="act">
          <button className="btn sm pri" disabled={!canWrite} onClick={() => setEdit("new")}>
            {schema.add}
          </button>
        </div>
      </div>
      {items.length === 0 ? (
        <div className="empty">Aucun élément.</div>
      ) : (
        <table>
          <thead>
            <tr>
              {schema.fields.map((f) => (
                <th key={f.k}>{f.label}</th>
              ))}
              <th></th>
            </tr>
          </thead>
          <tbody>
            {items.map((it) => (
              <tr key={it.id}>
                {schema.fields.map((f, i) => (
                  <td key={f.k} className={i === 0 ? "" : "muted"}>
                    {renderCell(f, i === 0 ? it.name : it.data[f.k])}
                  </td>
                ))}
                <td className="rowact">
                  {canWrite && (
                    <>
                      <button className="btn xs ghost" onClick={() => setEdit(it)}>
                        Éditer
                      </button>
                      <button className="btn xs ghost down" onClick={() => del(it)}>
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
      {edit && (
        <EditModal
          schema={schema}
          item={edit === "new" ? null : edit}
          onClose={() => setEdit(null)}
          onDone={() => {
            setEdit(null);
            load();
          }}
        />
      )}
    </div>
  );
}

function renderCell(f: Field, v: any) {
  if (f.type === "toggle") return v ? <Pill status="up" text="oui" /> : <Pill status="x" text="non" />;
  if (f.k === "name") return <b>{v}</b>;
  return f.type === "select" ? <Pill status="a" text={String(v ?? "—")} /> : <span className="mono">{v ?? "—"}</span>;
}

function EditModal({
  schema,
  item,
  onClose,
  onDone,
}: {
  schema: Schema;
  item: Item | null;
  onClose: () => void;
  onDone: () => void;
}) {
  const toast = useToast();
  const [vals, setVals] = useState<Record<string, any>>(() => {
    const v: Record<string, any> = {};
    schema.fields.forEach((f) => {
      v[f.k] = item ? (f.k === "name" ? item.name : item.data[f.k]) : f.type === "toggle" ? true : f.type === "select" ? f.opts![0] : "";
    });
    return v;
  });
  const [busy, setBusy] = useState(false);
  const set = (k: string, v: any) => setVals((s) => ({ ...s, [k]: v }));

  async function submit() {
    if (!vals.name) {
      toast("Nom requis", "err");
      return;
    }
    setBusy(true);
    const data: Record<string, any> = {};
    schema.fields.forEach((f) => {
      if (f.k !== "name") data[f.k] = vals[f.k];
    });
    const body = { name: vals.name, data };
    try {
      if (item) await api.put(`/config/${schema.kind}/${item.id}`, body);
      else await api.post(`/config/${schema.kind}`, body);
      toast(item ? "Modifié" : "Créé", "ok");
      onDone();
    } catch (e: any) {
      toast(e.message, "err");
      setBusy(false);
    }
  }

  return (
    <Modal
      title={item ? "Éditer " + item.name : "Nouveau — " + schema.title}
      onClose={onClose}
      footer={
        <>
          <button className="btn ghost" onClick={onClose}>
            Annuler
          </button>
          <button className="btn pri" disabled={busy} onClick={submit}>
            Enregistrer
          </button>
        </>
      }
    >
      {schema.fields.map((f) => (
        <div key={f.k}>
          <label className="flabel">{f.label}</label>
          {f.type === "toggle" ? (
            <label className="row" style={{ gap: 8, fontSize: 13 }}>
              <input type="checkbox" checked={!!vals[f.k]} onChange={(e) => set(f.k, e.target.checked)} /> {vals[f.k] ? "activé" : "désactivé"}
            </label>
          ) : f.type === "select" ? (
            <select className="field" value={vals[f.k]} onChange={(e) => set(f.k, e.target.value)}>
              {f.opts!.map((o) => (
                <option key={o} value={o}>
                  {o}
                </option>
              ))}
            </select>
          ) : (
            <input className="field" value={vals[f.k] ?? ""} onChange={(e) => set(f.k, e.target.value)} />
          )}
        </div>
      ))}
    </Modal>
  );
}
