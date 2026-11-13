-- Bootstrap for a fresh database, run by scripts/init-db.sh as the superuser.
-- Run once here rather than from a migration: creating roles needs CREATEROLE
-- and taking ownership of schema public needs to be its owner, and the
-- migrator role has neither. Migration 0001 then only manages objects it owns.
--
-- PostgreSQL 15 revoked CREATE on schema public from PUBLIC; making the
-- migrator the owner is what lets 0001's REVOKE actually take effect instead
-- of warning that no privilege could be revoked.

SELECT format('CREATE ROLE clover_migrator LOGIN PASSWORD %L', :'migrator_pw')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clover_migrator')\gexec

SELECT format('CREATE ROLE clover_app LOGIN PASSWORD %L', :'app_pw')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clover_app')\gexec

ALTER SCHEMA public OWNER TO clover_migrator;
