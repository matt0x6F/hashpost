-- Initialize PDS database
-- This script runs after the databases are created

\c hashpost_pds_dev;

-- Run PDS migrations
\i /docker-entrypoint-initdb.d/001_initial_schema.sql;
