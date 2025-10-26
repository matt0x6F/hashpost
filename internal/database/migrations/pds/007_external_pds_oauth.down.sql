-- Rollback External PDS OAuth Support

DROP INDEX IF EXISTS idx_external_pds_clients_endpoint;
DROP TABLE IF EXISTS external_pds_clients;
