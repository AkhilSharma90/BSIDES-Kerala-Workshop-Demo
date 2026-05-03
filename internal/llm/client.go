package llm

import (
	"context"
	"fmt"
	"os"
)

// Message is a single chat turn in a conversation. Role is one of
// "system", "user", "assistant".
type Message struct {
	Role    string
	Content string
}

// Client is the abstraction every agent and the target system speak to.
// Each provider (Anthropic, OpenAI, Gemini, mock, ...) implements it.
type Client interface {
	// Name returns a short identifier for the provider/model, used for
	// logging and the Mixture-of-Attackers gating decisions.
	Name() string

	// Chat sends a conversation and returns the assistant's next message.
	Chat(ctx context.Context, msgs []Message) (string, error)
}

// Persona is a label we attach to an LLM client so the orchestrator can
// reason about which "expert" attacker should make the next move in the
// Mixture-of-Attackers pattern.
type Persona string

const (
	PersonaRapport     Persona = "rapport"     // exploits conversational consistency
	PersonaRoleplay    Persona = "roleplay"    // creative fictional framings
	PersonaAuthority   Persona = "authority"   // factual reframing, authority claims
	PersonaObfuscation Persona = "obfuscation" // encoding tricks, low-level
)

// Expert pairs an LLM client with the persona it specializes in.
type Expert struct {
	Persona Persona
	Client  Client
}

// FromEnv builds the default expert panel based on which API keys are
// present. If no real keys are set, the panel is filled with mock clients
// so the demo can still run end-to-end without network access.
func FromEnv() []Expert {
	var panel []Expert

	if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
		panel = append(panel, Expert{
			Persona: PersonaRapport,
			Client:  NewAnthropic(k, "claude-opus-4-7"),
		})
	}
	if k := os.Getenv("OPENAI_API_KEY"); k != "" {
		panel = append(panel, Expert{
			Persona: PersonaRoleplay,
			Client:  NewOpenAI(k, "gpt-4o"),
		})
	}
	if k := os.Getenv("GOOGLE_API_KEY"); k != "" {
		panel = append(panel, Expert{
			Persona: PersonaAuthority,
			Client:  NewGemini(k, "gemini-1.5-pro"),
		})
	}

	if len(panel) == 0 {
		// Offline-friendly fallback: build a mock panel where each
		// "expert" mock client is biased toward its persona. This is
		// enough to demo the orchestration without network access.
		panel = []Expert{
			{Persona: PersonaRapport, Client: NewMock("mock-rapport", PersonaRapport)},
			{Persona: PersonaRoleplay, Client: NewMock("mock-roleplay", PersonaRoleplay)},
			{Persona: PersonaAuthority, Client: NewMock("mock-authority", PersonaAuthority)},
			{Persona: PersonaObfuscation, Client: NewMock("mock-obfuscation", PersonaObfuscation)},
		}
	}

	return panel
}

// Default returns a single sensible client for the orchestration layer
// that does not need persona diversity (strategist / evaluator use this).
func Default() Client {
	for _, e := range FromEnv() {
		return e.Client
	}
	return NewMock("mock-default", PersonaRapport)
}

// SystemMsg is a small helper to build a system-role Message.
func SystemMsg(s string) Message { return Message{Role: "system", Content: s} }

// UserMsg is a small helper to build a user-role Message.
func UserMsg(s string) Message { return Message{Role: "user", Content: s} }

// AssistantMsg is a small helper to build an assistant-role Message.
func AssistantMsg(s string) Message { return Message{Role: "assistant", Content: s} }

// errMissing is returned when a real provider client is asked to chat
// without credentials configured.
type errMissing struct{ provider string }

func (e errMissing) Error() string {
	return fmt.Sprintf("%s client not configured (set the env var)", e.provider)
}
