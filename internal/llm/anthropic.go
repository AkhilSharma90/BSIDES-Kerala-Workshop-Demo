package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Anthropic is a thin client for the Messages API. It is intentionally
// minimal — the tool only needs single-turn-style chat with a system
// prompt, no streaming, no tool use, no thinking budget.
type Anthropic struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewAnthropic(apiKey, model string) *Anthropic {
	return &Anthropic{apiKey: apiKey, model: model, http: http.DefaultClient}
}

func (a *Anthropic) Name() string { return "anthropic:" + a.model }

func (a *Anthropic) Chat(ctx context.Context, msgs []Message) (string, error) {
	if a.apiKey == "" {
		return "", errMissing{provider: "anthropic"}
	}

	var system string
	var conv []map[string]string
	for _, m := range msgs {
		if m.Role == "system" {
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
			continue
		}
		conv = append(conv, map[string]string{"role": m.Role, "content": m.Content})
	}

	body := map[string]any{
		"model":      a.model,
		"max_tokens": 1024,
		"messages":   conv,
	}
	if system != "" {
		body["system"] = system
	}

	buf, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	for _, c := range out.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic: no text in response")
}
