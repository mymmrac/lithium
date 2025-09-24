CREATE TABLE "user"
(
    id         BIGINT PRIMARY KEY,
    email      TEXT         NOT NULL,
    password   TEXT         NOT NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--bun:split

CREATE UNIQUE INDEX user_email ON "user" (email);
