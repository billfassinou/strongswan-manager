import { useAuth } from "../auth";

const ROLES: [string, string, string][] = [
  ["Administrateur", "admin", "Accès total (lecture + écriture)"],
  ["Opérateur", "operator", "Configuration & exploitation (écriture)"],
  ["Auditeur", "auditor", "Lecture seule + audit/conformité"],
  ["Lecture seule", "viewer", "Supervision uniquement"],
];

export function Admin() {
  const { me } = useAuth();
  return (
    <div>
      <div className="card mb">
        <div className="card-head">
          <h2>Session courante</h2>
        </div>
        <div style={{ padding: 18 }}>
          <div className="row" style={{ gap: 12 }}>
            <div className="avatar">{me?.identity.slice(0, 2).toUpperCase()}</div>
            <div>
              <b>{me?.identity}</b>
              <div className="muted">
                Rôle : {me?.role} · écriture : {me?.can_write ? "oui" : "non (lecture seule)"}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="card mb">
        <div className="card-head">
          <h2>Rôles &amp; permissions (RBAC)</h2>
        </div>
        <table>
          <thead>
            <tr>
              <th>Rôle</th>
              <th>Identifiant</th>
              <th>Permissions</th>
            </tr>
          </thead>
          <tbody>
            {ROLES.map((r) => (
              <tr key={r[1]}>
                <td>
                  <b>{r[0]}</b>
                </td>
                <td className="mono muted">{r[1]}</td>
                <td className="muted">{r[2]}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <div style={{ padding: "12px 18px" }} className="muted">
          <span style={{ fontSize: 12.5 }}>
            La gestion des comptes (création/désactivation), le multi-tenant et le SSO SAML/OIDC relèvent de l'édition
            Enterprise — non encore exposés par l'API.
          </span>
        </div>
      </div>

      <div className="card" style={{ padding: 18 }}>
        <h3 style={{ marginBottom: 10 }}>API &amp; documentation</h3>
        <div className="row" style={{ gap: 9 }}>
          <a className="btn sm" href="/api/v1/docs" target="_blank" rel="noreferrer">
            OpenAPI (Swagger)
          </a>
          <a className="btn sm ghost" href="/metrics" target="_blank" rel="noreferrer">
            Métriques Prometheus
          </a>
        </div>
      </div>
    </div>
  );
}
