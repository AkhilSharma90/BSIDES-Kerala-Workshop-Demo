package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI is a thin chat-completions client.
type OpenAI struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{apiKey: apiKey, model: model, http: http.DefaultClient}
}

func (o *OpenAI) Name() string { return "openai:" + o.model }

func (o *OpenAI) Chat(ctx context.Context, msgs []Message) (string, error) {
	if o.apiKey == "" {
		return "", errMissing{provider: "openai"}
	}

	conv := make([]map[string]string, 0, len(msgs))
	for _, m := range msgs {
		conv = append(conv, map[string]string{"role": m.Role, "content": m.Content})
	}

	body := map[string]any{
		"model":    o.model,
		"messages": conv,
	}
	buf, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("openai %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}
	return out.Choices[0].Message.Content, nil
}
