-- Initial schema. Runs automatically the first time the container starts.
-- (To re-run: docker compose down -v && docker compose up -d)

CREATE TABLE authors (
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    bio        TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE books (
    id               SERIAL PRIMARY KEY,
    author_id        INT         NOT NULL REFERENCES authors (id),
    title            TEXT        NOT NULL,
    isbn             TEXT        NOT NULL,
    published_year   INT,
    total_copies     INT         NOT NULL DEFAULT 1,
    available_copies INT         NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE members (
    id        SERIAL PRIMARY KEY,
    email     TEXT        NOT NULL UNIQUE,
    full_name TEXT        NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE loans (
    id          SERIAL PRIMARY KEY,
    book_id     INT         NOT NULL REFERENCES books (id),
    member_id   INT         NOT NULL REFERENCES members (id),
    borrowed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_at      TIMESTAMPTZ NOT NULL,
    returned_at TIMESTAMPTZ
);
