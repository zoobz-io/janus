CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO config (key, value) VALUES ('aperture_schema', '
metrics:
  - signal: "janus.session.created"
    name: "janus_sessions_created_total"
    type: counter
  - signal: "janus.session.revoked"
    name: "janus_sessions_revoked_total"
    type: counter
  - signal: "janus.user.created"
    name: "janus_users_created_total"
    type: counter
  - signal: "janus.tenant.created"
    name: "janus_tenants_created_total"
    type: counter
  - signal: "janus.identity.linked"
    name: "janus_identities_linked_total"
    type: counter
  - signal: "janus.tenant.app.authorized"
    name: "janus_tenant_apps_authorized_total"
    type: counter
  - signal: "janus.user.app.granted"
    name: "janus_user_apps_granted_total"
    type: counter
  - signal: "janus.member.added"
    name: "janus_members_added_total"
    type: counter
  - signal: "janus.member.removed"
    name: "janus_members_removed_total"
    type: counter

logs:
  whitelist: []

context:
  logs: ["user_id", "tenant_id"]
  traces: ["user_id", "tenant_id"]
') ON CONFLICT (key) DO NOTHING;
