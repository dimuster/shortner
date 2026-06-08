CREATE TABLE IF NOT EXISTS links (
                                     id          BIGSERIAL PRIMARY KEY,
                                     short_code  VARCHAR(16)  NOT NULL UNIQUE,
    original_url TEXT        NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_links_short_code ON links (short_code);

-- Seed data for manual testing
INSERT INTO links (short_code, original_url) VALUES
                                                 ('gh',    'https://github.com'),
                                                 ('yt',    'https://youtube.com'),
                                                 ('ggl',   'https://google.com')
    ON CONFLICT DO NOTHING;
