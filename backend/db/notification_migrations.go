package db

const createNotificationTablesSQL = `
-- Notification Templates Table
CREATE TABLE IF NOT EXISTS notification_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    subject VARCHAR(500),
    body TEXT NOT NULL,
    variables JSONB DEFAULT '[]',
    is_dlt_approved BOOLEAN DEFAULT FALSE,
    dlt_template_id VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_channel ON notification_templates(channel);
CREATE INDEX IF NOT EXISTS idx_notification_templates_provider ON notification_templates(provider);
CREATE INDEX IF NOT EXISTS idx_notification_templates_active ON notification_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_notification_templates_name ON notification_templates(name);

-- Notification Logs Table
CREATE TABLE IF NOT EXISTS notification_logs (
    id BIGSERIAL PRIMARY KEY,
    template_id BIGINT,
    channel VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(500),
    body TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    provider_id VARCHAR(255),
    response TEXT,
    error_msg TEXT,
    retry_count INTEGER DEFAULT 0,
    user_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES notification_templates(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_notification_logs_channel ON notification_logs(channel);
CREATE INDEX IF NOT EXISTS idx_notification_logs_provider ON notification_logs(provider);
CREATE INDEX IF NOT EXISTS idx_notification_logs_status ON notification_logs(status);
CREATE INDEX IF NOT EXISTS idx_notification_logs_recipient ON notification_logs(recipient);
CREATE INDEX IF NOT EXISTS idx_notification_logs_user_id ON notification_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_created_at ON notification_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notification_logs_template_id ON notification_logs(template_id);

-- Notification Configuration Table
CREATE TABLE IF NOT EXISTS notification_config (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    updated_by BIGINT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id),
    UNIQUE(provider, channel, config_key)
);

CREATE INDEX IF NOT EXISTS idx_notification_config_provider ON notification_config(provider);
CREATE INDEX IF NOT EXISTS idx_notification_config_channel ON notification_config(channel);
CREATE INDEX IF NOT EXISTS idx_notification_config_active ON notification_config(is_active);

-- Insert default notification configurations (empty keys for security)
INSERT INTO notification_config (provider, channel, config_key, config_value, is_active, updated_by) VALUES
-- Fast2SMS Configuration
('fast2sms', 'sms', 'api_key', '', FALSE, 1),
('fast2sms', 'sms', 'sender_id', 'FTXSMS', TRUE, 1),
('fast2sms', 'sms', 'base_url', 'https://www.fast2sms.com/dev/bulkV2', TRUE, 1),

-- SMTP Configuration
('smtp', 'email', 'host', 'smtp.gmail.com', TRUE, 1),
('smtp', 'email', 'port', '587', TRUE, 1),
('smtp', 'email', 'username', '', FALSE, 1),
('smtp', 'email', 'password', '', FALSE, 1),
('smtp', 'email', 'from_email', '', FALSE, 1),
('smtp', 'email', 'from_name', 'Fantasy Esports', TRUE, 1),

-- Amazon SES Configuration
('amazon_ses', 'email', 'access_key_id', '', FALSE, 1),
('amazon_ses', 'email', 'secret_access_key', '', FALSE, 1),
('amazon_ses', 'email', 'region', 'us-east-1', TRUE, 1),
('amazon_ses', 'email', 'from_email', '', FALSE, 1),
('amazon_ses', 'email', 'from_name', 'Fantasy Esports', TRUE, 1),

-- Mailchimp Configuration
('mailchimp', 'email', 'api_key', '', FALSE, 1),
('mailchimp', 'email', 'server_prefix', '', FALSE, 1),
('mailchimp', 'email', 'from_email', '', FALSE, 1),
('mailchimp', 'email', 'from_name', 'Fantasy Esports', TRUE, 1),

-- Firebase FCM Configuration
('firebase_fcm', 'push', 'server_key', '', FALSE, 1),
('firebase_fcm', 'push', 'project_id', 'interviewer-ai-2e692', TRUE, 1),
('firebase_fcm', 'push', 'base_url', 'https://fcm.googleapis.com/fcm/send', TRUE, 1),

-- OneSignal Configuration
('onesignal', 'push', 'app_id', '', FALSE, 1),
('onesignal', 'push', 'api_key', '', FALSE, 1),
('onesignal', 'push', 'base_url', 'https://onesignal.com/api/v1/notifications', TRUE, 1),

-- WhatsApp Cloud API Configuration
('whatsapp_cloud', 'whatsapp', 'access_token', '', FALSE, 1),
('whatsapp_cloud', 'whatsapp', 'phone_number_id', '', FALSE, 1),
('whatsapp_cloud', 'whatsapp', 'webhook_verify_token', '', FALSE, 1),
('whatsapp_cloud', 'whatsapp', 'base_url', 'https://graph.facebook.com/v18.0', TRUE, 1)

ON CONFLICT (provider, channel, config_key) DO NOTHING;

-- Insert default notification templates
INSERT INTO notification_templates (name, channel, provider, subject, body, variables, is_dlt_approved, dlt_template_id, created_by) VALUES
-- SMS Templates
('OTP Verification', 'sms', 'fast2sms', NULL, 'Your OTP for Fantasy Esports is {otp}. Valid for 5 minutes. Do not share with anyone.', '["otp"]', TRUE, 'DLT_TEMPLATE_001', 1),
('Welcome SMS', 'sms', 'fast2sms', NULL, 'Welcome to Fantasy Esports, {name}! Start your journey and win big. Download the app now.', '["name"]', TRUE, 'DLT_TEMPLATE_002', 1),
('Contest Reminder', 'sms', 'fast2sms', NULL, 'Hi {name}, join today''s contest for {match}. Only ₹{entry_fee} entry fee. Deadline: {deadline}', '["name", "match", "entry_fee", "deadline"]', TRUE, 'DLT_TEMPLATE_003', 1),
('Wallet Low Balance', 'sms', 'fast2sms', NULL, 'Hi {name}, only ₹{balance} left in your wallet. Add money now to join exciting contests!', '["name", "balance"]', TRUE, 'DLT_TEMPLATE_004', 1),
('KYC Reminder', 'sms', 'fast2sms', NULL, 'Complete your KYC verification to unlock withdrawals and premium features. Verify now!', '[]', TRUE, 'DLT_TEMPLATE_005', 1),

-- Email Templates
('Welcome Email', 'email', 'smtp', 'Welcome to Fantasy Esports!', 'Hi {name},\n\nWelcome to Fantasy Esports! Get ready for an exciting journey.\n\nBest regards,\nFantasy Esports Team', '["name"]', FALSE, NULL, 1),
('KYC Incomplete Reminder', 'email', 'smtp', 'Complete Your KYC Verification', 'Hi {name},\n\nYour KYC verification is still pending. Please complete it to unlock all features.\n\nBest regards,\nFantasy Esports Team', '["name"]', FALSE, NULL, 1),
('Contest Win Notification', 'email', 'smtp', 'Congratulations! You Won ₹{amount}', 'Hi {name},\n\nCongratulations! You won ₹{amount} in the {contest_name} contest.\n\nBest regards,\nFantasy Esports Team', '["name", "amount", "contest_name"]', FALSE, NULL, 1),
('Weekly Newsletter', 'email', 'mailchimp', 'Weekly Fantasy Esports Newsletter', 'Hi {name},\n\nHere are this week''s exciting contests and matches.\n\nBest regards,\nFantasy Esports Team', '["name"]', FALSE, NULL, 1),

-- Push Notification Templates
('Contest Starting Soon', 'push', 'firebase_fcm', 'Contest Alert', '{contest_name} is starting in 30 minutes! Join now before it''s too late.', '["contest_name"]', FALSE, NULL, 1),
('Daily Engagement', 'push', 'firebase_fcm', 'Your Daily Challenge Awaits', 'Check out today''s exciting matches and create your winning team!', '[]', FALSE, NULL, 1),
('Leaderboard Update', 'push', 'firebase_fcm', 'Leaderboard Update', 'You moved to rank #{rank} in {contest_name}! Keep climbing!', '["rank", "contest_name"]', FALSE, NULL, 1),
('Referral Success', 'push', 'firebase_fcm', 'Referral Bonus!', 'Your friend {friend_name} joined using your code. You earned ₹{bonus}!', '["friend_name", "bonus"]', FALSE, NULL, 1),

-- WhatsApp Templates
('OTP WhatsApp', 'whatsapp', 'whatsapp_cloud', NULL, 'Your OTP for Fantasy Esports verification is *{otp}*. Valid for 5 minutes only.', '["otp"]', TRUE, 'WA_TEMPLATE_001', 1),
('Contest Invitation', 'whatsapp', 'whatsapp_cloud', NULL, 'Hi {name}! Join today''s *{match}* contest. Entry: ₹{entry_fee}. Prize Pool: ₹{prize_pool}. Join now!', '["name", "match", "entry_fee", "prize_pool"]', TRUE, 'WA_TEMPLATE_002', 1),
('Personal Engagement', 'whatsapp', 'whatsapp_cloud', NULL, 'Hi {name}, we miss you! Come back and claim your ₹{bonus} welcome back bonus. Valid till {expiry}.', '["name", "bonus", "expiry"]', TRUE, 'WA_TEMPLATE_003', 1)

ON CONFLICT DO NOTHING;

-- Add notification configuration to system_config if needed
INSERT INTO system_config (config_key, config_value, description) VALUES
('notification_settings', '{"retry_limit": 3, "retry_delay_minutes": 5, "batch_size": 100, "rate_limit_per_minute": 60}', 'General notification system settings'),
('cloudinary_settings', '{"cloud_name": "dwnxysjxp", "api_key": "684824545515239", "api_secret": "TaHGxQ0hRDOQW4mYN0hfHGPRxcc"}', 'Cloudinary CDN configuration')
ON CONFLICT (config_key) DO NOTHING;
`