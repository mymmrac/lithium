CREATE TABLE "user"
(
    id       BIGINT PRIMARY KEY,
    email    TEXT NOT NULL,
    password TEXT NOT NULL
);

--bun:split

CREATE UNIQUE INDEX user_email ON "user" (email);
