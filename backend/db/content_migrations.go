package db

const createContentManagementTables = `
-- Banners and Promotions Management
CREATE TABLE IF NOT EXISTS banners (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description VARCHAR(500),
    image_url VARCHAR(500) NOT NULL,
    link_url VARCHAR(500),
    position VARCHAR(20) NOT NULL CHECK (position IN ('top', 'middle', 'bottom', 'sidebar')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('promotion', 'announcement', 'sponsored')),
    priority INTEGER NOT NULL DEFAULT 0 CHECK (priority >= 0 AND priority <= 100),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    target_roles JSONB,
    metadata JSONB,
    click_count BIGINT DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id),
    CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_banners_position ON banners(position);
CREATE INDEX IF NOT EXISTS idx_banners_type ON banners(type);
CREATE INDEX IF NOT EXISTS idx_banners_active_dates ON banners(is_active, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_banners_priority ON banners(priority DESC);
CREATE INDEX IF NOT EXISTS idx_banners_created_by ON banners(created_by);

-- Email Templates for Marketing
CREATE TABLE IF NOT EXISTS email_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description VARCHAR(500),
    subject VARCHAR(200) NOT NULL,
    html_content TEXT NOT NULL,
    text_content TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('welcome', 'promotional', 'transactional', 'newsletter')),
    variables JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_email_templates_category ON email_templates(category);
CREATE INDEX IF NOT EXISTS idx_email_templates_active ON email_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_email_templates_created_by ON email_templates(created_by);

-- Marketing Campaigns
CREATE TABLE IF NOT EXISTS marketing_campaigns (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    subject VARCHAR(200) NOT NULL,
    email_template TEXT NOT NULL,
    target_segment VARCHAR(100) NOT NULL,
    target_criteria JSONB,
    scheduled_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'cancelled')),
    total_recipients INTEGER DEFAULT 0,
    sent_count INTEGER DEFAULT 0,
    delivered_count INTEGER DEFAULT 0,
    open_count INTEGER DEFAULT 0,
    click_count INTEGER DEFAULT 0,
    unsubscribe_count INTEGER DEFAULT 0,
    bounce_count INTEGER DEFAULT 0,
    metadata JSONB,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_marketing_campaigns_status ON marketing_campaigns(status);
CREATE INDEX IF NOT EXISTS idx_marketing_campaigns_scheduled ON marketing_campaigns(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_marketing_campaigns_segment ON marketing_campaigns(target_segment);
CREATE INDEX IF NOT EXISTS idx_marketing_campaigns_created_by ON marketing_campaigns(created_by);

-- SEO Content Management
CREATE TABLE IF NOT EXISTS seo_content (
    id BIGSERIAL PRIMARY KEY,
    page_type VARCHAR(100) NOT NULL,
    page_slug VARCHAR(200) NOT NULL UNIQUE,
    meta_title VARCHAR(60) NOT NULL,
    meta_description VARCHAR(160) NOT NULL,
    keywords JSONB,
    og_title VARCHAR(60),
    og_description VARCHAR(160),
    og_image VARCHAR(500),
    twitter_card VARCHAR(50),
    structured_data JSONB,
    content TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_seo_content_page_type ON seo_content(page_type);
CREATE INDEX IF NOT EXISTS idx_seo_content_slug ON seo_content(page_slug);
CREATE INDEX IF NOT EXISTS idx_seo_content_active ON seo_content(is_active);
CREATE INDEX IF NOT EXISTS idx_seo_content_created_by ON seo_content(created_by);

-- FAQ Sections
CREATE TABLE IF NOT EXISTS faq_sections (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description VARCHAR(500),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_faq_sections_sort_order ON faq_sections(sort_order);
CREATE INDEX IF NOT EXISTS idx_faq_sections_active ON faq_sections(is_active);
CREATE INDEX IF NOT EXISTS idx_faq_sections_created_by ON faq_sections(created_by);

-- FAQ Items
CREATE TABLE IF NOT EXISTS faq_items (
    id BIGSERIAL PRIMARY KEY,
    section_id BIGINT NOT NULL,
    question VARCHAR(500) NOT NULL,
    answer TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    view_count BIGINT DEFAULT 0,
    like_count BIGINT DEFAULT 0,
    tags JSONB,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (section_id) REFERENCES faq_sections(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_faq_items_section ON faq_items(section_id);
CREATE INDEX IF NOT EXISTS idx_faq_items_sort_order ON faq_items(section_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_faq_items_active ON faq_items(is_active);
CREATE INDEX IF NOT EXISTS idx_faq_items_view_count ON faq_items(view_count DESC);
CREATE INDEX IF NOT EXISTS idx_faq_items_created_by ON faq_items(created_by);

-- Legal Documents
CREATE TABLE IF NOT EXISTS legal_documents (
    id BIGSERIAL PRIMARY KEY,
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN ('terms', 'privacy', 'refund', 'cookie', 'disclaimer')),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    version VARCHAR(20) NOT NULL,
    effective_date TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    is_active BOOLEAN DEFAULT FALSE,
    metadata JSONB,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id),
    UNIQUE(document_type, version)
);

CREATE INDEX IF NOT EXISTS idx_legal_documents_type ON legal_documents(document_type);
CREATE INDEX IF NOT EXISTS idx_legal_documents_status ON legal_documents(status);
CREATE INDEX IF NOT EXISTS idx_legal_documents_active ON legal_documents(is_active);
CREATE INDEX IF NOT EXISTS idx_legal_documents_effective ON legal_documents(effective_date);
CREATE INDEX IF NOT EXISTS idx_legal_documents_created_by ON legal_documents(created_by);

-- Content Analytics
CREATE TABLE IF NOT EXISTS content_analytics (
    id BIGSERIAL PRIMARY KEY,
    content_type VARCHAR(50) NOT NULL CHECK (content_type IN ('banner', 'campaign', 'seo', 'faq', 'legal')),
    content_id BIGINT NOT NULL,
    date DATE NOT NULL,
    views BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    engagements BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    metrics JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(content_type, content_id, date)
);

CREATE INDEX IF NOT EXISTS idx_content_analytics_type_id ON content_analytics(content_type, content_id);
CREATE INDEX IF NOT EXISTS idx_content_analytics_date ON content_analytics(date DESC);
CREATE INDEX IF NOT EXISTS idx_content_analytics_views ON content_analytics(views DESC);

-- Content Scheduler (for automated content publishing)
CREATE TABLE IF NOT EXISTS content_scheduler (
    id BIGSERIAL PRIMARY KEY,
    content_type VARCHAR(50) NOT NULL,
    content_id BIGINT NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('publish', 'unpublish', 'archive')),
    scheduled_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'executed', 'failed', 'cancelled')),
    executed_at TIMESTAMP,
    error_message TEXT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_content_scheduler_scheduled ON content_scheduler(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_content_scheduler_status ON content_scheduler(status);
CREATE INDEX IF NOT EXISTS idx_content_scheduler_type_id ON content_scheduler(content_type, content_id);

-- Insert sample data for testing
INSERT INTO email_templates (name, description, subject, html_content, text_content, category, variables, created_by) VALUES
('Welcome Email', 'Welcome message for new users', 'Welcome to Fantasy Esports!', 
'<h1>Welcome {{name}}!</h1><p>Thanks for joining our platform.</p>', 
'Welcome {{name}}! Thanks for joining our platform.', 
'welcome', '{"name": "User Name", "email": "User Email"}', 1),
('Deposit Bonus', 'Promotional email for deposit bonus', 'Get 100% Bonus on Your First Deposit!',
'<h1>Special Offer for You!</h1><p>Get 100% bonus on your first deposit. Limited time offer!</p>',
'Special Offer for You! Get 100% bonus on your first deposit. Limited time offer!',
'promotional', '{"name": "User Name", "bonus_amount": "Bonus Amount"}', 1)
ON CONFLICT DO NOTHING;

-- Insert sample SEO content
INSERT INTO seo_content (page_type, page_slug, meta_title, meta_description, keywords, og_title, og_description, content, created_by) VALUES
('home', 'home', 'Fantasy Esports - Create Teams, Win Big!', 
'Join the ultimate fantasy esports platform. Create teams, compete in contests, and win real money!',
'["fantasy", "esports", "gaming", "contests", "valorant", "bgmi"]',
'Fantasy Esports - Create Teams, Win Big!',
'Join the ultimate fantasy esports platform. Create teams, compete in contests, and win real money!',
'<h1>Welcome to Fantasy Esports</h1><p>The ultimate destination for esports fantasy gaming.</p>',
1),
('games', 'games', 'Games - Fantasy Esports', 
'Explore all available games on our fantasy esports platform including Valorant, BGMI, CS2 and more.',
'["games", "valorant", "bgmi", "cs2", "esports"]',
'Games - Fantasy Esports',
'Explore all available games on our fantasy esports platform including Valorant, BGMI, CS2 and more.',
'<h1>Available Games</h1><p>Choose from our wide selection of esports games.</p>',
1)
ON CONFLICT (page_slug) DO NOTHING;

-- Insert sample FAQ sections
INSERT INTO faq_sections (name, description, sort_order, created_by) VALUES
('Getting Started', 'Basic questions for new users', 1, 1),
('Payments & Wallet', 'Questions about payments and wallet management', 2, 1),
('Contests & Teams', 'Questions about contests and team creation', 3, 1),
('Account & KYC', 'Questions about account and KYC verification', 4, 1)
ON CONFLICT DO NOTHING;

-- Insert sample FAQ items  
INSERT INTO faq_items (section_id, question, answer, sort_order, tags, created_by) VALUES
(1, 'How do I create my first fantasy team?', 'To create your first fantasy team, go to any upcoming match, click "Create Team", select 5 players within the 100 credit budget, choose your captain (2x points) and vice-captain (1.5x points), then save your team.', 1, '["team", "creation", "getting-started"]', 1),
(1, 'What is the minimum deposit amount?', 'The minimum deposit amount is ₹100. You can add money using various payment methods including UPI, cards, and net banking.', 2, '["deposit", "payment", "minimum"]', 1),
(2, 'How do I withdraw my winnings?', 'To withdraw winnings, go to your wallet, click "Withdraw", enter the amount (minimum ₹200), provide your bank account details, and submit. Withdrawals are processed within 1-2 business days.', 1, '["withdraw", "winnings", "bank"]', 1),
(2, 'What are the different wallet balances?', 'You have three wallet balances: Bonus Balance (from promotions), Deposit Balance (money you added), and Winning Balance (contest winnings). You can withdraw Deposit + Winning balances.', 2, '["wallet", "balance", "types"]', 1),
(3, 'How does the scoring system work?', 'Players earn points based on their real match performance. For example, in Valorant: Kill = 2 points, Death = -1 point, Assist = 1 point, Ace = 10 points. Captain gets 2x points, Vice-captain gets 1.5x points.', 1, '["scoring", "points", "captain"]', 1),
(4, 'What documents are required for KYC?', 'For KYC verification, you need: PAN Card, Aadhaar Card, and Bank Account Statement/Passbook. All documents should be clear and valid.', 1, '["kyc", "documents", "verification"]', 1)
ON CONFLICT DO NOTHING;

-- Insert sample legal documents
INSERT INTO legal_documents (document_type, title, content, version, effective_date, status, is_active, created_by) VALUES
('terms', 'Terms and Conditions', 'These terms and conditions govern your use of our fantasy esports platform...', '1.0', '2025-01-01 00:00:00', 'published', true, 1),
('privacy', 'Privacy Policy', 'This privacy policy explains how we collect, use, and protect your personal information...', '1.0', '2025-01-01 00:00:00', 'published', true, 1),
('refund', 'Refund Policy', 'Our refund policy outlines the conditions under which refunds may be processed...', '1.0', '2025-01-01 00:00:00', 'published', true, 1)
ON CONFLICT DO NOTHING;
`