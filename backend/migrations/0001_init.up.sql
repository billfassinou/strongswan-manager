-- Schéma initial (tranche verticale) — d'après §14 / §16.2 du cahier des charges.

CREATE TABLE IF NOT EXISTS users_admin (
    id         UUID PRIMARY KEY,
    identity   TEXT NOT NULL UNIQUE,
    pass_hash  TEXT NOT NULL,
    role       TEXT NOT NULL CHECK (role IN ('admin','operator','auditor','viewer')),
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS gateways (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    endpoint   TEXT NOT NULL,
    version    TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'unknown',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS secrets (
    id        UUID PRIMARY KEY,
    name      TEXT NOT NULL UNIQUE,
    type      TEXT NOT NULL,
    enc_value BYTEA NOT NULL,
    used_by   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS tunnels (
    id             UUID PRIMARY KEY,
    name           TEXT NOT NULL,
    gateway_id     UUID NOT NULL REFERENCES gateways(id) ON DELETE CASCADE,
    type           TEXT NOT NULL CHECK (type IN ('site-to-site','road-warrior','host-to-host')),
    ike_version    INT  NOT NULL CHECK (ike_version IN (1,2)),
    local_addr     TEXT NOT NULL DEFAULT '',
    remote_addr    TEXT NOT NULL DEFAULT '',
    local_subnets  TEXT[] NOT NULL DEFAULT '{}',
    remote_subnets TEXT[] NOT NULL DEFAULT '{}',
    auth_method    TEXT NOT NULL CHECK (auth_method IN ('psk','cert','eap')),
    secret_ref     TEXT,   -- référence un secret par son nom (voir table secrets.name)
    cert_ref       TEXT,   -- référence un certificat par son nom
    proposals_ike  TEXT[] NOT NULL DEFAULT '{}',
    proposals_esp  TEXT[] NOT NULL DEFAULT '{}',
    pfs            BOOLEAN NOT NULL DEFAULT TRUE,
    status         TEXT NOT NULL DEFAULT 'unknown',
    security_score INT NOT NULL DEFAULT 0,
    config_version INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (gateway_id, name)
);

CREATE TABLE IF NOT EXISTS config_versions (
    id         UUID PRIMARY KEY,
    tunnel_id  UUID NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
    n          INT  NOT NULL,
    author_id  UUID,
    message    TEXT NOT NULL DEFAULT '',
    snapshot   JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tunnel_id, n)
);

-- Journal d'audit append-only : les mises à jour et suppressions sont interdites
-- par un trigger, garantissant l'immuabilité (§9 / EF-11).
CREATE TABLE IF NOT EXISTS audit_log (
    id             UUID PRIMARY KEY,
    actor_id       UUID,
    action         TEXT NOT NULL,
    target         TEXT NOT NULL DEFAULT '',
    ts             TIMESTAMPTZ NOT NULL DEFAULT now(),
    prev_hash      TEXT NOT NULL DEFAULT '',
    integrity_hash TEXT NOT NULL
);

CREATE OR REPLACE FUNCTION audit_log_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'audit_log est append-only (opération % interdite)', TG_OP;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_log_no_update ON audit_log;
CREATE TRIGGER audit_log_no_update BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_immutable();
