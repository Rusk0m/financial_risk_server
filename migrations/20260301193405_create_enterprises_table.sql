-- +goose Up
CREATE TABLE enterprises (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    industry VARCHAR(100),
    annual_production_t NUMERIC(15, 2) NOT NULL,
    export_share_percent NUMERIC(5, 2) NOT NULL,
    main_currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_enterprises_name ON enterprises(name);
CREATE INDEX idx_enterprises_industry ON enterprises(industry);

-- +goose Down
DROP TABLE IF EXISTS enterprises CASCADE;
