CREATE TABLE action
(
    id          BIGINT PRIMARY KEY,
    project_id  BIGINT        NOT NULL REFERENCES project (id) ON DELETE RESTRICT,
    name        TEXT          NOT NULL,
    path        TEXT          NOT NULL,
    methods     VARCHAR(32)[] NOT NULL,
    "order"     INT           NOT NULL,
    module_path TEXT          NOT NULL,
    config      JSONB         NOT NULL,
    created_at  TIMESTAMP(0)  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP(0)  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--bun:split

CREATE INDEX action_project_id ON action (project_id);
