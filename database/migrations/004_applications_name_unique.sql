-- +goose Up
-- Application name becomes the label identity for normalization: API surfaces
-- carry the human-readable application name in place of the raw FK, resolved
-- through a Redis id<->label mapping. Uniqueness is the precondition for a name
-- to identify exactly one application.
ALTER TABLE applications ADD CONSTRAINT applications_name_key UNIQUE (name);

-- +goose Down
ALTER TABLE applications DROP CONSTRAINT applications_name_key;
