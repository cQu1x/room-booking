CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    capacity    INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS slots (
    id      UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_time   TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS slots_room_start_uidx ON slots(room_id, start_time);

CREATE TABLE IF NOT EXISTS bookings (
    id              UUID PRIMARY KEY,
    slot_id         UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('active', 'cancelled')) DEFAULT 'active',
    conference_link TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS bookings_slot_idx ON bookings(slot_id);