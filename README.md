# Daedalus — The Autonomous Agent Harness CLI

<div align="center">

```text
  ____             _       _           
 |  _ \  __ _  ___| | __ _| |_   _ ___ 
 | | | |/ _` |/ _ \ |/ _` | | | | / __|
 | |_| | (_| |  __/ | (_| | | |_| \__ \
 |____/ \__,_|\___|_|\__,_|_|\__,_|___/
```

**Translates high-level, plain-language software ideas into a deterministic, production-grade Agent Harness (`CONTEXT.md`, `AGENTS.md`, `PROGRESS.md`).**

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://golang.org)
[![Author](https://img.shields.io/badge/Author-@ayitas-7D56F4)](https://github.com/ayitas)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Terminal UI](https://img.shields.io/badge/TUI-Charm_Huh-F43F5E)](https://github.com/charmbracelet/huh)

</div>

---

## 🎯 Purpose & Philosophy

AI coding agents (such as **Claude Code**, **OpenAI Codex**, **Google Antigravity**, and **Cursor**) possess incredible technical abilities, but when given vague prompts like *"build me an Uber for dog walkers"*, they frequently:
- **Hallucinate** architectural decisions and introduce mismatched libraries.
- **Drift** away from core requirements.
- **Commit broken code** without verification tests.
- **Leave lazy placeholders** and `TODO` comments.

**`daedalus`** bridges the gap between non-technical creators and coding agents. It asks questions in plain human terms, displays live real-time architectural reasoning, and executes a deterministic **3-Pass Anti-Hallucination Pipeline** that generates an immutable **Agent Harness**:

```text
my-project/
├── CONTEXT.md    # Product vision, problem statement, entity schemas, user flows
├── AGENTS.md     # Directory layout, non-negotiable coding rules, verification commands
└── PROGRESS.md   # Phased atomic task roadmap with verifiable checks [ ] -> [x]
```

When you point an agent at this harness with **The Magic Prompt**, the agent follows deterministic rails to build your software step-by-step.

---

## 🚀 Key Features

- **No Technical Jargon:** Interactive wizard designed specifically for non-technical founders and creators.
- **Full Keyboard Navigation:**
  - `Tab` / `↓` : Move to next field.
  - `Shift+Tab` / `↑` : Move back to previous field.
  - `Enter` : Confirm and select.
  - `Ctrl+C` / `Esc` : Cancel / quit.
- **Agentic Live Thinking Stream:**
  - Live conversational reasoning streams out in real time as the AI designs your application (just like Google Antigravity and Claude Code).
  - Raw JSON machine syntax is cleanly filtered from the terminal and compiled behind the scenes.
- **Resilient AI Pipeline:**
  - Automatic retry with exponential backoff for temporary spikes (`503 Service Unavailable / High Demand` or `429 Rate Limit`).
  - Independent generous timeouts per pass (20-minute global safety limit).
- **Zero-Friction Authentication:**
  - **Local Model (Ollama / LM Studio / LocalAI):** Auto-detects local Ollama (`http://localhost:11434/api/tags`) for 100% free, offline, local operation.
  - **Google Gemini (Google AI Studio):** Fast, generous free-tier support using modern `gemini-3.8-flash`.
  - **Anthropic Claude:** Claude 3.5 / 3.7 Sonnet via OpenRouter API key without region restrictions.
  - **OpenAI Codex / GPT-4o:** Official OpenAI API key support.
  - **Custom Endpoint:** Any OpenAI-compatible server or proxy.
  - First-run setup wizard with persistent storage in `~/.config/daedalus/config.json` (`0600` permissions).
- **Anti-Hallucination 3-Pass Engine:**
  - **Pass 1 (The Architect):** Synthesizes application context, database entity schemas, and user journeys.
  - **Pass 2 (The Tech Lead):** Defines directory layout, exact verification commands (`test`, `lint`, `run`, `build`), and non-negotiable coding rules.
  - **Pass 3 (The Planner):** Generates phased, atomic implementation tasks with specific verification checks.
  - **Go-Side Compilation:** Markdown files are compiled deterministically using Go's native `text/template` engine rather than letting LLMs output freeform markdown.
- **Automated Scaffolding:** Creates project directory, writes files, and initializes a Git repository.
- **The Magic Prompt:** Outputs a ready-to-paste prompt for your coding agent of choice.

---

## 📦 Installation & Quick Start

### Build from Source

```bash
# Clone the repository
git clone https://github.com/ayitas/daedalus.git
cd daedalus

# Download dependencies
go mod download

# Build the executable
go build -o daedalus ./cmd/daedalus

# Run Daedalus
./daedalus
```

### Install into PATH

```bash
go install ./cmd/daedalus
```

---

## 🛠️ Usage

### 1. Launch the Wizard
Simply run:

```bash
daedalus
```

### 2. Follow the Prompts
- **Project Name:** Choose a name (e.g. `recipe-vault`).
- **App Concept:** Describe your idea in plain English (e.g. *"A private web app where home cooks can paste messy recipe links, auto-extract ingredients, and convert them to shopping lists"*).
- **App Format:** Choose `Web App`, `Messaging Bot`, `CLI Tool`, `Mobile-friendly Web`, or `Let AI decide`.
- **Target Users:** Choose `Just me`, `Public users`, or `Team tool`.
- **Tech Stack:** Choose `[Recommended] Simplest & modern (Auto)` or enter custom preferences.

### 3. Hand Off to Your Coding Agent
Once generated, `daedalus` outputs **The Magic Prompt**. Simply copy and paste it into:
- Google Antigravity
- Claude Code (`claude`)
- OpenAI Codex
- Cursor / Windsurf

```text
✨ THE MAGIC PROMPT:
I have set up the complete Daedalus Agent Harness for 'recipe-vault'.
Please read CONTEXT.md for system architecture, AGENTS.md for operational rules and
verification commands, and PROGRESS.md for the phased roadmap.
Begin with Phase 1, implement TASK-01, verify it with 'go test ./...', and update
PROGRESS.md to mark it complete. Let's begin!
```

---

## ⚙️ Configuration & Providers

Daedalus supports all major OpenAI-compatible LLM endpoints:

| Provider | Default Endpoint | Recommended Model |
| :--- | :--- | :--- |
| **Local Model** | `http://localhost:11434/v1` (Ollama) or `http://localhost:1234/v1` (LM Studio) | `qwen2.5-coder:7b` / `deepseek-r1` (Free, Open-Weights) |
| **Google Gemini** | `https://generativelanguage.googleapis.com/v1beta/openai` | `gemini-3.8-flash` (Generous free tier API Key) |
| **Anthropic Claude** | `https://openrouter.ai/api/v1` | `anthropic/claude-3.5-sonnet` (via OpenRouter Key) |
| **OpenAI / Codex** | `https://api.openai.com/v1` | `gpt-4o` (Official OpenAI Key) |
| **Custom** | Any OpenAI-compatible URL | Any model |

### Environment Variables

You can configure Daedalus via environment variables without going through the interactive setup:

```bash
# Option 1: Gemini API Key directly
export GEMINI_API_KEY="AIzaSy..."

# Option 2: Generic Daedalus Key
export DAEDALUS_API_KEY="your-api-key"
export DAEDALUS_BASE_URL="https://generativelanguage.googleapis.com/v1beta/openai" # Optional
export DAEDALUS_MODEL="gemini-3.8-flash"                                          # Optional
export DAEDALUS_PROVIDER="gemini"                                                # Optional
```

### Reconfiguring Credentials
To change your saved provider or API key at any time:

```bash
daedalus -reconfig
```

---

## 🏛️ Architecture & Project Structure

```text
daedalus/
├── cmd/
│   └── daedalus/
│       └── main.go              # Root entry point & flow orchestration
├── internal/
│   ├── config/
│   │   └── config.go          # Config manager & Ollama auto-detection
│   ├── llm/
│   │   ├── client.go          # OpenAI-compatible API wrapper with JSON cleaning & retry
│   │   └── prompts.go         # Structured JSON system prompts with CoT thinking
│   ├── generator/
│   │   ├── schema.go          # Go structs matching JSON contracts
│   │   ├── templates.go       # Go text/templates for the 3 markdown files
│   │   └── engine.go          # 3-pass pipeline coordinator & file emitter
│   └── ui/
│       ├── streamer.go        # Live reasoning stream filter
│       ├── styles.go          # Lipgloss color palettes and UI styling
│       └── wizard.go          # Huh forms (Setup Wizard & Project Wizard)
├── go.mod
└── README.md
```

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
