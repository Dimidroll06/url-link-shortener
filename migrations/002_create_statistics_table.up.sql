CREATE TABLE
    IF NOT EXISTS url_stats (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        url_id UUID NOT NULL REFERENCES urls (id) ON DELETE CASCADE,
        accessed_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            ip_address INET,
            user_agent TEXT,
            referer TEXT
    );

CREATE INDEX idx_stats_url_id ON url_stats (url_id);

CREATE INDEX idx_stats_accessed_at ON url_stats (accessed_at);

COMMENT ON TABLE url_stats IS 'Статистика переходов по ссылкам';