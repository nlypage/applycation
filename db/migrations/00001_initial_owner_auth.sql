-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE FUNCTION applycation_set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE owners (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  password_hash text NOT NULL,
  password_changed_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE owner_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id uuid NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
  session_token_hash text NOT NULL UNIQUE,
  user_agent text,
  ip_address inet,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX owner_sessions_owner_id_idx ON owner_sessions(owner_id);
CREATE INDEX owner_sessions_expires_at_idx ON owner_sessions(expires_at);
CREATE INDEX owner_sessions_active_idx ON owner_sessions(owner_id) WHERE revoked_at IS NULL;

CREATE TABLE credentials (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id uuid NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
  provider text NOT NULL,
  subject text NOT NULL,
  ciphertext bytea NOT NULL,
  nonce bytea NOT NULL,
  algorithm text NOT NULL DEFAULT 'aes-256-gcm',
  key_version integer NOT NULL DEFAULT 1,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(owner_id, provider, subject)
);

CREATE INDEX credentials_owner_id_idx ON credentials(owner_id);
CREATE INDEX credentials_provider_subject_idx ON credentials(provider, subject);

CREATE TRIGGER owners_set_updated_at
BEFORE UPDATE ON owners
FOR EACH ROW
EXECUTE FUNCTION applycation_set_updated_at();

CREATE TRIGGER owner_sessions_set_updated_at
BEFORE UPDATE ON owner_sessions
FOR EACH ROW
EXECUTE FUNCTION applycation_set_updated_at();

CREATE TRIGGER credentials_set_updated_at
BEFORE UPDATE ON credentials
FOR EACH ROW
EXECUTE FUNCTION applycation_set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS credentials_set_updated_at ON credentials;
DROP TRIGGER IF EXISTS owner_sessions_set_updated_at ON owner_sessions;
DROP TRIGGER IF EXISTS owners_set_updated_at ON owners;

DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS owner_sessions;
DROP TABLE IF EXISTS owners;

DROP FUNCTION IF EXISTS applycation_set_updated_at();
