-- +goose Up
CREATE TABLE linked_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    external_subject TEXT NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, external_subject)
);

CREATE INDEX idx_linked_identities_user_id ON linked_identities(user_id);

-- +goose Down
DROP TABLE linked_identities;
