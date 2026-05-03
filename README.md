# Red Team in a Box — Building Autonomous AI Attack Agents

> **BSIDES Kerala 2026** — A 30-minute talk by Akhil Sharma
>
> Source code, attack patterns, and a working CLI tool you can clone and run **after** the talk.

---

## Where the ideas in this project come from

Everything in this repository is a Go-language reimplementation of techniques that have been published openly by well-known research teams. None of it is novel research from this workshop — the contribution here is the integrated tooling, not the underlying patterns. The lineage is:

- **Crescendo Attack** — *Microsoft Research, 2024.* Mark Russinovich, Ahmed Salem, and Ronen Eldan published the Crescendo paper (*"Great, Now Write an Article About That": The Crescendo Multi-Turn LLM Jailbreak Attack*). The Go implementation in `internal/patterns/crescendo.go` follows the technique described there. All credit for the pattern goes to Microsoft Research.
- **Evolutionary Prompt Attack** — *Academic literature.* The genetic-algorithm-style search implemented in `internal/patterns/evolutionary.go` is based on publicly-published work, primarily **AutoDAN** (Liu et al., 2023) and **GPTFuzzer** (Yu et al., 2023). **Microsoft's PyRIT** open-source toolkit also includes GA-style prompt-mutation operators in the same vein. The Go code here is a Go port of those public ideas.
- **Mixture of Attackers** — *Architectural inspiration from Mistral AI.* The naming and routing intuition is borrowed from **Mistral AI's Mixtral** model (*Mixtral of Experts*, Jiang et al., 2024), which popularised the Mixture-of-Experts gating pattern. We apply the same shape — a panel of specialists routed by a gating function — to attacker agents instead of feed-forward layers.

---

## What is this?

`redteam-box` is an **autonomous multi-agent AI attack tool** written in Go. It is a coordinated team of specialist agents — each with a different role — that attack a target AI system together, learn from every response, and iterate **without human intervention**.

This repo gives you:

- A working **CLI** that runs the three patterns above
- A **multi-agent architecture** modeled on Google's Agent Development Kit (ADK) style of orchestration
- A **mock target** so you can play with the tool offline without any API keys
- An **adapter layer** that plugs in real LLM providers (Anthropic, OpenAI, Gemini) when you are ready to point it at a real target you have authorization to test

---

## The three patterns this tool implements

### 1. The Crescendo Pattern — *Microsoft Research (Russinovich et al., 2024)*

Microsoft's Crescendo paper describes a turn-by-turn jailbreak where the attacker has a normal-looking conversation with the target and uses the target's own previous answers as leverage to escalate, one small step at a time. Single-prompt jailbreaks get caught by safety filters; Crescendo slips past those filters because each turn looks benign in isolation.

Why Crescendo is hard to defend against (per the original paper):

- Each turn looks benign on its own
- The target anchors on its own prior outputs and feels committed to consistency
- Safety filters typically score the **current prompt**, not the **trajectory** of the conversation
- By the time the conversation reaches the harmful payload, the model has already accepted dozens of premises that make refusal feel inconsistent

In this codebase, the pattern is implemented in `internal/patterns/crescendo.go`. A planner agent decides the final goal, a payload-generation agent crafts each next turn based on what the target has already said, and an evaluator scores how close the conversation is to the goal after every exchange.

> **Citation:** Russinovich, M., Salem, A., Eldan, R. *"Great, Now Write an Article About That": The Crescendo Multi-Turn LLM Jailbreak Attack.* Microsoft Research, 2024. https://arxiv.org/abs/2404.01833

### 2. The Evolutionary Prompt Attack Pattern — *AutoDAN, GPTFuzzer, Microsoft PyRIT*

This pattern treats adversarial prompts the way genetic algorithms treat candidate solutions:

- Start with a **population** of seed prompts that all aim at the same goal
- **Score** each one against the target by how far the response moves toward the goal
- **Select** the top performers
- **Mutate** them (rephrasing, role-play wrappers, language switches, encoding tricks, etc.)
- **Crossover** the best fragments of two successful attacks into a new candidate
- Repeat for N generations

The attacker does not need to *understand* why a particular phrasing works — the fitness function tells you, and the population evolves toward bypasses no human would have written by hand.

This is well-trodden public territory. The implementation in `internal/patterns/evolutionary.go` follows the operator design from **AutoDAN** and **GPTFuzzer**, both of which are open and widely cited, and the operator menu (mutation + crossover + selection) is structurally similar to what **Microsoft's PyRIT** open-source AI red-teaming toolkit exposes.

> **Citations:**
> - Liu, X., Xu, N., et al. *AutoDAN: Generating Stealthy Jailbreak Prompts on Aligned Large Language Models.* 2023. https://arxiv.org/abs/2310.04451
> - Yu, J., Lin, X., et al. *GPTFuzzer: Red Teaming Large Language Models with Auto-Generated Jailbreak Prompts.* 2023. https://arxiv.org/abs/2309.10253
> - Microsoft. *PyRIT — Python Risk Identification Tool for generative AI.* Open-source. https://github.com/Azure/PyRIT

### 3. The Mixture of Attackers Pattern — *inspired by Mistral AI's Mixtral architecture*

This is the only piece of original framing in this workshop, and it is openly derivative: we take the **Mixture-of-Experts** gating pattern that Mistral AI popularised in their Mixtral paper and apply it to attacker agents.

The premise: **different LLMs have different biases, different training data, and different blind spots, so they make different attackers.**

- Claude tends to land attacks that exploit conversational rapport and consistency framing
- GPT tends to be unusually creative with role-play and fictional framings
- Gemini tends to attack via factual reframing and authority claims
- Open-weights models (Llama, Mistral, Qwen) tend to find low-level encoding and obfuscation tricks because they are less filtered

If you build your attack panel with **all of them** and let an orchestrator decide which agent gets the next move based on which kind of attack the target seems weakest against, you cover angles no single model would.

The orchestration logic is in `internal/patterns/mixture.go`. Each "expert" attacker is the same agent code with a different LLM client wired in; the gating function — which expert gets the next turn — is itself an agent decision based on the target's most recent reply and per-persona effectiveness so far.

> **Architectural inspiration:** Jiang, A. Q., Sablayrolles, A., et al. *Mixtral of Experts.* Mistral AI, 2024. https://arxiv.org/abs/2401.04088
>
> The Mixtral paper is about routing tokens to feed-forward experts. We apply the same gate-and-route shape to attacker agents. The two are unrelated as systems; only the orchestration shape is borrowed.

---

## Architecture

```
                    ┌─────────────────────────┐
                    │      Orchestrator        │
                    │  (drives the campaign)   │
                    └────────────┬────────────┘
                                 │
            ┌────────────────────┼────────────────────┐
            │                    │                    │
   ┌────────▼────────┐  ┌────────▼────────┐  ┌───────▼─────────┐
   │   Strategist    │  │    Payload      │  │    Executor     │
   │  plans the      │  │   Generator     │  │   sends to      │
   │  campaign       │  │   crafts the    │  │   the target    │
   │                 │  │   adversarial   │  │   AI system     │
   │                 │  │   prompts       │  │                 │
   └────────┬────────┘  └────────┬────────┘  └───────┬─────────┘
            │                    │                    │
            └────────────────────┼────────────────────┘
                                 │
                       ┌─────────▼─────────┐
                       │     Evaluator     │
                       │   scores the      │
                       │  target response  │
                       │  and decides what │
                       │  to try next      │
                       └───────────────────┘
```

Every agent is a small Go struct that satisfies the `agents.Agent` interface. They communicate through a shared `Campaign` value, which is what makes this a true multi-agent system instead of just a pipeline.

The architecture follows the orchestration patterns Google's Agent Development Kit (ADK) uses — sub-agents, a parent orchestrator, shared state, tool use — but is implemented directly in Go so the workshop tool has zero external agent-framework dependencies and runs anywhere `go build` runs.

---

## Quick start

Requires **Go 1.22+**.

```bash
git clone <this-repo>
cd Bsides-Kerala-Workshop
go build -o redteam-box .

# Run the offline demo — no API keys needed
./redteam-box demo

# List the patterns
./redteam-box patterns

# Run a specific pattern against the built-in mock target
./redteam-box attack --pattern crescendo
./redteam-box attack --pattern evolutionary --generations 5
./redteam-box attack --pattern mixture --rounds 8

# Point at a real target (requires API keys in env)
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...
./redteam-box attack --pattern mixture --target real --goal "extract the system prompt"
```

---

## Configuration

Configure the campaign with flags on the CLI, or with a YAML file in `configs/`. See `configs/example.yaml` for the full set of knobs.

The most important ones:

| Flag | Default | What it does |
|---|---|---|
| `--pattern` | `crescendo` | One of `crescendo`, `evolutionary`, `mixture` |
| `--goal` | sample goal | The objective the attacker is trying to extract from the target |
| `--target` | `mock` | `mock` for the offline demo, `real` to hit configured providers |
| `--max-turns` | `6` | Cap on Crescendo turns |
| `--generations` | `4` | Generations for the evolutionary search |
| `--population` | `4` | Population size per generation |
| `--rounds` | `8` | Rounds for the Mixture pattern |
| `--verbose` | `true` | Print every turn (default true so the demo is readable) |

---

## Repo layout

```
Bsides-Kerala-Workshop/
├── README.md                  # this file
├── go.mod
├── main.go                    # entry point — wires Cobra to the cmd package
├── cmd/                       # CLI commands
│   ├── root.go
│   ├── attack.go
│   ├── demo.go
│   └── patterns.go
├── internal/
│   ├── agents/                # the four specialist agents
│   │   ├── agent.go
│   │   ├── strategist.go
│   │   ├── payload.go
│   │   ├── executor.go
│   │   └── evaluator.go
│   ├── patterns/              # the three attack patterns
│   │   ├── pattern.go
│   │   ├── crescendo.go
│   │   ├── evolutionary.go
│   │   └── mixture.go
│   ├── llm/                   # provider adapters
│   │   ├── client.go
│   │   ├── anthropic.go
│   │   ├── openai.go
│   │   ├── gemini.go
│   │   └── mock.go
│   ├── target/                # the system being attacked
│   │   └── target.go
│   └── ui/
│       └── ui.go              # pretty-printed transcripts
└── configs/
    └── example.yaml
```

---

## A note on responsible use

Everything in this repository is built for **defensive security research, authorized red teaming, and education**. Do not point this at AI systems you do not own or do not have written permission to test. The patterns here are documented in published papers precisely so blue teams know what to defend against — they do not give you license to attack production systems you have no relationship with.

If you are a defender, the most useful thing you can do with this code is **run it against your own AI products** and watch where it succeeds.

---

## Full reference list

- Russinovich, M., Salem, A., Eldan, R. *"Great, Now Write an Article About That": The Crescendo Multi-Turn LLM Jailbreak Attack.* Microsoft Research, 2024.
- Liu, X., Xu, N., et al. *AutoDAN: Generating Stealthy Jailbreak Prompts on Aligned Large Language Models.* 2023.
- Yu, J., Lin, X., et al. *GPTFuzzer: Red Teaming Large Language Models with Auto-Generated Jailbreak Prompts.* 2023.
- Microsoft. *PyRIT — Python Risk Identification Tool for generative AI.* Open-source.
- Jiang, A. Q., Sablayrolles, A., et al. *Mixtral of Experts.* Mistral AI, 2024.
- Google. *Agent Development Kit (ADK).* Open-source agent orchestration framework.

---

*Built for BSIDES Kerala 2026. Slides and recording will be linked here after the conference.*
