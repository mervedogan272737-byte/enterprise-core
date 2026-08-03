-- ============================================================
-- 003 CUSTOMERS / CRM
-- ============================================================

CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    company TEXT,
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT customers_status_check
        CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_customers_organization_id
    ON customers(organization_id);

CREATE INDEX IF NOT EXISTS idx_customers_email
    ON customers(email);

CREATE INDEX IF NOT EXISTS idx_customers_company
    ON customers(company);

CREATE INDEX IF NOT EXISTS idx_customers_status
    ON customers(status);
