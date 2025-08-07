package db

const createAdvancedFeaturesTables = `
-- Achievement System
CREATE TABLE IF NOT EXISTS achievements (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    badge_icon VARCHAR(500),
    badge_color VARCHAR(20) DEFAULT '#FFD700',
    category VARCHAR(50) NOT NULL CHECK (category IN ('gameplay', 'social', 'progression', 'special')),
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('first_team', 'contest_win', 'referral_milestone', 'winning_streak', 'custom')),
    trigger_criteria JSONB NOT NULL,
    reward_type VARCHAR(50) CHECK (reward_type IN ('badge', 'bonus', 'title')),
    reward_value DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_hidden BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_achievements_category ON achievements(category);
CREATE INDEX IF NOT EXISTS idx_achievements_trigger_type ON achievements(trigger_type);
CREATE INDEX IF NOT EXISTS idx_achievements_active ON achievements(is_active);

-- User Achievements
CREATE TABLE IF NOT EXISTS user_achievements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    achievement_id BIGINT NOT NULL,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    progress_data JSONB,
    is_featured BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (achievement_id) REFERENCES achievements(id) ON DELETE CASCADE,
    UNIQUE(user_id, achievement_id)
);

CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_earned ON user_achievements(earned_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_achievements_featured ON user_achievements(is_featured);

-- Friends System
CREATE TABLE IF NOT EXISTS user_friends (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    friend_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'blocked')),
    requested_by BIGINT NOT NULL,
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (requested_by) REFERENCES users(id),
    UNIQUE(user_id, friend_id),
    CHECK (user_id != friend_id)
);

CREATE INDEX IF NOT EXISTS idx_user_friends_user ON user_friends(user_id);
CREATE INDEX IF NOT EXISTS idx_user_friends_status ON user_friends(status);
CREATE INDEX IF NOT EXISTS idx_user_friends_requested ON user_friends(requested_at DESC);

-- Friend Challenges
CREATE TABLE IF NOT EXISTS friend_challenges (
    id BIGSERIAL PRIMARY KEY,
    challenger_id BIGINT NOT NULL,
    challenged_id BIGINT NOT NULL,
    match_id BIGINT NOT NULL,
    challenge_type VARCHAR(50) DEFAULT 'head_to_head' CHECK (challenge_type IN ('head_to_head', 'score_battle', 'team_vs_team')),
    entry_fee DECIMAL(10,2) DEFAULT 0,
    prize_amount DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'completed', 'cancelled')),
    winner_id BIGINT,
    challenger_team_id BIGINT,
    challenged_team_id BIGINT,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (challenger_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (challenged_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id),
    FOREIGN KEY (winner_id) REFERENCES users(id),
    FOREIGN KEY (challenger_team_id) REFERENCES user_teams(id),
    FOREIGN KEY (challenged_team_id) REFERENCES user_teams(id)
);

CREATE INDEX IF NOT EXISTS idx_friend_challenges_challenger ON friend_challenges(challenger_id);
CREATE INDEX IF NOT EXISTS idx_friend_challenges_challenged ON friend_challenges(challenged_id);
CREATE INDEX IF NOT EXISTS idx_friend_challenges_status ON friend_challenges(status);
CREATE INDEX IF NOT EXISTS idx_friend_challenges_match ON friend_challenges(match_id);

-- Friend Activity Feed
CREATE TABLE IF NOT EXISTS friend_activities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    activity_type VARCHAR(50) NOT NULL CHECK (activity_type IN ('team_created', 'contest_joined', 'contest_won', 'achievement_earned', 'friend_added')),
    activity_data JSONB NOT NULL,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_friend_activities_user ON friend_activities(user_id);
CREATE INDEX IF NOT EXISTS idx_friend_activities_type ON friend_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_friend_activities_created ON friend_activities(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_friend_activities_public ON friend_activities(is_public);

-- Social Shares
CREATE TABLE IF NOT EXISTS social_shares (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    share_type VARCHAR(50) NOT NULL CHECK (share_type IN ('team_composition', 'contest_win', 'achievement', 'challenge')),
    platform VARCHAR(50) NOT NULL CHECK (platform IN ('twitter', 'facebook', 'whatsapp', 'instagram')),
    content_id BIGINT,
    share_data JSONB NOT NULL,
    share_url VARCHAR(1000),
    click_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_social_shares_user ON social_shares(user_id);
CREATE INDEX IF NOT EXISTS idx_social_shares_type ON social_shares(share_type);
CREATE INDEX IF NOT EXISTS idx_social_shares_platform ON social_shares(platform);
CREATE INDEX IF NOT EXISTS idx_social_shares_created ON social_shares(created_at DESC);

-- Tournament Brackets
CREATE TABLE IF NOT EXISTS tournament_brackets (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    stage_id BIGINT NOT NULL,
    bracket_type VARCHAR(50) DEFAULT 'single_elimination' CHECK (bracket_type IN ('single_elimination', 'double_elimination', 'round_robin', 'swiss')),
    bracket_data JSONB NOT NULL,
    current_round INTEGER DEFAULT 1,
    total_rounds INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'setup' CHECK (status IN ('setup', 'active', 'completed', 'cancelled')),
    auto_advance BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    FOREIGN KEY (stage_id) REFERENCES tournament_stages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tournament_brackets_tournament ON tournament_brackets(tournament_id);
CREATE INDEX IF NOT EXISTS idx_tournament_brackets_stage ON tournament_brackets(stage_id);
CREATE INDEX IF NOT EXISTS idx_tournament_brackets_status ON tournament_brackets(status);

-- Player Performance Predictions
CREATE TABLE IF NOT EXISTS player_predictions (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    match_id BIGINT NOT NULL,
    prediction_date DATE DEFAULT CURRENT_DATE,
    predicted_points DECIMAL(6,2) NOT NULL,
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 1),
    factors JSONB NOT NULL,
    actual_points DECIMAL(6,2),
    accuracy_score DECIMAL(3,2),
    model_version VARCHAR(20) DEFAULT '1.0',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    UNIQUE(player_id, match_id, prediction_date)
);

CREATE INDEX IF NOT EXISTS idx_player_predictions_player ON player_predictions(player_id);
CREATE INDEX IF NOT EXISTS idx_player_predictions_match ON player_predictions(match_id);
CREATE INDEX IF NOT EXISTS idx_player_predictions_date ON player_predictions(prediction_date DESC);
CREATE INDEX IF NOT EXISTS idx_player_predictions_confidence ON player_predictions(confidence_score DESC);

-- Advanced Game Analytics
CREATE TABLE IF NOT EXISTS game_analytics_advanced (
    id BIGSERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL,
    date DATE DEFAULT CURRENT_DATE,
    metric_type VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES games(id),
    UNIQUE(game_id, date, metric_type)
);

CREATE INDEX IF NOT EXISTS idx_game_analytics_advanced_game ON game_analytics_advanced(game_id);
CREATE INDEX IF NOT EXISTS idx_game_analytics_advanced_date ON game_analytics_advanced(date DESC);
CREATE INDEX IF NOT EXISTS idx_game_analytics_advanced_metric ON game_analytics_advanced(metric_type);

-- Fraud Detection
CREATE TABLE IF NOT EXISTS fraud_alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT NOT NULL,
    detection_data JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'investigating', 'resolved', 'false_positive')),
    assigned_to BIGINT,
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (assigned_to) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_fraud_alerts_user ON fraud_alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_fraud_alerts_type ON fraud_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_fraud_alerts_severity ON fraud_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_fraud_alerts_status ON fraud_alerts(status);
CREATE INDEX IF NOT EXISTS idx_fraud_alerts_created ON fraud_alerts(created_at DESC);

-- User Behavior Tracking
CREATE TABLE IF NOT EXISTS user_behavior_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    session_id VARCHAR(100),
    action VARCHAR(100) NOT NULL,
    context_data JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_user_behavior_logs_user ON user_behavior_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_behavior_logs_session ON user_behavior_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_user_behavior_logs_action ON user_behavior_logs(action);
CREATE INDEX IF NOT EXISTS idx_user_behavior_logs_created ON user_behavior_logs(created_at DESC);

-- Insert default achievements
INSERT INTO achievements (name, description, badge_icon, category, trigger_type, trigger_criteria, reward_type, reward_value, created_by) VALUES
('First Team Creator', 'Create your first fantasy team', '🏆', 'progression', 'first_team', '{"teams_created": 1}', 'bonus', 50.00, 1),
('Contest Champion', 'Win your first contest', '🥇', 'gameplay', 'contest_win', '{"contests_won": 1}', 'bonus', 100.00, 1),
('Referral Master', 'Refer 5 friends successfully', '👥', 'social', 'referral_milestone', '{"successful_referrals": 5}', 'bonus', 250.00, 1),
('Winning Streak', 'Win 3 contests in a row', '🔥', 'gameplay', 'winning_streak', '{"consecutive_wins": 3}', 'bonus', 500.00, 1),
('Social Butterfly', 'Add 10 friends', '🦋', 'social', 'custom', '{"friends_added": 10}', 'bonus', 75.00, 1),
('High Roller', 'Join a contest with 1000+ entry fee', '💎', 'progression', 'custom', '{"high_entry_fee": 1000}', 'badge', 0, 1),
('Team Captain', 'Create 100 teams', '⚡', 'progression', 'custom', '{"teams_created": 100}', 'title', 0, 1)
ON CONFLICT DO NOTHING;
`