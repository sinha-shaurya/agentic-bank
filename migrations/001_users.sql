CREATE TABLE IF NOT EXISTS users
    (
        id           BIGINT GENERATED ALWAYS AS IDENTITY,
        token_sha256 TEXT NOT NULL UNIQUE               ,
        created_at   TIMESTAMPTZ NOT NULL DEFAULT now() ,
        updated_at   TIMESTAMPTZ NOT NULL DEFAULT now() ,
        active       BOOL
    );
