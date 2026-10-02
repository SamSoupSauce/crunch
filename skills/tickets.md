```markdown
# SKILL: Ticket-Driven Development Workflow

## Protocol Overview
When assigned a task, your scope of execution is driven by a single ticket file in the `tickets/` directory (e.g., `tickets/042.md`). You must treat this file as your operational contract.

---

## Operating Rules

1. **Targeting & Scope:**
   - Focus exclusively on the requirements, constraints, and acceptance criteria defined in the designated active ticket.
   - Do not scan or read the entire `tickets/` directory by default. Keep context lean.

2. **On-Demand Historical Lookback:**
   - The codebase on disk is always **ground truth**. Historical tickets represent intent history and rationale.
   - If you encounter ambiguous architecture, unfamiliar abstractions, or cross-cutting changes, check the active ticket for reference breadcrumbs (e.g., `Depends on: 038.md`, `Supersedes: 021.md`).
   - You are permitted to inspect specific prior tickets (e.g., `tickets/038.md`) strictly on-demand to understand past decisions. Do not revert code to match outdated tickets if the source tree has evolved.

3. **Immutability of the Specification:**
   - Do NOT modify, delete, or rewrite the user's objective, instructions, or acceptance criteria in the ticket.
   - All agent output within the ticket must be appended under the `## Resolution` heading.

---

## Completion Contract (Resolution Logging)

Before concluding your run, you must update the active ticket file by completing the `## Resolution` block at the bottom of the file using the exact structure below:

```markdown
## Resolution
- **Status:** [Complete | Incomplete | Blocked]
- **Files Touched:**
  - `path/to/file1.ext`
  - `path/to/file2.ext`
- **Key Decisions & Deviations:**
  - Brief summary of architectural choices, trade-offs, or changes made from the original plan.
- **Verification:**
  - Command run: `[test/build/lint command]`
  - Result: `[Pass/Fail summary, exit code, or benchmark metric]`

```

Never consider a ticket finished until the code passes its verification steps and the resolution block is written.

