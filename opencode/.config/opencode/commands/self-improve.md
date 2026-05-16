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

## Config Quality Analysis (NEW)

In addition to conversation analysis, also evaluate the **configuration itself** for gaps:

### Agent Coverage Check
Check if all agents referenced in `AGENTS.md` and `devops.md` DELEGATION sections actually exist.
Missing agents should be created.

| Referenced Agent | Exists? |
|-----------------|---------|
| `@terraform` | agents/terraform.md |
| `@security` | agents/security.md |
| `@cicd` | agents/cicd.md |
| `@backend` | agents/backend.md |
| `@frontend` | agents/frontend.md |
| `@data-engineer` | agents/data-engineer.md |
| `@ansible` | agents/ansible.md? |
| `@python-dev` | agents/python-dev.md? |

### Skill Gap Analysis
Compare existing skills against the agent prompts. If an agent mentions a domain knowledge area
that has no corresponding skill — that's a gap.

| Domain | Referenced In | Skill Exists? |
|--------|--------------|--------------|
| Git workflow | AGENTS.md, devops agent | skills/git-workflow/ |
| Linux admin | Dotfiles context | skills/linux-admin/ |
| Secrets/vault | security agent, AGENTS.md | skills/vault-secrets/ |

### Command Coverage Check
For every agent, check if there's a command that helps invoke that agent's core workflow:

| Agent | Core Workflow Command | Exists? |
|-------|----------------------|---------|
| terraform | tf-plan, tf-apply | ✅ |
| security | sec-audit | ✅ |
| cicd | pipeline-lint | ✅ |
| devops | docker-build, k8s-check | ✅ |
| architect | infra-review | ✅ |
| data-engineer | cost-estimate | ✅ |

### Configuration Consistency Check
- Do `opencode.json` and `tui.json` reference the same instructions files?
- Are all skill directories properly structured (SKILL.md + optional assets)?
- Are agent frontmatter YAML `permission` blocks consistent with their role?

## Rules
- Do NOT remove existing instructions that are working correctly
- APPEND new rules, don't replace existing ones
- Be specific — no vague rules like "be better". Write exact, actionable instructions
- If a skill doesn't exist for the topic, CREATE a new one
- Every change must reference a specific conversation moment that triggered it
- For config quality findings (not conversation-triggered), note them as: "CONFIG GAP: [description]"
