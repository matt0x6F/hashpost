-- Initialize separate databases for PDS and AppView
-- This script runs when PostgreSQL container starts

-- Create PDS database and user
CREATE DATABASE hashpost_pds_dev;
CREATE USER hashpost_pds WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE hashpost_pds_dev TO hashpost_pds;
GRANT ALL PRIVILEGES ON SCHEMA public TO hashpost_pds;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO hashpost_pds;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO hashpost_pds;

-- Create AppView database and user  
CREATE DATABASE hashpost_appview_dev;
CREATE USER hashpost_appview WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE hashpost_appview_dev TO hashpost_appview;
GRANT ALL PRIVILEGES ON SCHEMA public TO hashpost_appview;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO hashpost_appview;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO hashpost_appview;

-- Enable UUID extension for both databases
\c hashpost_pds_dev;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c hashpost_appview_dev;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Run database-specific migrations
\i /docker-entrypoint-initdb.d/init-pds-database.sql;
\i /docker-entrypoint-initdb.d/init-appview-database.sql;

-- Grant permissions on existing tables
\i /docker-entrypoint-initdb.d/grant-permissions.sql;
