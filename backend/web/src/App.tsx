import { useEffect, useState } from "react";
import { api } from "./api";
import { useAuth } from "./auth";
import { NavCtx } from "./nav";
import { Login } from "./pages/Login";
import { Dashboard } from "./pages/Dashboard";
import { Tunnels } from "./pages/Tunnels";
import { Editor } from "./pages/Editor";
import { PKI } from "./pages/PKI";
import { Secrets } from "./pages/Secrets";
import { Audit } from "./pages/Audit";
import { Security } from "./pages/Security";
import { Gateways } from "./pages/Gateways";
import { Admin } from "./pages/Admin";
import { Crud } from "./pages/Crud";
import { Topo } from "./pages/Topo";
import { Daemon } from "./pages/Daemon";
import { AI } from "./pages/AI";
import { SCHEMAS } from "./schemas";

const ROLE_LABEL: Record<string, string> = {
  admin: "Administrateur",
  operator: "Opérateur",
  auditor: "Auditeur",
  viewer: "Lecture seule",
};

const NAV: [string, [string, string][]][] = [
  [
    "Supervision",
    [
      ["dash", "Tableau de bord"],
      ["tunnels", "Connexions"],
      ["topo", "Topologie"],
      ["monitoring", "Monitoring & alertes"],
      ["journaux", "Journal d'audit"],
    ],
  ],
  [
    "Configuration",
    [
      ["editor", "Éditeur de tunnel"],
      ["pki", "PKI & Certificats"],
      ["secrets", "Secrets"],
      ["users", "Utilisateurs VPN"],
    ],
  ],
  [
    "StrongSwan",
    [
      ["pools", "Pools & IP virtuelles"],
      ["radius", "RADIUS / AAA"],
      ["policies", "Politiques & routage"],
      ["authorities", "Autorités & enrôlement"],
      ["daemoncfg", "Paramètres du démon"],
    ],
  ],
  [
    "Avancé",
    [
      ["security", "Sécurité & Conformité"],
      ["ai", "Assistant & anomalies IA"],
      ["fleet", "Passerelles & ZTP"],
      ["admin", "Administration"],
    ],
  ],
];

const TITLES: Record<string, string> = Object.fromEntries(NAV.flatMap(([, items]) => items));

// Icônes (markup SVG interne, comme la maquette app.html).
const ICONS: Record<string, string> = {
  dash: '<path d="M3 3h7v9H3zM14 3h7v5h-7zM14 12h7v9h-7zM3 16h7v5H3z"/>',
  tunnels: '<path d="M3 12h18M6 12a3 3 0 013-3h6a3 3 0 013 3M6 12a3 3 0 003 3h6a3 3 0 003-3"/>',
  topo: '<circle cx="5" cy="6" r="2.5"/><circle cx="19" cy="6" r="2.5"/><circle cx="12" cy="18" r="2.5"/><path d="M6.8 7.4l3.6 8.4M17.2 7.4l-3.6 8.4M7 6h10"/>',
  monitoring: '<path d="M3 12h4l3 8 4-16 3 8h4"/>',
  journaux: '<path d="M4 6h16M4 12h16M4 18h10"/>',
  editor: '<path d="M12 20h9M16.5 3.5a2.1 2.1 0 013 3L7 19l-4 1 1-4z"/>',
  pki: '<circle cx="8" cy="8" r="4"/><path d="M11 11l7 7M15 15l2 2 3-1 1-3-2-2"/>',
  secrets: '<path d="M7 10V7a5 5 0 0110 0v3M5 10h14v11H5z"/>',
  users: '<path d="M16 21v-2a4 4 0 00-8 0v2M12 11a4 4 0 100-8 4 4 0 000 8"/>',
  pools: '<path d="M3 7h18v4H3zM3 13h18v4H3zM7 9h.01M7 15h.01"/>',
  radius: '<path d="M4 5h16v6H4zM4 13h16v6H4zM8 8h.01M8 16h.01"/>',
  policies: '<path d="M3 6h6l3 6 3-6h6M3 18h6l3-6"/>',
  authorities: '<path d="M12 3l8 3v6c0 5-3.4 8.5-8 10-4.6-1.5-8-5-8-10V6zM9 12l2 2 4-4"/>',
  daemoncfg: '<path d="M4 6h16M4 12h16M4 18h16"/>',
  security: '<path d="M12 3l8 3v6c0 5-3.4 8.5-8 10-4.6-1.5-8-5-8-10V6z"/>',
  ai: '<path d="M12 3a4 4 0 014 4c1.5.5 3 2 3 4a4 4 0 01-2 3.5M12 3a4 4 0 00-4 4c-1.5.5-3 2-3 4a4 4 0 002 3.5M9 21h6M10 21v-3M14 21v-3"/>',
  fleet: '<path d="M4 7h16v5H4zM4 15h16v5H4zM8 10h.01M8 18h.01"/>',
  admin: '<circle cx="12" cy="12" r="3"/><path d="M19.4 13a1.7 1.7 0 00.3 1.9M4.6 13a1.7 1.7 0 01-.3 1.9M12 4.6V2M12 22v-2.6"/>',
};

export function App() {
  const { me, ready, logout } = useAuth();
  const [page, setPage] = useState("dash");
  const [arg, setArg] = useState<any>(null);
  const [down, setDown] = useState(0);

  useEffect(() => {
    if (!me) return;
    let stop = false;
    const tick = () =>
      api
        .get("/tunnels")
        .then((r) => !stop && setDown((r.items || []).filter((t: any) => t.status === "down").length))
        .catch(() => {});
    tick();
    const t = setInterval(tick, 5000);
    return () => {
      stop = true;
      clearInterval(t);
    };
  }, [me]);

  if (!ready) return <div className="login-wrap">Chargement…</div>;
  if (!me) return <Login />;

  const go = (p: string, a: any = null) => {
    setArg(a);
    setPage(p);
  };

  const initials = me.identity.slice(0, 2).toUpperCase();

  function render() {
    switch (page) {
      case "dash":
        return <Dashboard />;
      case "tunnels":
        return <Tunnels />;
      case "editor":
        return <Editor />;
      case "pki":
        return <PKI />;
      case "secrets":
        return <Secrets />;
      case "journaux":
        return <Audit />;
      case "security":
        return <Security />;
      case "fleet":
        return <Gateways />;
      case "admin":
        return <Admin />;
      case "topo":
        return <Topo />;
      case "daemoncfg":
        return <Daemon />;
      case "ai":
        return <AI />;
      case "pools":
        return <Crud schema={SCHEMAS.pool} />;
      case "radius":
        return <Crud schema={SCHEMAS.radius} />;
      case "policies":
        return <Crud schema={SCHEMAS.policy} />;
      case "authorities":
        return <Crud schema={SCHEMAS.authority} />;
      case "users":
        return <Crud schema={SCHEMAS.vpnuser} />;
      case "monitoring":
        return <Crud schema={SCHEMAS.alert} />;
      default:
        return <div className="empty">Page inconnue</div>;
    }
  }

  return (
    <NavCtx.Provider value={{ page, arg, go }}>
      <div className="app">
        <aside className="side">
          <div className="brand">
            <svg className="mark" viewBox="0 0 32 32" fill="none">
              <rect x="1.5" y="1.5" width="29" height="29" rx="8" fill="var(--accent)" />
              <path d="M16 7l7 3v5c0 4.4-2.9 7.6-7 9-4.1-1.4-7-4.6-7-9v-5l7-3z" fill="none" stroke="#fff" strokeWidth="1.7" strokeLinejoin="round" />
              <path d="M13 16.2l2.2 2.2L20 13.4" stroke="#fff" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <div className="name">
              StrongSwan
              <br />
              <span>Manager</span>
            </div>
          </div>
          <nav>
            {NAV.map(([grp, items]) => (
              <div key={grp}>
                <div className="navlabel">{grp}</div>
                {items.map(([id, label]) => (
                  <button key={id} className={"navitem " + (id === page ? "active" : "")} onClick={() => go(id)}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" dangerouslySetInnerHTML={{ __html: ICONS[id] || "" }} />
                    {label}
                    {id === "tunnels" && down > 0 && <span className="badge d">{down}</span>}
                  </button>
                ))}
              </div>
            ))}
          </nav>
          <div className="side-foot">
            <div className="avatar">{initials}</div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ color: "var(--text)", fontWeight: 600 }}>{me.identity}</div>
              {ROLE_LABEL[me.role] || me.role}
            </div>
            <button className="btn xs ghost" onClick={logout} title="Se déconnecter">
              ⇥
            </button>
          </div>
        </aside>

        <div className="main">
          <header className="topbar">
            <div>
              <h1>{TITLES[page] || ""}</h1>
              <div className="crumb">
                {me.identity} · {ROLE_LABEL[me.role] || me.role}
              </div>
            </div>
            <div style={{ marginLeft: "auto", display: "flex", gap: 10 }}>
              <ThemeToggle />
              {me.can_write && (
                <button className="btn pri" onClick={() => go("editor")}>
                  + Nouveau tunnel
                </button>
              )}
            </div>
          </header>
          <div className="content">{render()}</div>
        </div>
      </div>
    </NavCtx.Provider>
  );
}

function ThemeToggle() {
  const [theme, setTheme] = useState<string>(() => document.documentElement.getAttribute("data-theme") || "");
  function toggle() {
    const cur = theme || (matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
    const next = cur === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    setTheme(next);
  }
  return (
    <button className="btn ghost" onClick={toggle} title="Thème clair/sombre">
      ◑
    </button>
  );
}
