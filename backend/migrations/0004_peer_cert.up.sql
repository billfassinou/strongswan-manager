-- Certificat du pair pour un site-à-site authentifié par certificats.
ALTER TABLE tunnels ADD COLUMN IF NOT EXISTS peer_cert_ref TEXT;
