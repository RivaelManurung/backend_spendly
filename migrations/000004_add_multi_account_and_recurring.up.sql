-- =============================================
-- ACCOUNTS TABLE (Multiple Account Types)
-- Inspired by Money Manager Expense & Budget
-- =============================================
CREATE TABLE accounts (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                VARCHAR(100) NOT NULL,
    type                VARCHAR(20) NOT NULL CHECK (type IN ('cash','debit','credit','savings','investment','loan','e_wallet')),
    initial_balance     NUMERIC(15,2) NOT NULL DEFAULT 0,
    current_balance     NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency            VARCHAR(10) NOT NULL DEFAULT 'IDR',
    color               VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    icon                VARCHAR(50) NOT NULL DEFAULT 'wallet',
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    exclude_from_total  BOOLEAN NOT NULL DEFAULT FALSE,
    credit_limit        NUMERIC(15,2),
    payment_due_day     SMALLINT CHECK (payment_due_day >= 1 AND payment_due_day <= 31),
    notes               TEXT,
    sort_order          INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_accounts_user ON accounts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_type ON accounts(user_id, type) WHERE deleted_at IS NULL;

-- Default accounts trigger: add Cash account on new user (optional, handled in service)
COMMENT ON TABLE accounts IS 'User financial accounts (cash, bank, investment, etc.) for double-entry tracking';

-- =============================================
-- ACCOUNT_TRANSFERS TABLE (Internal Transfer)
-- Transfer between accounts is NOT an expense/income
-- =============================================
CREATE TABLE account_transfers (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    to_account_id   BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount          NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    fee             NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10) NOT NULL DEFAULT 'IDR',
    notes           TEXT,
    transfer_date   DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_different_accounts CHECK (from_account_id != to_account_id)
);

CREATE INDEX idx_transfers_user ON account_transfers(user_id, transfer_date DESC);
CREATE INDEX idx_transfers_from_acc ON account_transfers(from_account_id);
CREATE INDEX idx_transfers_to_acc ON account_transfers(to_account_id);

-- =============================================
-- RECURRING TRANSACTIONS TABLE
-- Template/Bookmark for repeating transactions
-- =============================================
CREATE TABLE recurring_transactions (
    id            BIGSERIAL PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id    BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    category_id   BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    title         VARCHAR(150) NOT NULL,
    amount        NUMERIC(15,2) NOT NULL,
    currency      VARCHAR(10) NOT NULL DEFAULT 'IDR',
    type          VARCHAR(10) NOT NULL CHECK (type IN ('income','expense')),
    frequency     VARCHAR(10) NOT NULL CHECK (frequency IN ('daily','weekly','monthly','yearly')),
    start_date    DATE NOT NULL,
    end_date      DATE,
    next_due_date DATE NOT NULL,
    last_run_at   TIMESTAMPTZ,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    auto_post     BOOLEAN NOT NULL DEFAULT FALSE,
    notes         TEXT,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurring_user ON recurring_transactions(user_id) WHERE is_active = TRUE;
CREATE INDEX idx_recurring_due ON recurring_transactions(next_due_date) WHERE is_active = TRUE;

-- =============================================
-- NET WORTH SNAPSHOTS TABLE (Asset Graph)
-- Daily/Monthly snapshot for trend analysis
-- =============================================
CREATE TABLE net_worth_snapshots (
    id            BIGSERIAL PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_assets  NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_debt    NUMERIC(15,2) NOT NULL DEFAULT 0,
    net_worth     NUMERIC(15,2) NOT NULL DEFAULT 0,
    accounts_json JSONB,  -- Snapshot of each account balance at that time
    recorded_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period_label  VARCHAR(20)  -- e.g. "2025-04" for monthly grouping
);

CREATE INDEX idx_networth_user_date ON net_worth_snapshots(user_id, recorded_at DESC);
CREATE UNIQUE INDEX idx_networth_user_period ON net_worth_snapshots(user_id, period_label) WHERE period_label IS NOT NULL;

-- =============================================
-- Update TRANSACTIONS TABLE: add account_id, receipt_url
-- =============================================
ALTER TABLE transactions ADD COLUMN account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;
ALTER TABLE transactions ADD COLUMN receipt_url TEXT;
ALTER TABLE transactions ADD COLUMN is_transfer BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE transactions ADD COLUMN transfer_id BIGINT REFERENCES account_transfers(id) ON DELETE SET NULL;

CREATE INDEX idx_txn_account ON transactions(account_id) WHERE account_id IS NOT NULL AND deleted_at IS NULL;

-- =============================================
-- Update BUDGETS TABLE: add period_type (weekly/monthly/yearly)
-- =============================================
ALTER TABLE budgets ADD COLUMN period_start DATE;
ALTER TABLE budgets ADD COLUMN period_end DATE;
ALTER TABLE budgets RENAME COLUMN period TO period_label;

COMMENT ON TABLE budgets IS 'Per-category budget planning with weekly/monthly/yearly granularity';
