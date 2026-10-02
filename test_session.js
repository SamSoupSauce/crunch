const assert = require('assert');

// Mock localStorage for Node test environment
const mockStorage = new Map();
global.localStorage = {
    getItem: (key) => mockStorage.get(key) || null,
    setItem: (key, val) => mockStorage.set(key, String(val)),
    removeItem: (key) => mockStorage.delete(key),
    clear: () => mockStorage.clear()
};

const { persistSession, clearSession, resolvePlayerRole } = require('./session.js');

console.log('Running session.js test suite...');

// Test 1: persistSession and clearSession
mockStorage.clear();
persistSession('ROOM-123', 'player_1');
assert.strictEqual(
    localStorage.getItem('ROOM-123'),
    JSON.stringify({ roomCode: 'ROOM-123', role: 'player_1' }),
    'Test 1 failed: session not persisted correctly'
);

clearSession('ROOM-123');
assert.strictEqual(
    localStorage.getItem('ROOM-123'),
    null,
    'Test 1 failed: session not cleared'
);
console.log('✅ Test 1 Passed: persistSession and clearSession');

// Test 2: Seamless re-attachment when stored as player_1
mockStorage.clear();
persistSession('ROOM-456', 'player_1');
// Even if both slots are taken on server, client storage re-attaches as player_1
const roleP1 = resolvePlayerRole('ROOM-456', { openSlots: [] });
assert.strictEqual(roleP1, 'player_1', 'Test 2 failed: expected player_1 re-attachment');
console.log('✅ Test 2 Passed: Seamless re-attachment as Player 1');

// Test 3: Seamless re-attachment when stored as player_2
mockStorage.clear();
persistSession('ROOM-789', 'player_2');
const roleP2 = resolvePlayerRole('ROOM-789', { openSlots: [] });
assert.strictEqual(roleP2, 'player_2', 'Test 3 failed: expected player_2 re-attachment');
console.log('✅ Test 3 Passed: Seamless re-attachment as Player 2');

// Test 4: Vacant Player 1 (openSlots: [1, 2] or [1]) -> player_1
mockStorage.clear();
const roleVacantP1Both = resolvePlayerRole('ROOM-NEW', { openSlots: [1, 2] });
assert.strictEqual(roleVacantP1Both, 'player_1', 'Test 4 failed: expected player_1 when both vacant');

const roleVacantP1Only = resolvePlayerRole('ROOM-NEW', { openSlots: [1] });
assert.strictEqual(roleVacantP1Only, 'player_1', 'Test 4 failed: expected player_1 when slot 1 vacant');
console.log('✅ Test 4 Passed: Assign player_1 when slot 1 is vacant');

// Test 5: Player 1 taken and Player 2 vacant (openSlots: [2]) -> player_2
mockStorage.clear();
const roleVacantP2Only = resolvePlayerRole('ROOM-JOIN', { openSlots: [2] });
assert.strictEqual(roleVacantP2Only, 'player_2', 'Test 5 failed: expected player_2 when slot 2 vacant');
console.log('✅ Test 5 Passed: Assign player_2 when slot 1 taken and slot 2 vacant');

// Test 6: Both slots claimed (openSlots: []) -> spectator
mockStorage.clear();
const roleFull = resolvePlayerRole('ROOM-FULL', { openSlots: [] });
assert.strictEqual(roleFull, 'spectator', 'Test 6 failed: expected spectator when all slots claimed');
console.log('✅ Test 6 Passed: Assign spectator when both slots claimed');

// Test 7: Handles alternative currentRoomState format (playerCount)
mockStorage.clear();
assert.strictEqual(resolvePlayerRole('ROOM-C0', { playerCount: 0 }), 'player_1');
assert.strictEqual(resolvePlayerRole('ROOM-C1', { playerCount: 1 }), 'player_2');
assert.strictEqual(resolvePlayerRole('ROOM-C2', { playerCount: 2 }), 'spectator');
console.log('✅ Test 7 Passed: Compatible with playerCount room state');

// Test 8: Bare-bones error resilience (null / undefined / missing state)
mockStorage.clear();
assert.strictEqual(resolvePlayerRole(null, null), 'spectator');
assert.strictEqual(resolvePlayerRole('ROOM-ERR', null), 'spectator');
persistSession(null, null);
clearSession(null);
console.log('✅ Test 8 Passed: Bare-bones resilience with missing inputs');

console.log('🎉 All session.js tests passed successfully!');
