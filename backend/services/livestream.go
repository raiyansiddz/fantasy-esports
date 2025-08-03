package services

import (
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"fantasy-esports-backend/pkg/logger"
)

type LiveStreamService struct {
	db *sql.DB
}

// LiveStream represents a live stream configuration
type LiveStream struct {
	ID               int64     `json:"id"`
	MatchID          int64     `json:"match_id"`
	StreamURL        string    `json:"stream_url"`
	EmbedURL         string    `json:"embed_url"`
	Platform         string    `json:"platform"`
	StreamKey        *string   `json:"stream_key,omitempty"`
	IsActive         bool      `json:"is_active"`
	ViewerCount      *int      `json:"viewer_count,omitempty"`
	StreamTitle      *string   `json:"stream_title,omitempty"`
	StreamDescription *string  `json:"stream_description,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StreamPlatform represents supported streaming platforms
type StreamPlatform struct {
	Name         string `json:"name"`
	EmbedPattern string `json:"embed_pattern"`
	URLPattern   string `json:"url_pattern"`
}

// NewLiveStreamService creates a new live stream service instance
func NewLiveStreamService(db *sql.DB) *LiveStreamService {
	// Create match_streams table if it doesn't exist
	createTable := `
		CREATE TABLE IF NOT EXISTS match_streams (
			id BIGSERIAL PRIMARY KEY,
			match_id BIGINT NOT NULL,
			stream_url VARCHAR(500) NOT NULL,
			embed_url VARCHAR(500),
			platform VARCHAR(50) NOT NULL,
			stream_key VARCHAR(200),
			is_stream_active BOOLEAN DEFAULT false,
			viewer_count INTEGER,
			stream_title VARCHAR(200),
			stream_description TEXT,
			started_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
			UNIQUE(match_id)
		);

		CREATE INDEX IF NOT EXISTS idx_match_streams_match_id ON match_streams(match_id);
		CREATE INDEX IF NOT EXISTS idx_match_streams_active ON match_streams(is_stream_active);
	`
	
	_, err := db.Exec(createTable)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create match_streams table: %v", err))
	}

	return &LiveStreamService{
		db: db,
	}
}

// GetSupportedPlatforms returns list of supported streaming platforms
func (s *LiveStreamService) GetSupportedPlatforms() []StreamPlatform {
	return []StreamPlatform{
		{
			Name:         "youtube",
			EmbedPattern: "https://www.youtube.com/embed/%s",
			URLPattern:   `(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]{11})`,
		},
		{
			Name:         "twitch",
			EmbedPattern: "https://player.twitch.tv/?channel=%s&parent=yourdomain.com",
			URLPattern:   `twitch\.tv/([a-zA-Z0-9_]{4,25})`,
		},
		{
			Name:         "facebook",
			EmbedPattern: "https://www.facebook.com/plugins/video.php?href=%s",
			URLPattern:   `facebook\.com/.*?/videos/(\d+)`,
		},
	}
}

// SetMatchLiveStream configures live stream for a match
func (s *LiveStreamService) SetMatchLiveStream(matchID int64, streamURL string, streamTitle *string, streamDescription *string) (*LiveStream, error) {
	logger.Info(fmt.Sprintf("Setting live stream for match %d: %s", matchID, streamURL))

	// Validate and detect platform
	platform, embedURL, err := s.detectPlatformAndGenerateEmbed(streamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid stream URL: %w", err)
	}

	// Check if match exists
	var matchExists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM matches WHERE id = $1)", matchID).Scan(&matchExists)
	if err != nil || !matchExists {
		return nil, fmt.Errorf("match not found")
	}

	// Insert or update stream configuration
	var stream LiveStream
	err = s.db.QueryRow(`
		INSERT INTO match_streams (match_id, stream_url, embed_url, platform, stream_title, stream_description, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (match_id) 
		DO UPDATE SET 
			stream_url = EXCLUDED.stream_url,
			embed_url = EXCLUDED.embed_url,
			platform = EXCLUDED.platform,
			stream_title = EXCLUDED.stream_title,
			stream_description = EXCLUDED.stream_description,
			updated_at = NOW()
		RETURNING id, match_id, stream_url, embed_url, platform, stream_key, 
				  is_stream_active, viewer_count, stream_title, stream_description, 
				  started_at, created_at, updated_at`,
		matchID, streamURL, embedURL, platform, streamTitle, streamDescription).Scan(
		&stream.ID, &stream.MatchID, &stream.StreamURL, &stream.EmbedURL,
		&stream.Platform, &stream.StreamKey, &stream.IsActive, &stream.ViewerCount,
		&stream.StreamTitle, &stream.StreamDescription, &stream.StartedAt,
		&stream.CreatedAt, &stream.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to save stream configuration: %w", err)
	}

	return &stream, nil
}

// ActivateMatchStream activates/deactivates a match stream
func (s *LiveStreamService) ActivateMatchStream(matchID int64, activate bool) error {
	var startedAt *time.Time
	if activate {
		now := time.Now()
		startedAt = &now
	}

	_, err := s.db.Exec(`
		UPDATE match_streams 
		SET is_stream_active = $1, started_at = $2, updated_at = NOW()
		WHERE match_id = $3`,
		activate, startedAt, matchID)

	if err != nil {
		return fmt.Errorf("failed to update stream status: %w", err)
	}

	status := "deactivated"
	if activate {
		status = "activated"
	}
	logger.Info(fmt.Sprintf("Live stream %s for match %d", status, matchID))

	return nil
}

// GetMatchLiveStream retrieves live stream configuration for a match
func (s *LiveStreamService) GetMatchLiveStream(matchID int64) (*LiveStream, error) {
	var stream LiveStream
	err := s.db.QueryRow(`
		SELECT id, match_id, stream_url, embed_url, platform, stream_key,
			   is_stream_active, viewer_count, stream_title, stream_description,
			   started_at, created_at, updated_at
		FROM match_streams 
		WHERE match_id = $1`, matchID).Scan(
		&stream.ID, &stream.MatchID, &stream.StreamURL, &stream.EmbedURL,
		&stream.Platform, &stream.StreamKey, &stream.IsActive, &stream.ViewerCount,
		&stream.StreamTitle, &stream.StreamDescription, &stream.StartedAt,
		&stream.CreatedAt, &stream.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no live stream configured for match")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get stream configuration: %w", err)
	}

	return &stream, nil
}

// RemoveMatchLiveStream removes live stream configuration for a match
func (s *LiveStreamService) RemoveMatchLiveStream(matchID int64) error {
	result, err := s.db.Exec("DELETE FROM match_streams WHERE match_id = $1", matchID)
	if err != nil {
		return fmt.Errorf("failed to remove stream configuration: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no stream configuration found for match")
	}

	logger.Info(fmt.Sprintf("Removed live stream configuration for match %d", matchID))
	return nil
}

// GetActiveLiveStreams retrieves all currently active live streams
func (s *LiveStreamService) GetActiveLiveStreams() ([]LiveStream, error) {
	rows, err := s.db.Query(`
		SELECT ms.id, ms.match_id, ms.stream_url, ms.embed_url, ms.platform, ms.stream_key,
			   ms.is_stream_active, ms.viewer_count, ms.stream_title, ms.stream_description,
			   ms.started_at, ms.created_at, ms.updated_at
		FROM match_streams ms
		JOIN matches m ON ms.match_id = m.id
		WHERE ms.is_stream_active = true
		ORDER BY ms.started_at DESC`)

	if err != nil {
		return nil, fmt.Errorf("failed to get active streams: %w", err)
	}
	defer rows.Close()

	// Initialize streams slice to ensure it's never nil  
	streams := make([]LiveStream, 0)
	for rows.Next() {
		var stream LiveStream
		err := rows.Scan(
			&stream.ID, &stream.MatchID, &stream.StreamURL, &stream.EmbedURL,
			&stream.Platform, &stream.StreamKey, &stream.IsActive, &stream.ViewerCount,
			&stream.StreamTitle, &stream.StreamDescription, &stream.StartedAt,
			&stream.CreatedAt, &stream.UpdatedAt,
		)
		if err != nil {
			continue
		}
		streams = append(streams, stream)
	}

	return streams, nil
}

// UpdateStreamViewerCount updates the viewer count for a stream
func (s *LiveStreamService) UpdateStreamViewerCount(matchID int64, viewerCount int) error {
	_, err := s.db.Exec(`
		UPDATE match_streams 
		SET viewer_count = $1, updated_at = NOW()
		WHERE match_id = $2 AND is_stream_active = true`,
		viewerCount, matchID)

	if err != nil {
		return fmt.Errorf("failed to update viewer count: %w", err)
	}

	return nil
}

// detectPlatformAndGenerateEmbed detects the streaming platform and generates embed URL
func (s *LiveStreamService) detectPlatformAndGenerateEmbed(streamURL string) (platform string, embedURL string, err error) {
	// Validate URL format
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL format")
	}

	// Normalize URL
	normalizedURL := strings.ToLower(parsedURL.String())
	
	platforms := s.GetSupportedPlatforms()
	
	for _, p := range platforms {
		regex, err := regexp.Compile(p.URLPattern)
		if err != nil {
			continue
		}
		
		matches := regex.FindStringSubmatch(normalizedURL)
		if len(matches) > 1 {
			// Extract video/stream ID and generate embed URL
			videoID := matches[1]
			embedURL := fmt.Sprintf(p.EmbedPattern, videoID)
			return p.Name, embedURL, nil
		}
	}
	
	// If no specific platform detected, return as generic stream
	return "generic", streamURL, nil
}

// ValidateStreamURL validates if a stream URL is accessible
func (s *LiveStreamService) ValidateStreamURL(streamURL string) error {
	// Basic URL validation
	_, err := url.Parse(streamURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if URL is from supported platform
	_, _, err = s.detectPlatformAndGenerateEmbed(streamURL)
	if err != nil {
		return fmt.Errorf("unsupported platform or invalid URL: %w", err)
	}

	return nil
}

// GetMatchStreamStats returns streaming statistics for a match
func (s *LiveStreamService) GetMatchStreamStats(matchID int64) (map[string]interface{}, error) {
	stream, err := s.GetMatchLiveStream(matchID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"match_id":         stream.MatchID,
		"platform":         stream.Platform,
		"is_active":        stream.IsActive,
		"viewer_count":     stream.ViewerCount,
		"started_at":       stream.StartedAt,
		"stream_duration":  nil,
	}

	// Calculate stream duration if active and started
	if stream.IsActive && stream.StartedAt != nil {
		duration := time.Since(*stream.StartedAt)
		stats["stream_duration"] = duration.String()
	}

	return stats, nil
}

// AutoDeactivateExpiredStreams deactivates streams for completed matches
func (s *LiveStreamService) AutoDeactivateExpiredStreams() error {
	result, err := s.db.Exec(`
		UPDATE match_streams 
		SET is_stream_active = false, updated_at = NOW()
		FROM matches 
		WHERE match_streams.match_id = matches.id 
		AND matches.status = 'completed' 
		AND match_streams.is_stream_active = true`)

	if err != nil {
		return fmt.Errorf("failed to deactivate expired streams: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Info(fmt.Sprintf("Auto-deactivated %d expired live streams", rowsAffected))
	}

	return nil
}