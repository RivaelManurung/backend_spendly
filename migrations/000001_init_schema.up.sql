CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- USERS TABLE
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(200),
    avatar_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','pending_verification','deactivated')),
    currency_preference VARCHAR(10) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted ON users(id) WHERE deleted_at IS NULL;

-- AUTH PROVIDERS TABLE
CREATE TABLE auth_providers (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_auth_provider_user UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_auth_user_id ON auth_providers(user_id);

-- SYSTEM CATEGORIES TABLE
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('income','expense')),
    icon VARCHAR(50),
    color_hex VARCHAR(7),
    ai_tags TEXT[],
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_type ON categories(type);
CREATE INDEX idx_categories_is_system ON categories(is_system);

-- USER CATEGORIES TABLE
CREATE TABLE user_categories (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    custom_name VARCHAR(100),
    custom_icon VARCHAR(50),
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_category UNIQUE (user_id, category_id)
);

CREATE INDEX idx_user_categories_user ON user_categories(user_id);

-- TRANSACTIONS TABLE
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    amount NUMERIC(15,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    amount_in_base NUMERIC(15,2) NOT NULL,
    description TEXT,
    merchant VARCHAR(255),
    source VARCHAR(30) NOT NULL DEFAULT 'manual' CHECK (source IN ('manual','import','ocr','bank_sync')),
    transaction_date DATE NOT NULL,
    ai_category_suggestion VARCHAR(100),
    ai_confidence_score NUMERIC(4,3),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_txn_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX idx_txn_user_category ON transactions(user_id, category_id);
CREATE INDEX idx_txn_merchant ON transactions(user_id, merchant);
CREATE INDEX idx_txn_source ON transactions(user_id, source);
CREATE INDEX idx_txn_active ON transactions(user_id, transaction_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_txn_metadata ON transactions USING GIN(metadata);

-- TRANSACTION TAGS TABLE
CREATE TABLE transaction_tags (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    CONSTRAINT uq_txn_tag UNIQUE (transaction_id, tag)
);

CREATE INDEX idx_tags_transaction ON transaction_tags(transaction_id);
CREATE INDEX idx_tags_user_tag ON transaction_tags(tag);

-- BUDGETS TABLE
CREATE TABLE budgets (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    limit_amount NUMERIC(15,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    period_type VARCHAR(10) NOT NULL CHECK (period_type IN ('weekly','monthly','yearly')),
    period_start DATE,
    period_end DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_user ON budgets(user_id);
CREATE INDEX idx_budgets_active ON budgets(user_id) WHERE is_active = TRUE;

-- ANALYSIS SNAPSHOTS TABLE
CREATE TABLE analysis_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_type VARCHAR(10) NOT NULL CHECK (period_type IN ('daily','weekly','monthly','yearly')),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_income NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_expense NUMERIC(15,2) NOT NULL DEFAULT 0,
    net_cashflow NUMERIC(15,2) NOT NULL,
    top_expense_category VARCHAR(100),
    category_breakdown JSONB,
    merchant_breakdown JSONB,
    daily_trend JSONB,
    transaction_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_snapshot_user_period UNIQUE (user_id, period_type, period_start)
);

CREATE INDEX idx_snapshots_user_period ON analysis_snapshots(user_id, period_type, period_start DESC);

-- AI INSIGHTS TABLE
CREATE TABLE ai_insights (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    snapshot_id BIGINT REFERENCES analysis_snapshots(id) ON DELETE SET NULL,
    insight_type VARCHAR(50) NOT NULL CHECK (insight_type IN ('spending_alert','saving_opportunity','anomaly','monthly_summary','budget_warning','positive_trend')),
    insight TEXT NOT NULL,
    warning TEXT,
    recommendation TEXT,
    supporting_data JSONB,
    model VARCHAR(100) NOT NULL,
    confidence_score NUMERIC(4,3),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    is_helpful BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_insights_user_created ON ai_insights(user_id, created_at DESC);
CREATE INDEX idx_insights_unread ON ai_insights(user_id) WHERE is_read = FALSE;
CREATE INDEX idx_insights_type ON ai_insights(user_id, insight_type);
