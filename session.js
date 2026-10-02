/**
 * Client-side player assignment and session recovery logic.
 * Minimal, stateless, dependency-free vanilla JavaScript.
 */

/**
 * Persists active room association in localStorage keyed by roomCode.
 * @param {string} roomCode - The target room identifier.
 * @param {"player_1" | "player_2" | "spectator"} role - The player role.
 */
function persistSession(roomCode, role) {
    if (!roomCode || !role) return;
    try {
        localStorage.setItem(roomCode, JSON.stringify({ roomCode, role }));
    } catch (e) {
        // Bare-bones, silent fallback if localStorage is unavailable or quota exceeded
    }
}

/**
 * Clears the active room association from localStorage.
 * @param {string} roomCode - The target room identifier.
 */
function clearSession(roomCode) {
    if (!roomCode) return;
    try {
        localStorage.removeItem(roomCode);
    } catch (e) {
        // Bare-bones, silent fallback
    }
}

/**
 * Resolves player role based on client localStorage and current room state.
 *
 * Rules:
 * 1. If stored locally as a role, re-attach seamlessly with that role.
 * 2. If no role in storage and Player 1 is vacant, assign/prompt Player 1.
 * 3. If Player 1 is taken and Player 2 is vacant, assign Player 2.
 * 4. If both slots are claimed, assign spectator/observer.
 *
 * @param {string} roomCode - The target room identifier.
 * @param {Object} [currentRoomState] - Current room details from server.
 * @returns {"player_1" | "player_2" | "spectator"} Resolved role.
 */
function resolvePlayerRole(roomCode, currentRoomState) {
    // 1. Check local client storage for existing session
    if (roomCode) {
        try {
            const raw = localStorage.getItem(roomCode);
            if (raw) {
                const session = JSON.parse(raw);
                if (session && session.role) {
                    return session.role;
                }
            }
        } catch (e) {}
    }

    if (!currentRoomState) {
        return 'spectator';
    }

    // Determine slot vacancies from current room state
    let p1Vacant = false;
    let p2Vacant = false;

    if (Array.isArray(currentRoomState.openSlots)) {
        p1Vacant = currentRoomState.openSlots.includes(1);
        p2Vacant = currentRoomState.openSlots.includes(2);
    } else if (typeof currentRoomState.playerCount === 'number') {
        p1Vacant = currentRoomState.playerCount === 0;
        p2Vacant = currentRoomState.playerCount === 1;
    } else {
        p1Vacant = Boolean(currentRoomState.p1Vacant ?? !currentRoomState.player1);
        p2Vacant = Boolean(currentRoomState.p2Vacant ?? !currentRoomState.player2);
    }

    // 2. If Player 1 is vacant -> Player 1
    if (p1Vacant) {
        return 'player_1';
    }

    // 3. If Player 1 is taken and Player 2 is vacant -> Player 2
    if (p2Vacant) {
        return 'player_2';
    }

    // 4. If both slots are claimed -> Spectator
    return 'spectator';
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        persistSession,
        clearSession,
        resolvePlayerRole
    };
}
