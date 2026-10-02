package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crunch/server/internal/room"
)

// Server handles HTTP coordination for multiplayer rooms.
type Server struct {
	Store *room.Store
}

// NewServer creates a new API Server instance.
func NewServer(store *room.Store) *Server {
	return &Server{Store: store}
}

// HandleCreateRoom handles POST /rooms.
// Creates a room that lasts 24 hours. The host is always Player 1.
func (s *Server) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rInstance := s.Store.CreateRoom()
	playerNum, token, _ := rInstance.JoinSlot(1)

	writeJSON(w, http.StatusCreated, map[string]any{
		"roomId":    rInstance.ID,
		"player":    playerNum,
		"token":     token,
		"openSlots": rInstance.OpenSlots(),
		"createdAt": rInstance.CreatedAt,
		"expiresAt": rInstance.ExpiresAt,
	})
}

// JoinRoomRequest defines payload when joining a room.
type JoinRoomRequest struct {
	Player int `json:"player,omitempty"` // optional slot preference: 1 or 2
}

// HandleJoinRoom handles POST /rooms/{id}/join.
// Allows picking an open slot (1 or 2), or picks any open slot if player is 0.
// Returns error if no empty slots exist or if requested slot is taken.
func (s *Server) HandleJoinRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rInstance, err := s.Store.GetRoom(roomID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req JoinRoomRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	playerNum, token, err := rInstance.JoinSlot(req.Player)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	openSlots := rInstance.OpenSlots()
	msg := fmt.Sprintf("Joined as player %d.", playerNum)
	if len(openSlots) > 0 {
		msg += fmt.Sprintf(" Slot %d is still open.", openSlots[0])
	} else {
		msg += " Both players joined! Player 1 goes first."
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"roomId":    rInstance.ID,
		"player":    playerNum,
		"token":     token,
		"openSlots": openSlots,
		"message":   msg,
	})
}

// TurnRequest defines the payload for taking a turn.
type TurnRequest struct {
	Player int    `json:"player"`
	Token  string `json:"token,omitempty"`
	Data   any    `json:"data"` // No validation on game data
}

// HandleTurn handles POST and GET on /rooms/{id}/turn.
// POST: takes a turn (only works if player number is correct, no validation on data).
// GET: returns current turn and turns history.
func (s *Server) HandleTurn(w http.ResponseWriter, r *http.Request, roomID string) {
	rInstance, err := s.Store.GetRoom(roomID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		since := 0
		if q := r.URL.Query().Get("since"); q != "" {
			since, _ = strconv.Atoi(q)
		}
		turns, err := rInstance.GetTurns(since)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		_, _, currentTurn, _, _, _, _, _ := rInstance.Snapshot()
		writeJSON(w, http.StatusOK, map[string]any{
			"roomId":      rInstance.ID,
			"currentTurn": currentTurn,
			"turns":       turns,
		})

	case http.MethodPost:
		var req TurnRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		playerNum := req.Player
		// Allow token-based player resolution if player number was not explicitly passed
		if playerNum == 0 && req.Token != "" {
			if resolved, ok := rInstance.GetPlayerFromToken(req.Token); ok {
				playerNum = resolved
			}
		}

		turn, err := rInstance.TakeTurn(playerNum, req.Data)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		_, _, nextTurn, _, _, _, _, _ := rInstance.Snapshot()

		writeJSON(w, http.StatusOK, map[string]any{
			"success":  true,
			"turn":     turn,
			"nextTurn": nextTurn,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MessageRequest defines the payload for sending a message.
type MessageRequest struct {
	Player  int    `json:"player"`
	Token   string `json:"token,omitempty"`
	Message any    `json:"message"`
}

// HandleMessages handles GET and POST on /rooms/{id}/messages.
func (s *Server) HandleMessages(w http.ResponseWriter, r *http.Request, roomID string) {
	rInstance, err := s.Store.GetRoom(roomID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		since := 0
		if q := r.URL.Query().Get("since"); q != "" {
			since, _ = strconv.Atoi(q)
		}

		msgs, err := rInstance.GetMessages(since)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"roomId":   rInstance.ID,
			"messages": msgs,
		})

	case http.MethodPost:
		var req MessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		playerNum := req.Player
		if playerNum == 0 && req.Token != "" {
			if resolved, ok := rInstance.GetPlayerFromToken(req.Token); ok {
				playerNum = resolved
			}
		}

		msg, err := rInstance.AddMessage(playerNum, req.Message)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, msg)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleRoomDetails handles GET /rooms/{id}.
func (s *Server) HandleRoomDetails(w http.ResponseWriter, r *http.Request, roomID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rInstance, err := s.Store.GetRoom(roomID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	id, count, currentTurn, createdAt, expiresAt, turnsCount, msgCount, openSlots := rInstance.Snapshot()
	turns, _ := rInstance.GetTurns(0)

	writeJSON(w, http.StatusOK, map[string]any{
		"roomId":        id,
		"playerCount":   count,
		"openSlots":     openSlots,
		"currentTurn":   currentTurn,
		"turnsCount":    turnsCount,
		"messagesCount": msgCount,
		"turns":         turns,
		"createdAt":     createdAt,
		"expiresAt":     expiresAt,
	})
}

// HandleRoomSlots handles GET /rooms/{id}/slots.
func (s *Server) HandleRoomSlots(w http.ResponseWriter, r *http.Request, roomID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rInstance, err := s.Store.GetRoom(roomID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	openSlots := rInstance.OpenSlots()
	lastActivity := rInstance.LastActivityTime()
	secondsSinceLastMove := int(time.Since(lastActivity).Seconds())
	slotOpenedByInactivity := rInstance.SlotOpenedByInactivity()

	writeJSON(w, http.StatusOK, map[string]any{
		"roomId":                 rInstance.ID,
		"openSlots":              openSlots,
		"currentTurn":            rInstance.CurrentTurn,
		"lastActivityAt":         lastActivity,
		"secondsSinceLastMove":   secondsSinceLastMove,
		"slotOpenedByInactivity": slotOpenedByInactivity,
	})
}

// HandleRoomDispatch dispatches requests under /rooms/{id}/*
func (s *Server) HandleRoomDispatch(w http.ResponseWriter, r *http.Request) {
	// Strip prefix /rooms/ or /api/rooms/
	path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	path = strings.TrimPrefix(path, "/rooms/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	roomID := parts[0]

	if len(parts) == 1 {
		s.HandleRoomDetails(w, r, roomID)
		return
	}

	switch parts[1] {
	case "slots":
		s.HandleRoomSlots(w, r, roomID)
	case "join":
		s.HandleJoinRoom(w, r, roomID)
	case "turn":
		s.HandleTurn(w, r, roomID)
	case "messages":
		s.HandleMessages(w, r, roomID)
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
