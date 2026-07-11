import { useEffect, useState } from "react";
import { api, Gateway } from "../api";
import { Pill } from "../ui";

export function Gateways() {
  const [items, setItems] = useState<Gateway[]>([]);
  useEffect(() => {
    api.get("/gateways").then((r) => setItems(r.items || []));
  }, []);

  return (
    <div>
      <div className="grid cols-4 mb">
        <Stat lbl="Passerelles" val={items.length} />
        <Stat lbl="En ligne" val={items.filter((g) => g.status !== "down").length} cls="up" />
        <Stat lbl="Injoignables" val={items.filter((g) => g.status === "down").length} cls="down" />
        <Stat lbl="≥ 6.0" val={items.filter((g) => g.version.startsWith("6")).length} />
      </div>
      <div className="card">
        <div className="card-head">
          <h2>Parc de passerelles</h2>
        </div>
        <table>
          <thead>
            <tr>
              <th>Nom</th>
              <th>Endpoint VICI</th>
              <th>Version</th>
              <th>État</th>
            </tr>
          </thead>
          <tbody>
            {items.map((g) => (
              <tr key={g.id}>
                <td>
                  <b>{g.name}</b>
                </td>
                <td className="mono muted">{g.endpoint}</td>
                <td className={"mono " + (g.version.startsWith("5") ? "warn" : "")}>{g.version || "—"}</td>
                <td>
                  <Pill status={g.status} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
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
