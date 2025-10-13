-- Initialize AppView database
-- This script runs after the databases are created

\c hashpost_appview_dev;

-- Run AppView migrations
\i /docker-entrypoint-initdb.d/001_appview_schema.sql;
