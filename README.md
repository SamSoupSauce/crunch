# 💥 CRUNCH (Rock Chess) — Open Protocol Specification

An open-source, zero-sum tactical board game built around physical force limits, 3-axis spatial control, and psychological deception.

Playable in any modern web browser or on a physical hex grid with stones and paper.

---

## 📜 Protocol Rules & Axioms

1. **3-Axis Hex Matrix:** Played on a hexagonal grid with 6 cardinal directions (60° vectors).
2. **Universal Rock Ownership:** Rock colors are purely cosmetic/psychological. **Either player can push ANY rock on the board.**
3. **Physical Force Limit (1 Unit = 1 Rock):** A player unit can only push a single rock. Immovable multi-rock lines act as static barriers.
4. **Follow-Through Movement:** Pushing a rock advances **both** the rock and your player unit 1 hex forward along the chosen vector.
5. **The Crunch Victory Condition:** Pushing a rock directly into an enemy unit who has a boundary wall or static rock behind them traps them instantly → **CRUNCH ACHIEVED**.

---

## ⚙️ Engine Features

- **Infinite Board Radius Scaling (R):** Type any radius size to generate dynamic hex grids.
- **Outer Buffer Corridor (R-1):** Guarantees proper clearance for pushing rocks outward from the starting perimeter ring (R-2).
- **Surgical Handicap Draft Phase:** Players taking a rock handicap enter a manual draft mode to custom-position their reduced rock pool along the perimeter.
- **Zero Dependencies:** Pure HTML5, SVG, Tailwind CSS (CDN), and Web Audio API synthesizer. Single-file architecture.

---

## ⛓️ On-Chain Protocol & Vector NFTs

Because Crunch is completely deterministic and zero-sum, every match can be represented as an immutable, reproducible sequence of state transformations.

### 1. Deterministic Seeding & Replays
A full match history can be encoded into a compact byte payload:
`[Radius R] + [P1/P2 Handicaps] + [Draft Coordinates] + [Move Vector String]`

Any engine adhering to the Crunch Open Protocol can ingest this payload and replay the entire match step-by-step with 100% fidelity.

### 2. Animated SVG NFTs
By storing the move vector payload on-chain, the entire match can be minted as a dynamic NFT. 
- **Zero External Dependencies:** The NFT metadata directly generates a self-contained, looping SVG animation.
- **Visual Replay:** The rendered SVG plays back the complete match vector-by-vector, displaying the opening draft, long-range positioning, and the final lethal Crunch collision.
- **Permanent Record:** A match isn't just recorded in text; its spatial evolution becomes a permanent piece of interactive generative art.

---

## 📄 License

Licensed under the permissive **MIT License**. Feel free to fork, hack, host, manufacture physical sets, or deploy on-chain smart contracts.
