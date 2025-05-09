CREATE TABLE events (
  id BIGSERIAL PRIMARY KEY,
  site_id BIGINT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
  action TEXT NOT NULL,
  count INT NOT NULL DEFAULT 1,
  device_type TEXT NOT NULL,
  os TEXT NOT NULL,
  browser TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
  updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_events_site_id ON events(site_id);
CREATE INDEX idx_events_created_at ON events(created_at);
