CREATE TABLE project
(
    id         BIGINT PRIMARY KEY,
    owner_id   BIGINT       NOT NULL REFERENCES "user" (id) ON DELETE RESTRICT,
    name       TEXT         NOT NULL,
    sub_domain TEXT         NOT NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--bun:split

CREATE INDEX project_owner_id ON project (owner_id);

--bun:split

CREATE UNIQUE INDEX project_sub_domain ON project (sub_domain);
