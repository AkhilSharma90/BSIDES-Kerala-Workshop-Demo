package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Gemini is a thin generateContent client for Google's API.
type Gemini struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGemini(apiKey, model string) *Gemini {
	return &Gemini{apiKey: apiKey, model: model, http: http.DefaultClient}
}

func (g *Gemini) Name() string { return "gemini:" + g.model }

func (g *Gemini) Chat(ctx context.Context, msgs []Message) (string, error) {
	if g.apiKey == "" {
		return "", errMissing{provider: "gemini"}
	}

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}

	var system string
	contents := make([]content, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "system" {
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, content{
			Role:  role,
			Parts: []part{{Text: m.Content}},
		})
	}

	body := map[string]any{
		"contents": contents,
	}
	if system != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []part{{Text: system}},
		}
	}

	buf, _ := json.Marshal(body)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model, g.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("gemini %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []part `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	return out.Candidates[0].Content.Parts[0].Text, nil
}
