ALTER TABLE certificates DROP COLUMN IF EXISTS revoked_at;
ALTER TABLE cert_authorities DROP COLUMN IF EXISTS crl_pem;
ALTER TABLE cert_authorities DROP COLUMN IF EXISTS crl_number;
