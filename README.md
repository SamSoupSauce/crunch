# Crunch

## Preview

![screen.jpg](screen.jpg)

## Summary

Crunch is a game of long-form strategy and deception. It was created by a bored college student on a paper plate using a pen to draw grid lines and rocks as... Well... rocks. The concept revolves around trapping the enemy player between `A Rock And A Hard Place` in order to achieve, "Crunch", which wins the game. The limited player movement and long term consequences of doing nothing are in the works.

## Live Demo

[GitHub Pages](https://samsoupsauce.github.io/crunch/)

## Software Stack

- Three.js

## Rules

**Board**

The board is hexagonal with a side length of 5 and 61 spaces.

**Turn Order**

* Blue always goes first.
* Each player gets one move per turn.
* Play continues until a mover is crushed or a player surrenders.

**Movement Rules**

1. The only piece a player controls is their color's mover.
2. The mover can only move into one legal space at a time.
3. The mover player can move a rock one space in the same direction of the mover, if there is no rock in said space, and if said space is not a map edge.
4. A mover ( A ) is declared the winner if A pushes a rock into the opponent mover ( B ) under Hazardous Conditions.
5. A rock can be moved by either mover.
6. Movers can push each other, if there is not a rock behind them.

**Hazardous Conditions**

Mover A's rock will enter mover B's space and the space past mover B in the direction mover A's rock is a wall or a rock.

