package room

import (
	"sync"
	"testing"
	"time"
)

func TestRoomLifecycle(t *testing.T) {
	store := NewStore()
	r := store.CreateRoom()

	// Initial checks
	if r.ExpiresAt.Sub(r.CreatedAt) != 24*time.Hour {
		t.Errorf("expected 24h expiration, got %v", r.ExpiresAt.Sub(r.CreatedAt))
	}
	if r.PlayerCount != 0 {
		t.Errorf("expected 0 players initially, got %d", r.PlayerCount)
	}

	// Player 1 joins
	p1, token1, err := r.Join()
	if err != nil || p1 != 1 || token1 == "" {
		t.Fatalf("failed p1 join: p=%d, err=%v", p1, err)
	}

	// Player 2 joins
	p2, token2, err := r.Join()
	if err != nil || p2 != 2 || token2 == "" {
		t.Fatalf("failed p2 join: p=%d, err=%v", p2, err)
	}

	// 2 players joined: current turn must be 1 (first to join goes first)
	if r.CurrentTurn != 1 {
		t.Errorf("expected current turn 1, got %d", r.CurrentTurn)
	}

	// Player 3 joins -> rejected
	_, _, err = r.Join()
	if err != ErrNoEmptySlots {
		t.Errorf("expected ErrNoEmptySlots, got %v", err)
	}

	// Turn: Player 2 tries out of turn -> rejected
	_, err = r.TakeTurn(2, "move_data")
	if err == nil {
		t.Fatalf("expected error when player 2 moves out of turn")
	}

	// Turn: Player 1 takes turn -> succeeds
	turn1, err := r.TakeTurn(1, map[string]int{"x": 10, "y": 20})
	if err != nil {
		t.Fatalf("failed taking turn 1: %v", err)
	}
	if turn1.TurnNumber != 1 || turn1.Player != 1 {
		t.Errorf("unexpected turn 1 output: %+v", turn1)
	}

	// Current turn should now be 2
	if r.CurrentTurn != 2 {
		t.Errorf("expected current turn 2, got %d", r.CurrentTurn)
	}

	// Player 2 takes turn
	turn2, err := r.TakeTurn(2, "any arbitrary data")
	if err != nil {
		t.Fatalf("failed taking turn 2: %v", err)
	}
	if turn2.TurnNumber != 2 || turn2.Player != 2 {
		t.Errorf("unexpected turn 2 output: %+v", turn2)
	}

	// Messages
	msg1, err := r.AddMessage(1, "hello from 1")
	if err != nil || msg1.ID != 1 {
		t.Fatalf("failed adding message: %v", err)
	}
	msg2, err := r.AddMessage(2, "hello from 2")
	if err != nil || msg2.ID != 2 {
		t.Fatalf("failed adding message: %v", err)
	}

	allMsgs, err := r.GetMessages(0)
	if err != nil || len(allMsgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(allMsgs))
	}
}

func TestConcurrentTurns(t *testing.T) {
	store := NewStore()
	r := store.CreateRoom()
	_, _, _ = r.Join()
	_, _, _ = r.Join()

	var wg sync.WaitGroup
	// Attempt concurrent turns
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = r.TakeTurn(1, "t1")
		}()
		go func() {
			defer wg.Done()
			_, _ = r.TakeTurn(2, "t2")
		}()
	}
	wg.Wait()

	// Should not panic, and turns list must match alternating players
	for i, turn := range r.Turns {
		expectedPlayer := (i % 2) + 1
		if turn.Player != expectedPlayer {
			t.Errorf("turn %d: expected player %d, got %d", i, expectedPlayer, turn.Player)
		}
	}
}

func TestPickingOpenSlots(t *testing.T) {
	store := NewStore()
	r := store.CreateRoom()

	// 1. Host is always player 1: unhosted room only has slot 1 open
	open := r.OpenSlots()
	if len(open) != 1 || open[0] != 1 {
		t.Fatalf("expected [1] open slot initially (host is always player 1), got %v", open)
	}

	// Trying to join slot 2 first fails because host must be player 1
	_, _, err := r.JoinSlot(2)
	if err != ErrSlotTaken {
		t.Fatalf("expected ErrSlotTaken when attempting to join slot 2 before player 1, got %v", err)
	}

	// Host joins (passes 1 or 0)
	p, token, err := r.JoinSlot(1)
	if err != nil || p != 1 || token == "" {
		t.Fatalf("expected successful join to slot 1, got p=%d, err=%v", p, err)
	}

	// 2. Player 2 is the only one if player 1 has joined
	open = r.OpenSlots()
	if len(open) != 1 || open[0] != 2 {
		t.Fatalf("expected only [2] open slot after player 1 joined, got %v", open)
	}

	// Trying to join slot 1 again fails
	_, _, err = r.JoinSlot(1)
	if err != ErrSlotTaken {
		t.Fatalf("expected ErrSlotTaken when attempting to join slot 1 again, got %v", err)
	}

	// Player 2 joins
	p2, token2, err := r.JoinSlot(2)
	if err != nil || p2 != 2 || token2 == "" {
		t.Fatalf("expected successful join to slot 2, got p=%d, err=%v", p2, err)
	}

	// Both slots are taken -> OpenSlots is empty
	open = r.OpenSlots()
	if len(open) != 0 {
		t.Fatalf("expected 0 open slots, got %v", open)
	}

	// Game starts, player 1 goes first
	if r.CurrentTurn != 1 {
		t.Fatalf("expected current turn 1, got %d", r.CurrentTurn)
	}

	// Third player tries to join -> ErrNoEmptySlots
	_, _, err = r.JoinSlot(0)
	if err != ErrNoEmptySlots {
		t.Fatalf("expected ErrNoEmptySlots, got %v", err)
	}
}

func TestInactivitySlotReopening(t *testing.T) {
	store := NewStore()
	r := store.CreateRoom()

	_, _, err := r.JoinSlot(1)
	if err != nil {
		t.Fatalf("join 1 failed: %v", err)
	}
	_, _, err = r.JoinSlot(2)
	if err != nil {
		t.Fatalf("join 2 failed: %v", err)
	}

	// Initially active: no open slots
	if len(r.OpenSlots()) != 0 {
		t.Fatalf("expected 0 open slots right after game start")
	}
	if r.SlotOpenedByInactivity() {
		t.Fatalf("expected SlotOpenedByInactivity() == false")
	}

	// Simulate 6 minutes of inactivity while waiting for Player 1's turn
	r.LastMoveAt = time.Now().Add(-6 * time.Minute)

	// Since >5 minutes have passed, slot 1 should reopen
	if !r.SlotOpenedByInactivity() {
		t.Fatalf("expected SlotOpenedByInactivity() == true after 6m inactivity")
	}
	open := r.OpenSlots()
	if len(open) != 1 || open[0] != 1 {
		t.Fatalf("expected [1] open slot due to player 1 inactivity, got %v", open)
	}

	// New player takes over slot 1
	p, newToken, err := r.JoinSlot(1)
	if err != nil || p != 1 {
		t.Fatalf("failed taking over inactive slot 1: %v", err)
	}
	if pResolved, ok := r.GetPlayerFromToken(newToken); !ok || pResolved != 1 {
		t.Fatalf("new token should resolve to player 1, got %d, %v", pResolved, ok)
	}

	// Slot is now filled again
	if len(r.OpenSlots()) != 0 {
		t.Fatalf("expected 0 open slots after takeover")
	}

	// Player 1 takes turn
	_, err = r.TakeTurn(1, "move1")
	if err != nil {
		t.Fatalf("expected player 1 to be able to take turn, got %v", err)
	}
	if r.CurrentTurn != 2 {
		t.Fatalf("expected current turn to advance to 2, got %d", r.CurrentTurn)
	}

	// Simulate 6 minutes of inactivity while waiting for Player 2's turn
	r.LastMoveAt = time.Now().Add(-6 * time.Minute)
	open = r.OpenSlots()
	if len(open) != 1 || open[0] != 2 {
		t.Fatalf("expected [2] open slot due to player 2 inactivity, got %v", open)
	}
}
