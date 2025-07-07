-- Silence Project Database Initialization Script
-- This script creates all necessary databases for the Silence project services

-- Enable logging
\set ON_ERROR_STOP on

-- Create databases for all services (only if they don't exist)
SELECT 'CREATE DATABASE silence_auth OWNER postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'silence_auth')\gexec

SELECT 'CREATE DATABASE silence_server_manager OWNER postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'silence_server_manager')\gexec

SELECT 'CREATE DATABASE silence_vpn OWNER postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'silence_vpn')\gexec

SELECT 'CREATE DATABASE silence_analytics OWNER postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'silence_analytics')\gexec

-- Create roles for different services (optional, for better security)
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'silence_auth_user') THEN
      CREATE ROLE silence_auth_user WITH LOGIN PASSWORD 'auth_password';
   END IF;
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'silence_server_manager_user') THEN
      CREATE ROLE silence_server_manager_user WITH LOGIN PASSWORD 'server_manager_password';
   END IF;
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'silence_vpn_user') THEN
      CREATE ROLE silence_vpn_user WITH LOGIN PASSWORD 'vpn_password';
   END IF;
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'silence_analytics_user') THEN
      CREATE ROLE silence_analytics_user WITH LOGIN PASSWORD 'analytics_password';
   END IF;
END
$do$;

-- Grant privileges to service roles
GRANT ALL PRIVILEGES ON DATABASE silence_auth TO silence_auth_user;
GRANT ALL PRIVILEGES ON DATABASE silence_server_manager TO silence_server_manager_user;
GRANT ALL PRIVILEGES ON DATABASE silence_vpn TO silence_vpn_user;
GRANT ALL PRIVILEGES ON DATABASE silence_analytics TO silence_analytics_user;

-- Grant connect privileges to postgres user (for migrations)
GRANT ALL PRIVILEGES ON DATABASE silence_auth TO postgres;
GRANT ALL PRIVILEGES ON DATABASE silence_server_manager TO postgres;
GRANT ALL PRIVILEGES ON DATABASE silence_vpn TO postgres;
GRANT ALL PRIVILEGES ON DATABASE silence_analytics TO postgres;

-- Create extensions that might be needed
\c silence_auth;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c silence_server_manager;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

\c silence_vpn;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

\c silence_analytics;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Log completion
\echo 'Database initialization completed successfully'
