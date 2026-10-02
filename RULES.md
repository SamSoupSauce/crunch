# 💥 CRUNCH — Official Rules

## 1. Overview & Objective

**Crunch** is a two-player, turn-based abstract strategy game played on a hexagonal grid. The core concept revolves around maneuvering and weaponizing neutral obstacles to trap the enemy between *"A Rock and a Hard Place"*.

**Objective:** Win the game by forcing the opponent's mover into a lethal trap—pushing a rock into them when they cannot retreat—thereby executing a **Crunch**, or by forcing the opponent to surrender.

---

## 2. Components & Board Setup

### The Board
- **Hexagonal Matrix:** Standard board has a radius of $R = 5$ (side length of 5 hexagons), containing **61 hexagonal spaces**.
- **Six Vectors:** Movement occurs along the 6 cardinal axes ($60^\circ$ increments) of the hexagonal grid.
- **Outer Edge:** The outer boundary constitutes a solid wall. Objects and movers cannot leave the board.

### Pieces
- **Movers:** 2 player pieces:
  - **Player 1 (Blue)**
  - **Player 2 (Red)**
- **Rocks:** 16 neutral rocks (8 starting on each player's side in standard setup).
  - *Universal Ownership:* Rock colors are positional/cosmetic. Once play begins, **either player can push any rock** on the board.

---

## 3. Turn Structure

1. **First Move:** Blue (Player 1) always takes the first turn.
2. **Turn Flow:** Players alternate turns. Each player performs exactly **one action** per turn.
3. **No Passing:** A player must make a legal move on their turn.
4. **Game End:** Play continues until a mover is crushed (**Crunch**) or a player surrenders.

---

## 4. Movement Mechanics

A player controls only their respective mover. On their turn, a player may perform one of three actions:

### A. Simple Move
- A mover may move **one space** into any adjacent empty hex.

### B. Pushing a Rock
- A mover can push an adjacent rock by moving into the rock's hex, displacing the rock **one space forward** in the same line of direction (*follow-through movement*).
- **Single-Piece Force Limit:** A mover can only push **one rock at a time**. Two or more rocks in a row cannot be pushed.
- **Valid Destination:** A rock can only be pushed into an **empty legal hex**. A rock cannot be pushed off the edge of the board or into another rock.
- **Universal Manipulation:** Any rock on the board can be pushed by either mover.

### C. Pushing the Opposing Mover
- A mover can push the opposing mover if the space directly behind the opponent (along the same vector) is an **empty hex**.
- Both movers advance one space in that direction.
- A mover **cannot** push the opposing mover if the space behind them is blocked by a rock or the board edge (unless pushing a rock into them to execute a Crunch).

---

## 5. Hazardous Conditions & Winning (The "Crunch")

A player achieves **Crunch** and immediately wins the match when they push a rock into the opposing mover under **Hazardous Conditions**.

### Hazardous Conditions Definition
Hazardous Conditions occur when:
1. Mover A pushes a rock into Mover B's current hex, **AND**
2. The space directly behind Mover B (along the vector of the push) is blocked by either:
   - A **wall** (the board boundary / edge), OR
   - Another **rock**.

When both conditions are met, Mover B is trapped between the incoming rock and a hard place, resulting in an immediate **Crunch** victory for Mover A.

---

## 6. End of Game

The game concludes immediately when:
- **Crunch Achieved:** A player successfully crushes the opponent under Hazardous Conditions.
- **Surrender:** A player chooses to resign at any point during their turn.
