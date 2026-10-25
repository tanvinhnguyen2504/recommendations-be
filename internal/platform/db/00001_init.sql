CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS gm_places (
  id BIGSERIAL PRIMARY KEY,
  place_key      TEXT         NOT NULL,
  name           TEXT         NOT NULL,
  address_line   TEXT,
  address        JSONB,
  district       TEXT,
  province       TEXT,
  latitude       NUMERIC(9,6),
  longitude      NUMERIC(9,6),
  phone_number   TEXT,
  website        TEXT,
  stars          NUMERIC(2,1),
  ratings        INT,
  price_level    TEXT,
  types          TEXT[]       NOT NULL DEFAULT '{}',
  open_hours     JSONB,
  images         JSONB,
  category       TEXT,                               -- food | drink, from the query that found it
  is_fnb         BOOLEAN      NOT NULL DEFAULT TRUE,
  first_seen_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_seen_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_task_id   BIGINT
);

CREATE INDEX IF NOT EXISTS gm_places_name_trgm_idx ON gm_places USING gin (name gin_trgm_ops);
