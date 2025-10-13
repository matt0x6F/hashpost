-- Grant permissions on existing tables
-- This script runs after tables are created

-- Grant permissions for PDS database
\c hashpost_pds_dev;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO hashpost_pds;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO hashpost_pds;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO hashpost_pds;

-- Grant permissions for AppView database
\c hashpost_appview_dev;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO hashpost_appview;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO hashpost_appview;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO hashpost_appview;
