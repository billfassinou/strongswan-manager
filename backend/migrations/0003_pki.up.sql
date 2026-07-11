-- PKI interne : autorité de certification et certificats X.509 (EF-04).

CREATE TABLE IF NOT EXISTS cert_authorities (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    cert_pem   TEXT NOT NULL,
    key_enc    BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS certificates (
    id         UUID PRIMARY KEY,
    ca_id      UUID REFERENCES cert_authorities(id) ON DELETE SET NULL,
    name       TEXT NOT NULL UNIQUE,
    cn         TEXT NOT NULL,
    kind       TEXT NOT NULL DEFAULT 'server',
    serial     TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'valid' CHECK (status IN ('valid','revoked','expired')),
    not_before TIMESTAMPTZ NOT NULL,
    not_after  TIMESTAMPTZ NOT NULL,
    cert_pem   TEXT NOT NULL,
    key_enc    BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
