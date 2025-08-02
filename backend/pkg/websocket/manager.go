package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"fantasy-esports-backend/models"
	"github.com/gorilla/websocket"
)

// ConnectionManager manages WebSocket connections for real-time leaderboards
type ConnectionManager struct {
	connections map[string]*LeaderboardConnection
	contestSubs map[int64][]*LeaderboardConnection // contest_id -> connections
	mutex       sync.RWMutex
	broadcast   chan models.RealTimeLeaderboardUpdate
	register    chan *LeaderboardConnection
	unregister  chan *LeaderboardConnection
}

type LeaderboardConnection struct {
	UserID       int64                    `json:"user_id"`
	ContestID    int64                    `json:"contest_id"`
	ConnectionID string                   `json:"connection_id"`
	Conn         *websocket.Conn          `json:"-"`
	Send         chan models.RealTimeWebSocketMessage `json:"-"`
	ConnectedAt  time.Time                `json:"connected_at"`
	LastPing     time.Time                `json:"last_ping"`
	IsActive     bool                     `json:"is_active"`
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*LeaderboardConnection),
		contestSubs: make(map[int64][]*LeaderboardConnection),
		broadcast:   make(chan models.RealTimeLeaderboardUpdate),
		register:    make(chan *LeaderboardConnection),
		unregister:  make(chan *LeaderboardConnection),
	}
}

func (cm *ConnectionManager) Start() {
	go cm.run()
}

func (cm *ConnectionManager) run() {
	ticker := time.NewTicker(30 * time.Second) // Ping every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case conn := <-cm.register:
			cm.registerConnection(conn)

		case conn := <-cm.unregister:
			cm.unregisterConnection(conn)

		case update := <-cm.broadcast:
			cm.broadcastToContest(update)

		case <-ticker.C:
			cm.cleanupInactiveConnections()
		}
	}
}

func (cm *ConnectionManager) RegisterConnection(conn *LeaderboardConnection) {
	cm.register <- conn
}

func (cm *ConnectionManager) UnregisterConnection(conn *LeaderboardConnection) {
	cm.unregister <- conn
}

func (cm *ConnectionManager) BroadcastUpdate(update models.RealTimeLeaderboardUpdate) {
	cm.broadcast <- update
}

func (cm *ConnectionManager) registerConnection(conn *LeaderboardConnection) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Add to connections map
	cm.connections[conn.ConnectionID] = conn

	// Add to contest subscriptions
	if cm.contestSubs[conn.ContestID] == nil {
		cm.contestSubs[conn.ContestID] = make([]*LeaderboardConnection, 0)
	}
	cm.contestSubs[conn.ContestID] = append(cm.contestSubs[conn.ContestID], conn)

	conn.IsActive = true
	conn.LastPing = time.Now()

	log.Printf("Registered leaderboard connection: User %d for Contest %d", conn.UserID, conn.ContestID)

	// Send connection confirmation
	confirmMsg := models.RealTimeWebSocketMessage{
		Type:      "connection_status",
		ContestID: conn.ContestID,
		Data: models.LeaderboardConnectionStatus{
			Connected: true,
			ContestID: conn.ContestID,
			UserID:    conn.UserID,
		},
		Timestamp: time.Now(),
		MessageID: generateMessageID(),
	}

	select {
	case conn.Send <- confirmMsg:
	default:
		close(conn.Send)
	}
}

func (cm *ConnectionManager) unregisterConnection(conn *LeaderboardConnection) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Remove from connections map
	delete(cm.connections, conn.ConnectionID)

	// Remove from contest subscriptions
	if subs, exists := cm.contestSubs[conn.ContestID]; exists {
		for i, sub := range subs {
			if sub.ConnectionID == conn.ConnectionID {
				cm.contestSubs[conn.ContestID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		// Clean up empty contest subscription lists
		if len(cm.contestSubs[conn.ContestID]) == 0 {
			delete(cm.contestSubs, conn.ContestID)
		}
	}

	conn.IsActive = false
	close(conn.Send)

	log.Printf("Unregistered leaderboard connection: User %d for Contest %d", conn.UserID, conn.ContestID)
}

func (cm *ConnectionManager) broadcastToContest(update models.RealTimeLeaderboardUpdate) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections, exists := cm.contestSubs[update.ContestID]
	if !exists || len(connections) == 0 {
		return
	}

	message := models.RealTimeWebSocketMessage{
		Type:      "leaderboard_update",
		ContestID: update.ContestID,
		Data:      update,
		Timestamp: time.Now(),
		MessageID: generateMessageID(),
	}

	log.Printf("Broadcasting leaderboard update to %d connections for contest %d", len(connections), update.ContestID)

	for _, conn := range connections {
		if !conn.IsActive {
			continue
		}

		select {
		case conn.Send <- message:
		default:
			// Connection is blocked, close it
			cm.unregister <- conn
		}
	}
}

func (cm *ConnectionManager) GetContestConnectionCount(contestID int64) int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if connections, exists := cm.contestSubs[contestID]; exists {
		activeCount := 0
		for _, conn := range connections {
			if conn.IsActive {
				activeCount++
			}
		}
		return activeCount
	}
	return 0
}

func (cm *ConnectionManager) cleanupInactiveConnections() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cutoff := time.Now().Add(-2 * time.Minute) // 2 minutes timeout

	for connID, conn := range cm.connections {
		if conn.LastPing.Before(cutoff) {
			log.Printf("Cleaning up inactive connection: %s", connID)
			cm.unregister <- conn
		}
	}
}

func (cm *ConnectionManager) UpdateConnectionPing(connectionID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if conn, exists := cm.connections[connectionID]; exists {
		conn.LastPing = time.Now()
	}
}

// Helper function to generate unique message IDs
func generateMessageID() string {
	return time.Now().Format("20060102150405.000000")
}

// SendPersonalizedUpdate sends an update specific to a user
func (cm *ConnectionManager) SendPersonalizedUpdate(userID int64, contestID int64, update models.RealTimeWebSocketMessage) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, conn := range cm.connections {
		if conn.UserID == userID && conn.ContestID == contestID && conn.IsActive {
			select {
			case conn.Send <- update:
			default:
				// Connection blocked, will be cleaned up later
			}
			break
		}
	}
}