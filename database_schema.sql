-- ============================================================================
-- MONZO CLONE DATABASE SCHEMA
-- ============================================================================
-- Banking application database schema with support for:
-- - User accounts and KYC
-- - Transactions and balance tracking
-- - Cards (physical and virtual)
-- - Savings pots
-- - Payees and payment management
-- - Standing orders and scheduled payments
-- - Direct debits
-- - Notifications and devices
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- CORE USER AND ACCOUNT TABLES
-- ============================================================================

-- USERS TABLE
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    postcode VARCHAR(20),
    country VARCHAR(2) DEFAULT 'GB',
    kyc_status VARCHAR(20) DEFAULT 'pending',
    kyc_verified_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_kyc_status CHECK (kyc_status IN ('pending', 'verified', 'failed', 'review'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_kyc_status ON users(kyc_status);

COMMENT ON TABLE users IS 'Core user accounts with KYC verification';
COMMENT ON COLUMN users.kyc_status IS 'Know Your Customer verification status';

-- ACCOUNTS TABLE
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_number VARCHAR(8) UNIQUE NOT NULL,
    sort_code VARCHAR(6) NOT NULL,
    account_type VARCHAR(20) DEFAULT 'personal',
    balance DECIMAL(15, 2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'GBP',
    is_primary BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_account_type CHECK (account_type IN ('personal', 'business', 'joint')),
    CONSTRAINT chk_status CHECK (status IN ('active', 'frozen', 'closed')),
    CONSTRAINT chk_balance CHECK (balance >= -1000.00)
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_account_number ON accounts(account_number);
CREATE UNIQUE INDEX idx_accounts_primary_per_user ON accounts(user_id, is_primary) WHERE is_primary = true;

COMMENT ON TABLE accounts IS 'Bank accounts belonging to users';
COMMENT ON CONSTRAINT chk_balance ON accounts IS 'Allows £1000 overdraft limit';

-- ACCOUNT_BALANCES TABLE (Balance history and auditing)
CREATE TABLE account_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    balance DECIMAL(15, 2) NOT NULL,
    available_balance DECIMAL(15, 2) NOT NULL,
    pending_balance DECIMAL(15, 2) DEFAULT 0.00,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_account_balances_account_date ON account_balances(account_id, recorded_at DESC);

COMMENT ON TABLE account_balances IS 'Historical balance snapshots for auditing';

-- ============================================================================
-- PAYEES AND PAYMENT RECIPIENTS
-- ============================================================================

-- PAYEES TABLE
CREATE TABLE payees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payee_type VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    account_number VARCHAR(34),
    sort_code VARCHAR(6),
    iban VARCHAR(34),
    swift_code VARCHAR(11),
    email VARCHAR(255),
    phone VARCHAR(20),
    reference VARCHAR(255),
    is_verified BOOLEAN DEFAULT false,
    is_favorite BOOLEAN DEFAULT false,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_payee_type CHECK (payee_type IN
        ('uk_bank', 'international', 'monzo_user', 'business', 'contact')),
    CONSTRAINT chk_uk_bank_details CHECK (
        (payee_type = 'uk_bank' AND account_number IS NOT NULL AND sort_code IS NOT NULL)
        OR payee_type != 'uk_bank'
    ),
    CONSTRAINT chk_international_details CHECK (
        (payee_type = 'international' AND iban IS NOT NULL)
        OR payee_type != 'international'
    )
);

CREATE INDEX idx_payees_user_id ON payees(user_id);
CREATE INDEX idx_payees_is_favorite ON payees(user_id, is_favorite) WHERE is_favorite = true;
CREATE INDEX idx_payees_last_used ON payees(user_id, last_used_at DESC);
CREATE INDEX idx_payees_name ON payees(user_id, name);

COMMENT ON TABLE payees IS 'Saved payment recipients for quick payments';
COMMENT ON COLUMN payees.payee_type IS 'Type of payee: UK bank, international, or internal Monzo user';

-- CONTACTS TABLE (P2P payments within Monzo)
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    account_number VARCHAR(8),
    sort_code VARCHAR(6),
    email VARCHAR(255),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_contacts_user_id ON contacts(user_id);

COMMENT ON TABLE contacts IS 'User contacts for peer-to-peer payments';

-- ============================================================================
-- RECURRING PAYMENTS
-- ============================================================================

-- STANDING_ORDERS TABLE
CREATE TABLE standing_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    payee_id UUID NOT NULL REFERENCES payees(id),
    reference VARCHAR(255) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'GBP',
    frequency VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    next_payment_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    total_payments INTEGER,
    payments_made INTEGER DEFAULT 0,
    last_executed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_standing_order_frequency CHECK (frequency IN
        ('daily', 'weekly', 'fortnightly', 'monthly', 'quarterly', 'annually')),
    CONSTRAINT chk_standing_order_status CHECK (status IN
        ('active', 'paused', 'completed', 'cancelled', 'failed')),
    CONSTRAINT chk_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_dates_valid CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_standing_orders_account_id ON standing_orders(account_id);
CREATE INDEX idx_standing_orders_payee_id ON standing_orders(payee_id);
CREATE INDEX idx_standing_orders_next_payment ON standing_orders(next_payment_date, status)
    WHERE status = 'active';
CREATE INDEX idx_standing_orders_status ON standing_orders(status);

COMMENT ON TABLE standing_orders IS 'Recurring automatic payments (e.g., rent, subscriptions)';
COMMENT ON COLUMN standing_orders.total_payments IS 'NULL means unlimited/ongoing';

-- SCHEDULED_PAYMENTS TABLE
CREATE TABLE scheduled_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    payee_id UUID NOT NULL REFERENCES payees(id),
    reference VARCHAR(255) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'GBP',
    scheduled_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    transaction_id UUID,
    failure_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    executed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_scheduled_payment_status CHECK (status IN
        ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT chk_scheduled_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_future_date CHECK (scheduled_date >= CURRENT_DATE)
);

CREATE INDEX idx_scheduled_payments_account_id ON scheduled_payments(account_id);
CREATE INDEX idx_scheduled_payments_payee_id ON scheduled_payments(payee_id);
CREATE INDEX idx_scheduled_payments_scheduled_date ON scheduled_payments(scheduled_date, status)
    WHERE status IN ('pending', 'processing');
CREATE INDEX idx_scheduled_payments_status ON scheduled_payments(status);

COMMENT ON TABLE scheduled_payments IS 'One-time future-dated payments';

-- DIRECT_DEBITS TABLE
CREATE TABLE direct_debits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    mandate_reference VARCHAR(50) UNIQUE NOT NULL,
    originator_name VARCHAR(255) NOT NULL,
    originator_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_date DATE NOT NULL,
    cancelled_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_direct_debit_status CHECK (status IN
        ('active', 'cancelled', 'suspended', 'expired'))
);

CREATE INDEX idx_direct_debits_account_id ON direct_debits(account_id);
CREATE INDEX idx_direct_debits_mandate_ref ON direct_debits(mandate_reference);
CREATE INDEX idx_direct_debits_status ON direct_debits(status);

COMMENT ON TABLE direct_debits IS 'Direct debit mandates from third parties';

-- PAYMENT_TEMPLATES TABLE
CREATE TABLE payment_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payee_id UUID REFERENCES payees(id),
    template_name VARCHAR(255) NOT NULL,
    amount DECIMAL(15, 2),
    reference VARCHAR(255),
    notes TEXT,
    is_favorite BOOLEAN DEFAULT false,
    use_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_templates_user_id ON payment_templates(user_id);
CREATE INDEX idx_payment_templates_is_favorite ON payment_templates(user_id, is_favorite)
    WHERE is_favorite = true;

COMMENT ON TABLE payment_templates IS 'Saved payment configurations for quick reuse';

-- ============================================================================
-- TRANSACTIONS AND CATEGORIES
-- ============================================================================

-- CATEGORIES TABLE
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    color VARCHAR(7),
    parent_category_id UUID REFERENCES categories(id),
    is_system BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_categories_parent ON categories(parent_category_id);

COMMENT ON TABLE categories IS 'Transaction categories for spending insights';

-- MERCHANTS TABLE
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    logo_url TEXT,
    default_category_id UUID REFERENCES categories(id),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_merchants_name ON merchants(name);
CREATE INDEX idx_merchants_metadata ON merchants USING gin(metadata);

COMMENT ON TABLE merchants IS 'Merchant information with branding and categorization';

-- TRANSACTIONS TABLE (Partitioned by date)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    merchant_id UUID REFERENCES merchants(id),
    category_id UUID REFERENCES categories(id),
    contact_id UUID REFERENCES contacts(id),
    payee_id UUID REFERENCES payees(id),
    standing_order_id UUID REFERENCES standing_orders(id),
    scheduled_payment_id UUID REFERENCES scheduled_payments(id),
    direct_debit_id UUID REFERENCES direct_debits(id),
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'GBP',
    description TEXT,
    notes TEXT,
    running_balance DECIMAL(15, 2),
    status VARCHAR(20) DEFAULT 'completed',
    declined_reason TEXT,
    metadata JSONB,
    transaction_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    settled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_transaction_type CHECK (transaction_type IN
        ('debit', 'credit', 'transfer_out', 'transfer_in', 'payment', 'refund', 'fee', 'interest')),
    CONSTRAINT chk_transaction_status CHECK (status IN ('pending', 'completed', 'declined', 'reversed'))
) PARTITION BY RANGE (transaction_date);

-- Create partitions for 2026
CREATE TABLE transactions_2026_q1 PARTITION OF transactions
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');
CREATE TABLE transactions_2026_q2 PARTITION OF transactions
    FOR VALUES FROM ('2026-04-01') TO ('2026-07-01');
CREATE TABLE transactions_2026_q3 PARTITION OF transactions
    FOR VALUES FROM ('2026-07-01') TO ('2026-10-01');
CREATE TABLE transactions_2026_q4 PARTITION OF transactions
    FOR VALUES FROM ('2026-10-01') TO ('2027-01-01');

-- Create partitions for 2027 (forward planning)
CREATE TABLE transactions_2027_q1 PARTITION OF transactions
    FOR VALUES FROM ('2027-01-01') TO ('2027-04-01');

CREATE INDEX idx_transactions_account_date ON transactions(account_id, transaction_date DESC);
CREATE INDEX idx_transactions_merchant ON transactions(merchant_id);
CREATE INDEX idx_transactions_category ON transactions(category_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_metadata ON transactions USING gin(metadata);
CREATE INDEX idx_transactions_standing_order ON transactions(standing_order_id);
CREATE INDEX idx_transactions_scheduled_payment ON transactions(scheduled_payment_id);
CREATE INDEX idx_transactions_direct_debit ON transactions(direct_debit_id);
CREATE INDEX idx_transactions_payee ON transactions(payee_id);

COMMENT ON TABLE transactions IS 'All financial transactions with partitioning for performance';

-- Add foreign key constraint for scheduled_payments after transactions table exists
ALTER TABLE scheduled_payments
ADD CONSTRAINT fk_scheduled_payments_transaction
FOREIGN KEY (transaction_id) REFERENCES transactions(id);

-- ============================================================================
-- CARDS AND POTS
-- ============================================================================

-- CARDS TABLE
CREATE TABLE cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    card_number_encrypted BYTEA NOT NULL,
    card_type VARCHAR(20) DEFAULT 'debit',
    last_four VARCHAR(4) NOT NULL,
    expiry_date DATE NOT NULL,
    cvv_encrypted BYTEA NOT NULL,
    is_frozen BOOLEAN DEFAULT false,
    is_virtual BOOLEAN DEFAULT false,
    spending_limit DECIMAL(15, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_card_type CHECK (card_type IN ('debit', 'credit', 'virtual'))
);

CREATE INDEX idx_cards_account_id ON cards(account_id);
CREATE INDEX idx_cards_last_four ON cards(last_four);

COMMENT ON TABLE cards IS 'Physical and virtual cards with encrypted sensitive data';
COMMENT ON COLUMN cards.card_number_encrypted IS 'PCI-DSS compliant encrypted storage';

-- POTS TABLE (Savings goals)
CREATE TABLE pots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    goal_type VARCHAR(50),
    target_amount DECIMAL(15, 2),
    current_amount DECIMAL(15, 2) DEFAULT 0.00,
    color VARCHAR(7) DEFAULT '#00A4DB',
    icon VARCHAR(50) DEFAULT 'piggy-bank',
    auto_deposit BOOLEAN DEFAULT false,
    auto_deposit_amount DECIMAL(15, 2),
    auto_deposit_frequency VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_auto_deposit_frequency CHECK (auto_deposit_frequency IN
        ('daily', 'weekly', 'monthly', 'payday'))
);

CREATE INDEX idx_pots_account_id ON pots(account_id);

COMMENT ON TABLE pots IS 'Savings pots for goal-based money management';

-- POT_TRANSACTIONS TABLE
CREATE TABLE pot_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pot_id UUID NOT NULL REFERENCES pots(id) ON DELETE CASCADE,
    source_account_id UUID NOT NULL REFERENCES accounts(id),
    amount DECIMAL(15, 2) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_pot_transaction_type CHECK (transaction_type IN ('deposit', 'withdrawal'))
);

CREATE INDEX idx_pot_transactions_pot_id ON pot_transactions(pot_id);
CREATE INDEX idx_pot_transactions_created_at ON pot_transactions(created_at DESC);

COMMENT ON TABLE pot_transactions IS 'Movements in and out of savings pots';

-- ============================================================================
-- NOTIFICATIONS AND DEVICES
-- ============================================================================

-- NOTIFICATIONS TABLE
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES transactions(id),
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,

    CONSTRAINT chk_notification_type CHECK (notification_type IN
        ('transaction', 'security', 'marketing', 'system', 'payment_request'))
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_sent_at ON notifications(sent_at DESC);

COMMENT ON TABLE notifications IS 'In-app and push notifications';

-- DEVICES TABLE
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) UNIQUE NOT NULL,
    device_type VARCHAR(20) NOT NULL,
    push_token TEXT,
    is_active BOOLEAN DEFAULT true,
    last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_device_type CHECK (device_type IN ('ios', 'android', 'web'))
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_device_id ON devices(device_id);

COMMENT ON TABLE devices IS 'User devices for push notifications and security';

-- ============================================================================
-- TRIGGERS AND AUTOMATION
-- ============================================================================

-- Trigger: Update standing order next payment date after execution
CREATE OR REPLACE FUNCTION update_standing_order_next_payment()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.payments_made > OLD.payments_made THEN
        UPDATE standing_orders
        SET next_payment_date = CASE NEW.frequency
            WHEN 'daily' THEN NEW.next_payment_date + INTERVAL '1 day'
            WHEN 'weekly' THEN NEW.next_payment_date + INTERVAL '1 week'
            WHEN 'fortnightly' THEN NEW.next_payment_date + INTERVAL '2 weeks'
            WHEN 'monthly' THEN NEW.next_payment_date + INTERVAL '1 month'
            WHEN 'quarterly' THEN NEW.next_payment_date + INTERVAL '3 months'
            WHEN 'annually' THEN NEW.next_payment_date + INTERVAL '1 year'
        END,
        last_executed_at = CURRENT_TIMESTAMP,
        status = CASE
            WHEN NEW.total_payments IS NOT NULL
                 AND NEW.payments_made >= NEW.total_payments
            THEN 'completed'
            ELSE NEW.status
        END
        WHERE id = NEW.id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_standing_order_executed
AFTER UPDATE OF payments_made ON standing_orders
FOR EACH ROW
EXECUTE FUNCTION update_standing_order_next_payment();

-- Trigger: Update payee last_used_at when used in transaction
CREATE OR REPLACE FUNCTION update_payee_last_used()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.payee_id IS NOT NULL THEN
        UPDATE payees
        SET last_used_at = CURRENT_TIMESTAMP
        WHERE id = NEW.payee_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_payee_used_in_transaction
AFTER INSERT ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_payee_last_used();

-- Trigger: Update account updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_accounts_updated_at
BEFORE UPDATE ON accounts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_cards_updated_at
BEFORE UPDATE ON cards
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_pots_updated_at
BEFORE UPDATE ON pots
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_payees_updated_at
BEFORE UPDATE ON payees
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_standing_orders_updated_at
BEFORE UPDATE ON standing_orders
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_scheduled_payments_updated_at
BEFORE UPDATE ON scheduled_payments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SEED DATA: DEFAULT CATEGORIES
-- ============================================================================

INSERT INTO categories (name, icon, color, is_system) VALUES
('General', '📦', '#95A6A6', true),
('Groceries', '🛒', '#EE7B30', true),
('Eating Out', '🍽️', '#EA4C89', true),
('Transport', '🚗', '#52CC7A', true),
('Shopping', '🛍️', '#D251B4', true),
('Entertainment', '🎬', '#FF6600', true),
('Bills', '📄', '#52A5FF', true),
('Cash', '💵', '#626262', true),
('Family', '👨‍👩‍👧‍👦', '#52CC7A', true),
('Finances', '💰', '#626262', true),
('Gifts', '🎁', '#EA4C89', true),
('Health', '⚕️', '#52A5FF', true),
('Holidays', '✈️', '#00D4CC', true),
('Personal Care', '💇', '#D251B4', true),
('Savings', '🐷', '#52CC7A', true),
('Subscriptions', '📱', '#FF6600', true),
('Transfers', '↔️', '#95A6A6', true);

-- ============================================================================
-- USEFUL QUERIES
-- ============================================================================

-- Get account balance with pending transactions
-- SELECT
--     a.id,
--     a.balance,
--     COALESCE(SUM(CASE WHEN t.status = 'pending' THEN t.amount ELSE 0 END), 0) as pending_amount,
--     a.balance + COALESCE(SUM(CASE WHEN t.status = 'pending' THEN t.amount ELSE 0 END), 0) as available_balance
-- FROM accounts a
-- LEFT JOIN transactions t ON a.id = t.account_id
-- WHERE a.id = 'account-uuid'
-- GROUP BY a.id;

-- Get upcoming standing orders
-- SELECT so.*, p.name as payee_name
-- FROM standing_orders so
-- JOIN payees p ON so.payee_id = p.id
-- WHERE so.status = 'active'
--   AND so.next_payment_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '7 days'
-- ORDER BY so.next_payment_date;

-- Get spending by category for current month
-- SELECT
--     c.name,
--     c.icon,
--     c.color,
--     COUNT(t.id) as transaction_count,
--     ABS(SUM(t.amount)) as total_spent
-- FROM transactions t
-- JOIN categories c ON t.category_id = c.id
-- WHERE t.account_id = 'account-uuid'
--   AND t.transaction_type = 'debit'
--   AND t.transaction_date >= date_trunc('month', CURRENT_DATE)
-- GROUP BY c.id, c.name, c.icon, c.color
-- ORDER BY total_spent DESC;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
