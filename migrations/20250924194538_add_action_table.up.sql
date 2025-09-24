CREATE TABLE action
(
    id          BIGINT PRIMARY KEY,
    project_id  BIGINT       NOT NULL REFERENCES project (id) ON DELETE CASCADE,
    name        TEXT         NOT NULL,
    url         TEXT         NOT NULL,
    module_path TEXT         NOT NULL,
    created_at  TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--bun:split

CREATE INDEX action_project_id ON action (project_id);

--bun:split

CREATE UNIQUE INDEX action_project_id_and_url ON action (project_id, url);
