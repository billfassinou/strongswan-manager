-- Stockage générique des entités de configuration (pools, RADIUS, politiques, autorités,
-- utilisateurs VPN, règles d'alerte, paramètres du démon…). Chaque item porte un `kind`
-- et ses champs en JSONB, ce qui évite une table par module tout en restant requêtable.
CREATE TABLE IF NOT EXISTS config_items (
    id         UUID PRIMARY KEY,
    kind       TEXT NOT NULL,
    name       TEXT NOT NULL,
    data       JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, name)
);
CREATE INDEX IF NOT EXISTS config_items_kind_idx ON config_items (kind);
