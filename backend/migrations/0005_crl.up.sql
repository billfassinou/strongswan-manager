-- Liste de révocation (CRL) portée par la CA, et horodatage de révocation des certificats.
ALTER TABLE cert_authorities ADD COLUMN IF NOT EXISTS crl_number BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cert_authorities ADD COLUMN IF NOT EXISTS crl_pem TEXT NOT NULL DEFAULT '';
ALTER TABLE certificates ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;
