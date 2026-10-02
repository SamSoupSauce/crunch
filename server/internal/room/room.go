package room

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	ErrRoomFull        = errors.New("no empty slots")
	ErrNoEmptySlots    = errors.New("no empty slots")
	ErrSlotTaken       = errors.New("slot already taken")
	ErrNotYourTurn     = errors.New("not your turn")
	ErrGameNotStarted  = errors.New("game has not started yet; waiting for second player to join")
	ErrRoomExpired     = errors.New("room has expired")
	ErrInvalidPlayer   = errors.New("invalid player number; must be 1 or 2")
)

const (
	RoomTTL           = 24 * time.Hour
	InactivityTimeout = 5 * time.Minute
)

// Message represents a message exchanged in a room.
type Message struct {
	ID        int       `json:"id"`
	Player    int       `json:"player"`
	Content   any       `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Turn represents a move/turn taken by a player.
type Turn struct {
	TurnNumber int       `json:"turnNumber"`
	Player     int       `json:"player"`
	Data       any       `json:"data"` // No validation
	Timestamp  time.Time `json:"timestamp"`
}

// Room represents a multiplayer room with 24-hour lifetime.
type Room struct {
	ID          string            `json:"id"`
	CreatedAt   time.Time         `json:"createdAt"`
	ExpiresAt   time.Time         `json:"expiresAt"`
	PlayerCount int               `json:"playerCount"`
	CurrentTurn int               `json:"currentTurn"` // 1 goes first once 2 players join
	Messages    []Message         `json:"messages"`
	Turns       []Turn            `json:"turns"`
	P1Token     string            `json:"-"`
	P2Token     string            `json:"-"`
	Tokens      map[string]int    `json:"-"` // token -> player number
	LastMoveAt  time.Time         `json:"lastMoveAt"`
	GameStartAt time.Time         `json:"gameStartAt"`

	mu sync.RWMutex
}

// NewRoom creates a new room with a 24-hour expiration.
func NewRoom(id string) *Room {
	now := time.Now()
	return &Room{
		ID:          id,
		CreatedAt:   now,
		ExpiresAt:   now.Add(RoomTTL),
		PlayerCount: 0,
		CurrentTurn: 0,
		Messages:    make([]Message, 0),
		Turns:       make([]Turn, 0),
		Tokens:      make(map[string]int),
	}
}

// Join adds a player to the room, defaulting to any open slot.
func (r *Room) Join() (int, string, error) {
	return r.JoinSlot(0)
}

// JoinSlot assigns a requested slot (1 or 2), or the first open slot if requestedSlot is 0.
// Enforces:
// 1. Host is always player 1 (if no players, slot 1 is assigned).
// 2. Player 2 is the only one if player 1 has joined.
// 3. If both players joined, a slot only opens if the last move was > 5 minutes ago.
func (r *Room) JoinSlot(requestedSlot int) (int, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if time.Now().After(r.ExpiresAt) {
		return 0, "", ErrRoomExpired
	}

	available := r.openSlotsLocked()
	if len(available) == 0 {
		return 0, "", ErrNoEmptySlots
	}

	var slot int
	if requestedSlot == 0 {
		slot = available[0]
	} else {
		isAvail := false
		for _, s := range available {
			if s == requestedSlot {
				isAvail = true
				break
			}
		}
		if !isAvail {
			if requestedSlot == 1 || requestedSlot == 2 {
				return 0, "", ErrSlotTaken
			}
			return 0, "", ErrInvalidPlayer
		}
		slot = requestedSlot
	}

	token := generateToken()
	now := time.Now()

	if slot == 1 {
		if r.P1Token != "" {
			delete(r.Tokens, r.P1Token)
		}
		r.P1Token = token
	} else if slot == 2 {
		if r.P2Token != "" {
			delete(r.Tokens, r.P2Token)
		}
		r.P2Token = token
	}

	r.Tokens[token] = slot

	// Once 2 players have joined, player 1 goes first
	if r.P1Token != "" && r.P2Token != "" {
		r.PlayerCount = 2
		if r.CurrentTurn == 0 {
			r.CurrentTurn = 1
		}
		if r.GameStartAt.IsZero() {
			r.GameStartAt = now
		}
		// Reset inactivity timer for the newly joined occupant
		r.LastMoveAt = now
	} else {
		r.PlayerCount = 1
	}

	return slot, token, nil
}

// OpenSlots returns the unassigned or reopened player numbers (e.g. [1], [2], or []).
func (r *Room) OpenSlots() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.openSlotsLocked()
}

func (r *Room) openSlotsLocked() []int {
	// 1. Host is always player 1: If player 1 has not joined, only slot 1 is available.
	if r.P1Token == "" {
		return []int{1}
	}

	// 2. Player 2 is the only one if player 1 has joined.
	if r.P2Token == "" {
		return []int{2}
	}

	// 3. Both players have joined:
	// "If the last move was more than 5 minutes ago, a slot opens."
	lastActivity := r.lastActivityTimeLocked()
	if time.Since(lastActivity) > InactivityTimeout {
		stalledSlot := r.CurrentTurn
		if stalledSlot != 1 && stalledSlot != 2 {
			stalledSlot = 1
		}
		return []int{stalledSlot}
	}

	return []int{}
}

func (r *Room) lastActivityTimeLocked() time.Time {
	if !r.LastMoveAt.IsZero() {
		return r.LastMoveAt
	}
	if len(r.Turns) > 0 {
		return r.Turns[len(r.Turns)-1].Timestamp
	}
	if !r.GameStartAt.IsZero() {
		return r.GameStartAt
	}
	return r.CreatedAt
}

// LastActivityTime returns the time of the last move or game start.
func (r *Room) LastActivityTime() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastActivityTimeLocked()
}

// SlotOpenedByInactivity returns true if both players had joined but a slot opened due to >5m inactivity.
func (r *Room) SlotOpenedByInactivity() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.P1Token != "" && r.P2Token != "" && time.Since(r.lastActivityTimeLocked()) > InactivityTimeout
}

// TakeTurn records a turn for the specified player.
// Only works if the player number is correct. No validation on the move data.
func (r *Room) TakeTurn(player int, data any) (*Turn, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if time.Now().After(r.ExpiresAt) {
		return nil, ErrRoomExpired
	}

	if r.PlayerCount < 2 {
		return nil, ErrGameNotStarted
	}

	if player != 1 && player != 2 {
		return nil, ErrInvalidPlayer
	}

	if player != r.CurrentTurn {
		return nil, fmt.Errorf("%w: current turn is player %d", ErrNotYourTurn, r.CurrentTurn)
	}

	now := time.Now()
	turn := Turn{
		TurnNumber: len(r.Turns) + 1,
		Player:     player,
		Data:       data,
		Timestamp:  now,
	}

	r.Turns = append(r.Turns, turn)
	r.LastMoveAt = now

	// Alternate turn: 1 -> 2, 2 -> 1
	if r.CurrentTurn == 1 {
		r.CurrentTurn = 2
	} else {
		r.CurrentTurn = 1
	}

	return &turn, nil
}

// AddMessage records a message in the room.
func (r *Room) AddMessage(player int, content any) (*Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if time.Now().After(r.ExpiresAt) {
		return nil, ErrRoomExpired
	}

	if player != 1 && player != 2 {
		return nil, ErrInvalidPlayer
	}

	msg := Message{
		ID:        len(r.Messages) + 1,
		Player:    player,
		Content:   content,
		Timestamp: time.Now(),
	}

	r.Messages = append(r.Messages, msg)
	return &msg, nil
}

// GetMessages returns messages, optionally filtering by ID greater than sinceID.
func (r *Room) GetMessages(sinceID int) ([]Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if time.Now().After(r.ExpiresAt) {
		return nil, ErrRoomExpired
	}

	if sinceID <= 0 {
		msgs := make([]Message, len(r.Messages))
		copy(msgs, r.Messages)
		return msgs, nil
	}

	var msgs []Message
	for _, m := range r.Messages {
		if m.ID > sinceID {
			msgs = append(msgs, m)
		}
	}
	return msgs, nil
}

// GetTurns returns turns, optionally filtering by TurnNumber greater than sinceTurn.
func (r *Room) GetTurns(sinceTurn int) ([]Turn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if time.Now().After(r.ExpiresAt) {
		return nil, ErrRoomExpired
	}

	if sinceTurn <= 0 {
		turns := make([]Turn, len(r.Turns))
		copy(turns, r.Turns)
		return turns, nil
	}

	var turns []Turn
	for _, t := range r.Turns {
		if t.TurnNumber > sinceTurn {
			turns = append(turns, t)
		}
	}
	return turns, nil
}

// Snapshot returns a thread-safe copy of room metadata.
func (r *Room) Snapshot() (id string, count int, currentTurn int, createdAt, expiresAt time.Time, turnCount int, msgCount int, openSlots []int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slots := r.openSlotsLocked()
	return r.ID, r.PlayerCount, r.CurrentTurn, r.CreatedAt, r.ExpiresAt, len(r.Turns), len(r.Messages), slots
}

// GetPlayerFromToken resolves a player token if provided.
func (r *Room) GetPlayerFromToken(token string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.Tokens[token]
	return p, ok
}

// Store coordinates all active rooms and handles 24h expiration cleanup.
type Store struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewStore creates a new Room Store.
func NewStore() *Store {
	return &Store{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom generates a new room with 24-hour expiration.
func (s *Store) CreateRoom() *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	var id string
	for {
		id = generateRoomID()
		if _, exists := s.rooms[id]; !exists {
			break
		}
	}

	room := NewRoom(id)
	s.rooms[id] = room
	return room
}

// GetRoom retrieves a room by ID, automatically removing it if expired.
func (s *Store) GetRoom(id string) (*Room, error) {
	s.mu.RLock()
	room, exists := s.rooms[id]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("room not found")
	}

	if time.Now().After(room.ExpiresAt) {
		s.DeleteRoom(id)
		return nil, ErrRoomExpired
	}

	return room, nil
}

// DeleteRoom deletes a room from the store.
func (s *Store) DeleteRoom(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, id)
}

// CleanupExpired deletes all rooms whose 24h lifetime has passed.
func (s *Store) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0
	for id, room := range s.rooms {
		if now.After(room.ExpiresAt) {
			delete(s.rooms, id)
			count++
		}
	}
	return count
}

// StartJanitor starts a periodic cleanup loop for expired rooms.
func (s *Store) StartJanitor(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CleanupExpired()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func generateRoomID() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return fmt.Sprintf("ROOM-%s", string(b))
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
