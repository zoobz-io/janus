ALTER TABLE user_applications ADD COLUMN roles TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE user_applications ADD COLUMN scopes TEXT[] NOT NULL DEFAULT '{}';
