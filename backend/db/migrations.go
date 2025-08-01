package db

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    mobile VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    date_of_birth DATE,
    gender VARCHAR(10),
    avatar_url VARCHAR(500),
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    account_status VARCHAR(20) DEFAULT 'active',
    kyc_status VARCHAR(20) DEFAULT 'pending',
    referral_code VARCHAR(20) UNIQUE,
    referred_by_code VARCHAR(20),
    state VARCHAR(50),
    city VARCHAR(50),
    pincode VARCHAR(10),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_mobile ON users(mobile);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code);
CREATE INDEX IF NOT EXISTS idx_users_kyc_status ON users(kyc_status);
`

const createKYCDocumentsTable = `
CREATE TABLE IF NOT EXISTS kyc_documents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    document_type VARCHAR(20) NOT NULL,
    document_front_url VARCHAR(500) NOT NULL,
    document_back_url VARCHAR(500),
    document_number VARCHAR(50),
    additional_data JSONB,
    status VARCHAR(20) DEFAULT 'pending',
    verified_at TIMESTAMP,
    verified_by BIGINT,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_kyc_user_doc_type ON kyc_documents(user_id, document_type);
CREATE INDEX IF NOT EXISTS idx_kyc_status ON kyc_documents(status);
`

const createGamesTable = `
CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    category VARCHAR(50),
    description TEXT,
    logo_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    scoring_rules JSONB NOT NULL,
    player_roles JSONB,
    team_composition JSONB,
    min_players_per_team INTEGER DEFAULT 1,
    max_players_per_team INTEGER DEFAULT 2,
    total_team_size INTEGER DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_games_code ON games(code);
CREATE INDEX IF NOT EXISTS idx_games_active ON games(is_active);
`

const createTeamsTable = `
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    short_name VARCHAR(10),
    logo_url VARCHAR(500),
    region VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    social_links JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);
CREATE INDEX IF NOT EXISTS idx_teams_active ON teams(is_active);
`

const createPlayersTable = `
CREATE TABLE IF NOT EXISTS players (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    team_id BIGINT NOT NULL,
    game_id INTEGER NOT NULL,
    role VARCHAR(50),
    credit_value DECIMAL(4,1) NOT NULL DEFAULT 8.0,
    is_playing BOOLEAN DEFAULT TRUE,
    avatar_url VARCHAR(500),
    country VARCHAR(50),
    stats JSONB,
    form_score DECIMAL(3,1) DEFAULT 5.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id),
    FOREIGN KEY (game_id) REFERENCES games(id)
);

CREATE INDEX IF NOT EXISTS idx_players_team_game ON players(team_id, game_id);
CREATE INDEX IF NOT EXISTS idx_players_credit_value ON players(credit_value);
CREATE INDEX IF NOT EXISTS idx_players_is_playing ON players(is_playing);
`

const createTournamentsTable = `
CREATE TABLE IF NOT EXISTS tournaments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    game_id INTEGER NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    prize_pool DECIMAL(12,2),
    total_teams INTEGER,
    status VARCHAR(20) DEFAULT 'upcoming',
    is_featured BOOLEAN DEFAULT FALSE,
    logo_url VARCHAR(500),
    banner_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES games(id)
);

CREATE INDEX IF NOT EXISTS idx_tournaments_game_status ON tournaments(game_id, status);
CREATE INDEX IF NOT EXISTS idx_tournaments_dates ON tournaments(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_tournaments_featured ON tournaments(is_featured);
`

const createTournamentStagesTable = `
CREATE TABLE IF NOT EXISTS tournament_stages (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    stage_order INTEGER NOT NULL,
    stage_type VARCHAR(20) NOT NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    max_teams INTEGER,
    rules JSONB,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tournament_stages_tournament_order ON tournament_stages(tournament_id, stage_order);
`

const createMatchesTable = `
CREATE TABLE IF NOT EXISTS matches (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT,
    stage_id BIGINT,
    game_id INTEGER NOT NULL,
    name VARCHAR(200),
    scheduled_at TIMESTAMP NOT NULL,
    lock_time TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'upcoming',
    match_type VARCHAR(20) DEFAULT 'elimination',
    map VARCHAR(50),
    best_of INTEGER DEFAULT 1,
    result JSONB,
    winner_team_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id),
    FOREIGN KEY (stage_id) REFERENCES tournament_stages(id),
    FOREIGN KEY (game_id) REFERENCES games(id),
    FOREIGN KEY (winner_team_id) REFERENCES teams(id)
);

CREATE INDEX IF NOT EXISTS idx_matches_scheduled ON matches(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_matches_status_game ON matches(status, game_id);
CREATE INDEX IF NOT EXISTS idx_matches_tournament ON matches(tournament_id);
`

const createMatchParticipantsTable = `
CREATE TABLE IF NOT EXISTS match_participants (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    seed INTEGER,
    final_position INTEGER,
    team_score INTEGER DEFAULT 0,
    points_earned DECIMAL(8,2) DEFAULT 0,
    eliminated_at TIMESTAMP,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id),
    UNIQUE(match_id, team_id)
);

CREATE INDEX IF NOT EXISTS idx_match_participants_match_position ON match_participants(match_id, final_position);
`

const createMatchEventsTable = `
CREATE TABLE IF NOT EXISTS match_events (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL,
    player_id BIGINT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    points DECIMAL(5,2) NOT NULL,
    round_number INTEGER,
    game_time VARCHAR(20),
    description TEXT,
    additional_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT NOT NULL,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_match_events_match_player ON match_events(match_id, player_id);
CREATE INDEX IF NOT EXISTS idx_match_events_event_type ON match_events(event_type);
CREATE INDEX IF NOT EXISTS idx_match_events_created_at ON match_events(created_at);
`

const createContestsTable = `
CREATE TABLE IF NOT EXISTS contests (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    contest_type VARCHAR(20) DEFAULT 'public',
    entry_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    max_participants INTEGER NOT NULL,
    current_participants INTEGER DEFAULT 0,
    total_prize_pool DECIMAL(12,2) NOT NULL,
    is_guaranteed BOOLEAN DEFAULT FALSE,
    prize_distribution JSONB NOT NULL,
    contest_rules JSONB,
    status VARCHAR(20) DEFAULT 'upcoming',
    invite_code VARCHAR(20) UNIQUE,
    is_multi_entry BOOLEAN DEFAULT FALSE,
    max_entries_per_user INTEGER DEFAULT 1,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (match_id) REFERENCES matches(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_contests_match_status ON contests(match_id, status);
CREATE INDEX IF NOT EXISTS idx_contests_entry_fee ON contests(entry_fee);
CREATE INDEX IF NOT EXISTS idx_contests_invite_code ON contests(invite_code);
`

const createUserTeamsTable = `
CREATE TABLE IF NOT EXISTS user_teams (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    match_id BIGINT NOT NULL,
    team_name VARCHAR(100) NOT NULL,
    captain_player_id BIGINT NOT NULL,
    vice_captain_player_id BIGINT NOT NULL,
    total_credits_used DECIMAL(5,1) NOT NULL,
    total_points DECIMAL(8,2) DEFAULT 0,
    final_rank INTEGER DEFAULT 0,
    is_locked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (match_id) REFERENCES matches(id),
    FOREIGN KEY (captain_player_id) REFERENCES players(id),
    FOREIGN KEY (vice_captain_player_id) REFERENCES players(id)
);

CREATE INDEX IF NOT EXISTS idx_user_teams_user_match ON user_teams(user_id, match_id);
CREATE INDEX IF NOT EXISTS idx_user_teams_match_points ON user_teams(match_id, total_points DESC);
`

const createTeamPlayersTable = `
CREATE TABLE IF NOT EXISTS team_players (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL,
    player_id BIGINT NOT NULL,
    real_team_id BIGINT NOT NULL,
    is_captain BOOLEAN DEFAULT FALSE,
    is_vice_captain BOOLEAN DEFAULT FALSE,
    points_earned DECIMAL(6,2) DEFAULT 0,
    FOREIGN KEY (team_id) REFERENCES user_teams(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id),
    FOREIGN KEY (real_team_id) REFERENCES teams(id),
    UNIQUE(team_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_team_players_real_team ON team_players(real_team_id);
`

const createContestParticipantsTable = `
CREATE TABLE IF NOT EXISTS contest_participants (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    entry_fee_paid DECIMAL(10,2) NOT NULL,
    rank INTEGER DEFAULT 0,
    prize_won DECIMAL(10,2) DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (contest_id) REFERENCES contests(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (team_id) REFERENCES user_teams(id),
    UNIQUE(contest_id, team_id)
);

CREATE INDEX IF NOT EXISTS idx_contest_participants_contest_rank ON contest_participants(contest_id, rank);
CREATE INDEX IF NOT EXISTS idx_contest_participants_user_contests ON contest_participants(user_id, joined_at DESC);
`

const createUserWalletsTable = `
CREATE TABLE IF NOT EXISTS user_wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    bonus_balance DECIMAL(12,2) DEFAULT 0,
    deposit_balance DECIMAL(12,2) DEFAULT 0,
    winning_balance DECIMAL(12,2) DEFAULT 0,
    total_balance DECIMAL(12,2) GENERATED ALWAYS AS (bonus_balance + deposit_balance + winning_balance) STORED,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_wallets_total_balance ON user_wallets(total_balance);
`

const createWalletTransactionsTable = `
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    balance_type VARCHAR(10) NOT NULL,
    description TEXT,
    reference_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    gateway_response JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_type_status ON wallet_transactions(user_id, transaction_type, status);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference ON wallet_transactions(reference_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at DESC);
`

const createPaymentTransactionsTable = `
CREATE TABLE IF NOT EXISTS payment_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    transaction_id VARCHAR(100) UNIQUE NOT NULL,
    gateway VARCHAR(50) NOT NULL,
    gateway_transaction_id VARCHAR(200),
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    type VARCHAR(10) NOT NULL,
    status VARCHAR(20) DEFAULT 'initiated',
    gateway_response JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_payment_transactions_user_status ON payment_transactions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_gateway_txn ON payment_transactions(gateway_transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_transaction_id ON payment_transactions(transaction_id);
`

const createReferralsTable = `
CREATE TABLE IF NOT EXISTS referrals (
    id BIGSERIAL PRIMARY KEY,
    referrer_user_id BIGINT NOT NULL,
    referred_user_id BIGINT NOT NULL,
    referral_code VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    reward_amount DECIMAL(10,2) DEFAULT 0,
    completion_criteria VARCHAR(50),
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (referrer_user_id) REFERENCES users(id),
    FOREIGN KEY (referred_user_id) REFERENCES users(id),
    UNIQUE(referrer_user_id, referred_user_id)
);

CREATE INDEX IF NOT EXISTS idx_referrals_referrer ON referrals(referrer_user_id);
CREATE INDEX IF NOT EXISTS idx_referrals_status ON referrals(status);
`

const createAdminUsersTable = `
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    role VARCHAR(20) NOT NULL,
    permissions JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
`

const createSystemConfigTable = `
CREATE TABLE IF NOT EXISTS system_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value JSONB NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    updated_by BIGINT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id)
);

CREATE INDEX IF NOT EXISTS idx_system_config_config_key ON system_config(config_key);
`

const insertDefaultConfigs = `
INSERT INTO system_config (config_key, config_value, description) VALUES
('app_settings', '{"maintenance_mode": false, "min_app_version": "1.0.0", "force_update": false}', 'App global settings'),
('payment_settings', '{"min_deposit": 10, "max_deposit": 10000, "withdrawal_fee": 5}', 'Payment configuration'),
('contest_settings', '{"max_teams_per_user": 10, "edit_deadline_minutes": 10}', 'Contest rules'),
('scoring_settings', '{"auto_calculate": true, "manual_override": true}', 'Scoring system settings')
ON CONFLICT (config_key) DO NOTHING;
`

const insertSampleData = `
-- Insert sample games
INSERT INTO games (name, code, category, description, is_active, scoring_rules, player_roles, team_composition, total_team_size) VALUES
('Valorant', 'VAL', 'fps', 'Tactical FPS by Riot Games', true, 
'{"kill": 2, "death": -1, "assist": 1, "plant_defuse": 2, "first_blood": 3, "ace": 10, "mvp": 5}',
'["Duelist", "Controller", "Initiator", "Sentinel"]',
'{"total_players": 5, "min_per_role": 1, "max_per_role": 2}', 5),
('BGMI', 'BGMI', 'battle_royale', 'Battle Royale mobile game', true,
'{"kill": 2, "knock": 1, "revive": 1, "death": -2, "placement": {"1": 15, "2-5": 10, "6-10": 6}}',
'["Fragger", "Support", "IGL", "Sniper"]',
'{"total_players": 4, "min_per_role": 1, "max_per_role": 2}', 4),
('CS2', 'CS2', 'fps', 'Counter-Strike 2', true,
'{"kill": 2, "death": -1, "assist": 1, "plant_defuse": 2, "first_blood": 3, "ace": 10, "mvp": 5}',
'["Entry", "Support", "AWPer", "IGL", "Lurker"]',
'{"total_players": 5, "min_per_role": 1, "max_per_role": 2}', 5)
ON CONFLICT (code) DO NOTHING;

-- Insert sample teams
INSERT INTO teams (name, short_name, region, is_active) VALUES
('Team Liquid', 'TL', 'NA', true),
('Fnatic', 'FNC', 'EU', true),
('Sentinels', 'SEN', 'NA', true),
('LOUD', 'LOUD', 'BR', true),
('Paper Rex', 'PRX', 'APAC', true),
('DRX', 'DRX', 'KR', true)
ON CONFLICT DO NOTHING;

-- Insert sample players for each team and game
INSERT INTO players (name, team_id, game_id, role, credit_value, is_playing, country, stats, form_score) VALUES
-- Valorant players for Team Liquid
(1, 'ScreaM', 1, 1, 'Duelist', 9.5, true, 'Belgium', '{"kills": 18, "deaths": 12, "assists": 6, "headshots": 14, "aces": 1}', 8.5),
(2, 'Nivera', 1, 1, 'Sentinel', 8.5, true, 'France', '{"kills": 15, "deaths": 10, "assists": 8, "headshots": 10, "aces": 0}', 8.0),
(3, 'Jamppi', 1, 1, 'Controller', 8.0, true, 'Finland', '{"kills": 12, "deaths": 14, "assists": 12, "headshots": 6, "aces": 0}', 7.5),
(4, 'soulcas', 1, 1, 'Initiator', 7.5, true, 'UK', '{"kills": 10, "deaths": 13, "assists": 15, "headshots": 4, "aces": 0}', 7.0),
(5, 'Redgar', 1, 1, 'Controller', 7.0, true, 'Russia', '{"kills": 8, "deaths": 15, "assists": 18, "headshots": 3, "aces": 0}', 6.5),

-- Valorant players for Fnatic
(6, 'Boaster', 2, 1, 'Controller', 8.5, true, 'UK', '{"kills": 11, "deaths": 13, "assists": 16, "headshots": 5, "aces": 0}', 7.5),
(7, 'Chronicle', 2, 1, 'Initiator', 9.0, true, 'Russia', '{"kills": 16, "deaths": 11, "assists": 9, "headshots": 12, "aces": 1}', 8.5),
(8, 'Leo', 2, 1, 'Initiator', 8.5, true, 'Finland', '{"kills": 14, "deaths": 12, "assists": 11, "headshots": 8, "aces": 0}', 8.0),
(9, 'Alfajer', 2, 1, 'Sentinel', 9.5, true, 'Turkey', '{"kills": 19, "deaths": 10, "assists": 7, "headshots": 15, "aces": 2}', 9.0),
(10, 'Derke', 2, 1, 'Duelist', 10.0, true, 'Finland', '{"kills": 22, "deaths": 9, "assists": 5, "headshots": 18, "aces": 2}', 9.5),

-- BGMI players for Sentinels  
(11, 'TenZ', 3, 2, 'Fragger', 10.0, true, 'Canada', '{"kills": 25, "knock": 15, "revive": 5, "death": 3, "placement": 1}', 9.5),
(12, 'Sick', 3, 2, 'Support', 8.5, true, 'USA', '{"kills": 18, "knock": 12, "revive": 8, "death": 4, "placement": 2}', 8.0),
(13, 'ShahZaM', 3, 2, 'IGL', 8.0, true, 'USA', '{"kills": 15, "knock": 10, "revive": 6, "death": 5, "placement": 3}', 7.5),
(14, 'dapr', 3, 2, 'Sniper', 8.5, true, 'USA', '{"kills": 20, "knock": 8, "revive": 4, "death": 4, "placement": 1}', 8.5),

-- CS2 players for LOUD
(15, 'aspas', 4, 3, 'Entry', 10.0, true, 'Brazil', '{"kills": 24, "deaths": 11, "assists": 6, "headshots": 16, "aces": 2}', 9.5),
(16, 'Less', 4, 3, 'Support', 8.5, true, 'Brazil', '{"kills": 16, "deaths": 13, "assists": 10, "headshots": 9, "aces": 0}', 8.0),
(17, 'Cauanzin', 4, 3, 'IGL', 8.0, true, 'Brazil', '{"kills": 12, "deaths": 14, "assists": 14, "headshots": 6, "aces": 0}', 7.5),
(18, 'tuyz', 4, 3, 'AWPer', 9.0, true, 'Brazil', '{"kills": 18, "deaths": 12, "assists": 7, "headshots": 14, "aces": 1}', 8.5),
(19, 'Saadhak', 4, 3, 'Lurker', 8.5, true, 'Argentina', '{"kills": 14, "deaths": 13, "assists": 12, "headshots": 8, "aces": 0}', 8.0)
ON CONFLICT DO NOTHING;

-- Insert sample tournaments
INSERT INTO tournaments (name, game_id, description, start_date, end_date, prize_pool, total_teams, status, is_featured, logo_url, banner_url) VALUES
('VCT Masters 2025', 1, 'Premier Valorant tournament featuring top teams from around the world', '2025-08-01 10:00:00', '2025-08-15 23:59:59', 1000000.00, 16, 'upcoming', true, 'https://example.com/vct-logo.png', 'https://example.com/vct-banner.png'),
('BGMI World Championship', 2, 'The ultimate BGMI tournament with the best mobile esports teams', '2025-07-15 09:00:00', '2025-07-30 22:00:00', 750000.00, 32, 'live', true, 'https://example.com/bgmi-logo.png', 'https://example.com/bgmi-banner.png'),
('CS2 Major Championship', 3, 'Counter-Strike 2 Major tournament with legendary teams', '2025-09-01 12:00:00', '2025-09-20 20:00:00', 2000000.00, 24, 'upcoming', true, 'https://example.com/cs2-logo.png', 'https://example.com/cs2-banner.png'),
('VCT Regional Finals', 1, 'Regional qualifying tournament for VCT Masters', '2025-07-20 14:00:00', '2025-07-25 18:00:00', 250000.00, 8, 'live', false, NULL, NULL)
ON CONFLICT DO NOTHING;

-- Insert sample matches
INSERT INTO matches (tournament_id, game_id, name, scheduled_at, lock_time, status, match_type, map, best_of, result) VALUES
(1, 1, 'Team Liquid vs Fnatic - Group A', '2025-08-02 15:00:00', '2025-08-02 14:50:00', 'upcoming', 'group', 'Haven', 3, NULL),
(1, 1, 'Sentinels vs LOUD - Group B', '2025-08-02 18:00:00', '2025-08-02 17:50:00', 'upcoming', 'group', 'Bind', 3, NULL),
(2, 2, 'Squad Alpha vs Team Beta - Quarters', '2025-07-25 12:00:00', '2025-07-25 11:50:00', 'live', 'elimination', 'Erangel', 1, NULL),
(3, 3, 'Championship Finals', '2025-09-15 19:00:00', '2025-09-15 18:50:00', 'upcoming', 'final', 'Dust2', 5, NULL),
(4, 1, 'Regional Semi-Final', '2025-07-24 16:00:00', '2025-07-24 15:50:00', 'live', 'elimination', 'Icebox', 3, NULL)
ON CONFLICT DO NOTHING;

-- Insert match participants (link teams to matches)
INSERT INTO match_participants (match_id, team_id, seed, team_score, points_earned) VALUES
(1, 1, 1, 0, 0.0), -- Team Liquid in match 1
(1, 2, 2, 0, 0.0), -- Fnatic in match 1
(2, 3, 1, 0, 0.0), -- Sentinels in match 2
(2, 4, 2, 0, 0.0), -- LOUD in match 2
(3, 5, 1, 0, 0.0), -- Paper Rex in match 3
(3, 6, 2, 0, 0.0), -- DRX in match 3
(4, 1, 1, 0, 0.0), -- Team Liquid in match 4
(4, 4, 2, 0, 0.0), -- LOUD in match 4
(5, 2, 1, 0, 0.0), -- Fnatic in match 5
(5, 3, 2, 0, 0.0)  -- Sentinels in match 5
ON CONFLICT DO NOTHING;

-- Insert sample contests
INSERT INTO contests (match_id, name, contest_type, entry_fee, max_participants, total_prize_pool, is_guaranteed, prize_distribution, contest_rules, status, created_by) VALUES
(1, 'Mega Contest - TL vs FNC', 'public', 50.00, 10000, 450000.00, true, 
'[{"rank_from": 1, "rank_to": 1, "prize": 150000.00, "percentage": 33.33}, {"rank_from": 2, "rank_to": 10, "prize": 25000.00, "percentage": 55.56}, {"rank_from": 11, "rank_to": 100, "prize": 2750.00, "percentage": 11.11}]',
'{"team_size": 5, "captain_multiplier": 2.0, "vice_captain_multiplier": 1.5, "max_players_per_team": 2, "min_players_per_team": 1, "total_credits": 100}', 'upcoming', 1),

(2, 'Winner Takes All - SEN vs LOUD', 'public', 100.00, 5000, 450000.00, true,
'[{"rank_from": 1, "rank_to": 1, "prize": 300000.00, "percentage": 66.67}, {"rank_from": 2, "rank_to": 5, "prize": 37500.00, "percentage": 33.33}]',
'{"team_size": 4, "captain_multiplier": 2.0, "vice_captain_multiplier": 1.5, "max_players_per_team": 2, "min_players_per_team": 1, "total_credits": 100}', 'upcoming', 1),

(3, 'BGMI Squad Championship', 'public', 25.00, 20000, 400000.00, false,
'[{"rank_from": 1, "rank_to": 1, "prize": 100000.00, "percentage": 25.0}, {"rank_from": 2, "rank_to": 20, "prize": 15000.00, "percentage": 75.0}]',
'{"team_size": 4, "captain_multiplier": 2.0, "vice_captain_multiplier": 1.5, "max_players_per_team": 2, "min_players_per_team": 1, "total_credits": 100}', 'live', 1)
ON CONFLICT DO NOTHING;

-- Insert sample admin user
INSERT INTO admin_users (username, email, password_hash, full_name, role, permissions, is_active) VALUES
('admin', 'admin@fantasy-esports.com', '$2a$10$rQ7gJz5QZ5Z5Z5Z5Z5Z5Zu5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z', 'Super Admin', 'super_admin', 
'{"users": "full", "games": "full", "contests": "full", "scoring": "full", "finance": "full"}', true)
ON CONFLICT (username) DO NOTHING;
`