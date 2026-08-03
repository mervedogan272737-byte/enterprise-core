CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    position VARCHAR(150),
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT contacts_status_check CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_contacts_organization_id
ON contacts(organization_id);

CREATE INDEX IF NOT EXISTS idx_contacts_customer_id
ON contacts(customer_id);

CREATE INDEX IF NOT EXISTS idx_contacts_org_customer
ON contacts(organization_id, customer_id);
