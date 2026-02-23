CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
    IF NOT EXISTS urls (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        original_url TEXT NOT NULL,
        short_code VARCHAR(10) NOT NULL UNIQUE,
        created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            expires_at TIMESTAMP
        WITH
            TIME ZONE,
            is_active BOOLEAN DEFAULT TRUE
    );

CREATE INDEX idx_urls_short_code ON urls (short_code);

CREATE INDEX idx_urls_created_at ON urls (created_at);

COMMENT ON TABLE urls IS 'Хранение оригинальных и коротких ссылок';