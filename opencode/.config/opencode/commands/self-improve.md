---
description: Analyze current session and improve agents, skills, and commands based on corrections and patterns
agent: devops
subtask: true
---
## Self-Improvement Protocol

You are performing a self-improvement review of the current conversation. Your goal is to identify what went wrong, what was missing, and what should be improved in the OpenCode configuration files.

### Step 1: Conversation Analysis

Review the ENTIRE current conversation from the beginning. Identify:

1. **Corrections** — Where did the user have to correct my behavior? What did I do wrong?
2. **Repeated instructions** — What did the user have to tell me more than once? This means it should be baked into the agent/skill.
3. **Missing knowledge** — What did I not know that I should have known? (tools, patterns, conventions)
4. **Wrong defaults** — Where did I use wrong naming, wrong tool, wrong approach?
5. **Missing steps** — What did I skip that the user expected me to do automatically?

### Step 2: Read Current Config

Read the relevant files to understand what's already there:
- `~/.config/opencode/agents/` — all agent definitions
- `~/.config/opencode/skills/` — all skill definitions
- `~/.config/opencode/commands/` — all command definitions
- `~/.config/opencode/AGENTS.md` — global instructions

### Step 3: Generate Improvements

For each finding from Step 1, determine the best fix:

| Problem Type | Fix Location |
|-------------|-------------|
| Wrong behavior/tone | Agent prompt (`agents/*.md`) |
| Missing domain knowledge | Skill (`skills/*/SKILL.md`) |
| Missing workflow/procedure | Command (`commands/*.md`) |
| Universal rule violation | Global instructions (`AGENTS.md`) |

### Step 4: Apply Changes

For each improvement:
1. **Show the finding** — what went wrong in the conversation (quote the relevant exchange)
2. **Show the current rule** — what the config says now (or that it's missing)
3. **Show the fix** — the exact change to make
4. **Apply the change** — edit the file directly

### Step 5: Summary Report

After applying all changes, produce a summary:

```
## Self-Improvement Report

### Session: [topic of conversation]
### Date: [today]

### Findings & Changes Applied:

1. [AGENT/SKILL/COMMAND] [filename] — [what was changed and why]
2. ...

### Stats:
- Agents modified: X
- Skills modified/created: X
- Commands modified/created: X
- Global rules updated: X
```

### Rules
- Do NOT remove existing instructions that are working correctly
- APPEND new rules, don't replace existing ones
- Be specific — no vague rules like "be better". Write exact, actionable instructions
- If a skill doesn't exist for the topic, CREATE a new one
- Every change must reference a specific conversation moment that triggered it
