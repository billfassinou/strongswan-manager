DROP TRIGGER IF EXISTS audit_log_no_update ON audit_log;
DROP FUNCTION IF EXISTS audit_log_immutable();
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS tunnels;
DROP TABLE IF EXISTS secrets;
DROP TABLE IF EXISTS gateways;
DROP TABLE IF EXISTS users_admin;
