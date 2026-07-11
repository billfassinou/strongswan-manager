-- Certificat TLS du serveur lui-même (auto-généré au 1er démarrage, signé par la CA interne).
-- Persisté en base — et non sur disque — pour que le binaire autonome, qui n'a pas de volume,
-- resserve le MÊME certificat après un redémarrage (sinon l'empreinte change à chaque fois et
-- l'administrateur revoit un avertissement navigateur).
-- La clé privée est chiffrée avec le même cipher applicatif que les secrets et la CA.
CREATE TABLE IF NOT EXISTS server_tls (
    id          TEXT PRIMARY KEY,
    cert_pem    TEXT        NOT NULL,
    key_enc     BYTEA       NOT NULL,
    sans        TEXT        NOT NULL DEFAULT '',
    not_after   TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
