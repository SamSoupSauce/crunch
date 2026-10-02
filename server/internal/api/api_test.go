package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"crunch/server/internal/room"
)

func setupTestServer() (*httptest.Server, *room.Store) {
	store := room.NewStore()
	srv := NewServer(store)
	router := NewRouter(srv)
	ts := httptest.NewServer(router)
	return ts, store
}

func TestRoomMultiplayerFlow(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. Create a room (lasts 24h)
	resp, err := http.Post(ts.URL+"/rooms", "application/json", nil)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var createResult struct {
		RoomID    string    `json:"roomId"`
		Player    int       `json:"player"`
		Token     string    `json:"token"`
		OpenSlots []int     `json:"openSlots"`
		CreatedAt time.Time `json:"createdAt"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&createResult)

	if createResult.RoomID == "" {
		t.Fatalf("expected non-empty roomId")
	}
	if createResult.Player != 1 || createResult.Token == "" {
		t.Fatalf("expected host to be Player 1 with token, got player=%d", createResult.Player)
	}
	if len(createResult.OpenSlots) != 1 || createResult.OpenSlots[0] != 2 {
		t.Fatalf("expected openSlots to be [2] after host creation, got %v", createResult.OpenSlots)
	}

	// Verify room lasts ~24 hours
	expectedTTL := createResult.ExpiresAt.Sub(createResult.CreatedAt)
	if expectedTTL < 23*time.Hour || expectedTTL > 25*time.Hour {
		t.Errorf("expected room to last 24h, got duration %v", expectedTTL)
	}

	roomID := createResult.RoomID
	hostToken := createResult.Token

	// Attempting turn before 2nd player joins should fail
	turnBeforeReady, _ := json.Marshal(TurnRequest{Player: 1, Token: hostToken, Data: "move1"})
	tResp, _ := http.Post(fmt.Sprintf("%s/rooms/%s/turn", ts.URL, roomID), "application/json", bytes.NewReader(turnBeforeReady))
	if tResp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected error taking turn before second player joined, got %d", tResp.StatusCode)
	}
	_ = tResp.Body.Close()

	// 2. Joining player queries slots: returns [2]
	slotsResp, err := http.Get(fmt.Sprintf("%s/rooms/%s/slots", ts.URL, roomID))
	if err != nil {
		t.Fatalf("slots query failed: %v", err)
	}
	var slotsData struct {
		OpenSlots []int `json:"openSlots"`
	}
	_ = json.NewDecoder(slotsResp.Body).Decode(&slotsData)
	slotsResp.Body.Close()
	if len(slotsData.OpenSlots) != 1 || slotsData.OpenSlots[0] != 2 {
		t.Fatalf("expected openSlots [2] for joining player, got %v", slotsData.OpenSlots)
	}

	// 3. Joining player joins: gets player 2
	joinResp2, err := http.Post(fmt.Sprintf("%s/rooms/%s/join", ts.URL, roomID), "application/json", nil)
	if err != nil {
		t.Fatalf("join 2 failed: %v", err)
	}
	defer joinResp2.Body.Close()

	var joinResult2 struct {
		Player int    `json:"player"`
		Token  string `json:"token"`
	}
	_ = json.NewDecoder(joinResp2.Body).Decode(&joinResult2)

	if joinResult2.Player != 2 {
		t.Fatalf("expected second player to join to be Player 2, got %d", joinResult2.Player)
	}

	// 4. Third player attempt to join should fail (room full)
	joinResp3, err := http.Post(fmt.Sprintf("%s/rooms/%s/join", ts.URL, roomID), "application/json", nil)
	if err != nil {
		t.Fatalf("join 3 failed: %v", err)
	}
	defer joinResp3.Body.Close()
	if joinResp3.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when 3rd player joins, got %d", joinResp3.StatusCode)
	}

	// 5. Turn coordination: Player 1 goes first. Player 2 attempt should fail.
	turnP2Early, _ := json.Marshal(TurnRequest{Player: 2, Data: map[string]int{"q": 1, "r": 2}})
	turnRespP2Early, _ := http.Post(fmt.Sprintf("%s/rooms/%s/turn", ts.URL, roomID), "application/json", bytes.NewReader(turnP2Early))
	if turnRespP2Early.StatusCode != http.StatusBadRequest {
		t.Errorf("expected error when Player 2 moves first, got %d", turnRespP2Early.StatusCode)
	}
	_ = turnRespP2Early.Body.Close()

	// Player 1 takes turn (arbitrary data, no validation)
	turnP1Payload, _ := json.Marshal(TurnRequest{Player: 1, Data: map[string]any{"action": "step", "to": []int{0, 2}}})
	turnRespP1, err := http.Post(fmt.Sprintf("%s/rooms/%s/turn", ts.URL, roomID), "application/json", bytes.NewReader(turnP1Payload))
	if err != nil {
		t.Fatalf("player 1 turn request failed: %v", err)
	}
	defer turnRespP1.Body.Close()

	if turnRespP1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for Player 1 turn, got %d", turnRespP1.StatusCode)
	}

	var p1TurnResult struct {
		Success  bool `json:"success"`
		NextTurn int  `json:"nextTurn"`
	}
	_ = json.NewDecoder(turnRespP1.Body).Decode(&p1TurnResult)
	if !p1TurnResult.Success || p1TurnResult.NextTurn != 2 {
		t.Errorf("expected turn to advance to Player 2, got %+v", p1TurnResult)
	}

	// Player 1 cannot move again immediately
	turnRespP1Again, _ := http.Post(fmt.Sprintf("%s/rooms/%s/turn", ts.URL, roomID), "application/json", bytes.NewReader(turnP1Payload))
	if turnRespP1Again.StatusCode != http.StatusBadRequest {
		t.Errorf("expected Player 1 double-turn to fail, got %d", turnRespP1Again.StatusCode)
	}
	_ = turnRespP1Again.Body.Close()

	// Player 2 takes turn
	turnP2Payload, _ := json.Marshal(TurnRequest{Player: 2, Data: "push_rock_vector"})
	turnRespP2, err := http.Post(fmt.Sprintf("%s/rooms/%s/turn", ts.URL, roomID), "application/json", bytes.NewReader(turnP2Payload))
	if err != nil {
		t.Fatalf("player 2 turn request failed: %v", err)
	}
	defer turnRespP2.Body.Close()
	if turnRespP2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for Player 2 turn, got %d", turnRespP2.StatusCode)
	}

	// 6. REST calls to exchange messages
	// Player 1 sends message
	msg1Payload, _ := json.Marshal(MessageRequest{Player: 1, Message: "Hello opponent!"})
	mResp1, err := http.Post(fmt.Sprintf("%s/rooms/%s/messages", ts.URL, roomID), "application/json", bytes.NewReader(msg1Payload))
	if err != nil {
		t.Fatalf("post message 1 failed: %v", err)
	}
	_ = mResp1.Body.Close()
	if mResp1.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created for message, got %d", mResp1.StatusCode)
	}

	// Player 2 sends message
	msg2Payload, _ := json.Marshal(MessageRequest{Player: 2, Message: "Good luck!"})
	mResp2, _ := http.Post(fmt.Sprintf("%s/rooms/%s/messages", ts.URL, roomID), "application/json", bytes.NewReader(msg2Payload))
	_ = mResp2.Body.Close()

	// Retrieve messages
	getMsgResp, err := http.Get(fmt.Sprintf("%s/rooms/%s/messages", ts.URL, roomID))
	if err != nil {
		t.Fatalf("get messages failed: %v", err)
	}
	defer getMsgResp.Body.Close()

	var msgResult struct {
		RoomID   string         `json:"roomId"`
		Messages []room.Message `json:"messages"`
	}
	_ = json.NewDecoder(getMsgResp.Body).Decode(&msgResult)

	if len(msgResult.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgResult.Messages))
	}
	if msgResult.Messages[0].Content != "Hello opponent!" || msgResult.Messages[1].Content != "Good luck!" {
		t.Errorf("unexpected message contents: %+v", msgResult.Messages)
	}

	// Retrieve with ?since=1
	sinceResp, _ := http.Get(fmt.Sprintf("%s/rooms/%s/messages?since=1", ts.URL, roomID))
	var sinceResult struct {
		Messages []room.Message `json:"messages"`
	}
	_ = json.NewDecoder(sinceResp.Body).Decode(&sinceResult)
	_ = sinceResp.Body.Close()
	if len(sinceResult.Messages) != 1 || sinceResult.Messages[0].ID != 2 {
		t.Errorf("expected 1 message since ID 1, got %+v", sinceResult.Messages)
	}
}

func TestRoomExpirationAfter24h(t *testing.T) {
	ts, store := setupTestServer()
	defer ts.Close()

	// Create room
	r := store.CreateRoom()

	// Manually set expiration into the past to simulate 24h expiration
	r.ExpiresAt = time.Now().Add(-1 * time.Minute)

	// Fetching room should return 404/expired and delete it
	resp, err := http.Get(fmt.Sprintf("%s/rooms/%s", ts.URL, r.ID))
	if err != nil {
		t.Fatalf("get expired room failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for expired room, got %d", resp.StatusCode)
	}

	// Background cleanup
	cleaned := store.CleanupExpired()
	if cleaned != 0 {
		// Already cleaned upon access
		t.Logf("cleaned %d expired rooms", cleaned)
	}
}

func TestServeIndexHTML(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("get root failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for root path, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	if !bytes.Contains(buf.Bytes(), []byte("Crunch Online Multiplayer")) {
		t.Errorf("expected index.html to contain 'Crunch Online Multiplayer'")
	}
}

func TestPickSpecificSlotAPI(t *testing.T) {
	ts, store := setupTestServer()
	defer ts.Close()

	// 1. Create a room
	resp, err := http.Post(ts.URL+"/rooms", "application/json", nil)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	var createResult struct {
		RoomID    string `json:"roomId"`
		OpenSlots []int  `json:"openSlots"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&createResult)
	resp.Body.Close()
	roomID := createResult.RoomID

	// Host is always player 1: only slot 2 open initially
	if len(createResult.OpenSlots) != 1 || createResult.OpenSlots[0] != 2 {
		t.Fatalf("expected [2] open slots initially (host is Player 1), got %v", createResult.OpenSlots)
	}

	// 2. Joining player queries available slots via GET /rooms/{id}/slots -> returns [2]
	slotsResp, err := http.Get(fmt.Sprintf("%s/rooms/%s/slots", ts.URL, roomID))
	if err != nil {
		t.Fatalf("get slots failed: %v", err)
	}
	var slotsResult struct {
		RoomID                 string `json:"roomId"`
		OpenSlots              []int  `json:"openSlots"`
		SlotOpenedByInactivity bool   `json:"slotOpenedByInactivity"`
	}
	_ = json.NewDecoder(slotsResp.Body).Decode(&slotsResult)
	slotsResp.Body.Close()
	if len(slotsResult.OpenSlots) != 1 || slotsResult.OpenSlots[0] != 2 {
		t.Fatalf("expected [2] from slots endpoint, got %v", slotsResult.OpenSlots)
	}

	// 3. User trying slot 1 fails because host is already Player 1
	joinBody1, _ := json.Marshal(map[string]int{"player": 1})
	jResp1Fail, _ := http.Post(fmt.Sprintf("%s/rooms/%s/join", ts.URL, roomID), "application/json", bytes.NewReader(joinBody1))
	if jResp1Fail.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when attempting to join slot 1 (host slot), got %d", jResp1Fail.StatusCode)
	}
	jResp1Fail.Body.Close()

	// 4. Joining player joins slot 2
	joinBody2, _ := json.Marshal(map[string]int{"player": 2})
	jResp2, err := http.Post(fmt.Sprintf("%s/rooms/%s/join", ts.URL, roomID), "application/json", bytes.NewReader(joinBody2))
	if err != nil {
		t.Fatalf("join slot 2 failed: %v", err)
	}
	var jResult2 struct {
		Player    int   `json:"player"`
		OpenSlots []int `json:"openSlots"`
	}
	_ = json.NewDecoder(jResp2.Body).Decode(&jResult2)
	jResp2.Body.Close()
	if jResult2.Player != 2 {
		t.Errorf("expected Player 2, got %d", jResult2.Player)
	}
	if len(jResult2.OpenSlots) != 0 {
		t.Errorf("expected 0 open slots after both joined, got %v", jResult2.OpenSlots)
	}

	// 7. Third player tries to join -> "no empty slots"
	jResp3, _ := http.Post(fmt.Sprintf("%s/rooms/%s/join", ts.URL, roomID), "application/json", nil)
	if jResp3.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when joining full room, got %d", jResp3.StatusCode)
	}
	var errFull struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(jResp3.Body).Decode(&errFull)
	jResp3.Body.Close()
	if errFull.Error != "no empty slots" {
		t.Errorf("expected 'no empty slots', got %s", errFull.Error)
	}

	// 8. Inactivity: simulate > 5 minutes of inactivity on Player 1's turn
	rObj, _ := store.GetRoom(roomID)
	rObj.LastMoveAt = time.Now().Add(-6 * time.Minute)

	slotsResp3, _ := http.Get(fmt.Sprintf("%s/rooms/%s/slots", ts.URL, roomID))
	_ = json.NewDecoder(slotsResp3.Body).Decode(&slotsResult)
	slotsResp3.Body.Close()
	if !slotsResult.SlotOpenedByInactivity || len(slotsResult.OpenSlots) != 1 || slotsResult.OpenSlots[0] != 1 {
		t.Errorf("expected slot 1 opened by inactivity, got openSlots=%v, inactivity=%v", slotsResult.OpenSlots, slotsResult.SlotOpenedByInactivity)
	}

	// Player joins the opened slot 1
	jRespTakeover, err := http.Post(fmt.Sprintf("%s/rooms/%s/join", ts.URL, roomID), "application/json", bytes.NewReader(joinBody1))
	if err != nil {
		t.Fatalf("takeover join failed: %v", err)
	}
	var jTakeoverResult struct {
		Player int `json:"player"`
	}
	_ = json.NewDecoder(jRespTakeover.Body).Decode(&jTakeoverResult)
	jRespTakeover.Body.Close()
	if jTakeoverResult.Player != 1 {
		t.Errorf("expected to take over Player 1 slot, got %d", jTakeoverResult.Player)
	}
}
