package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type Client struct {
	Model     string       
	BaseURL   string       
	HTTP      *http.Client 
	Normalize bool         
}

func New(model string) *Client {
	return &Client{
		Model:     model,
		BaseURL:   "http://localhost:11434",
		HTTP:      &http.Client{Timeout: 60 * time.Second},
		Normalize: true,
	}
}

type reqBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type respBody struct {
	EmbeddingF32 []float32 `json:"embedding"`
	EmbeddingF64 []float64 `json:"-"`
}

func (r *respBody) UnmarshalJSON(b []byte) error {
	var f32 struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(b, &f32); err == nil && len(f32.Embedding) > 0 {
		r.EmbeddingF32 = f32.Embedding
		return nil
	}
	var f64 struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(b, &f64); err == nil && len(f64.Embedding) > 0 {
		r.EmbeddingF64 = f64.Embedding
		return nil
	}
	var any struct {
		Embedding any `json:"embedding"`
	}
	if err := json.Unmarshal(b, &any); err == nil && any.Embedding == nil {
		return nil
	}
	return fmt.Errorf("invalid embedding payload")
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	body, _ := json.Marshal(reqBody{Model: c.Model, Prompt: text})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embeddings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama status %d: %s", res.StatusCode, string(raw))
	}

	var out respBody
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	var v []float32
	switch {
	case len(out.EmbeddingF32) > 0:
		v = make([]float32, len(out.EmbeddingF32))
		copy(v, out.EmbeddingF32)
	case len(out.EmbeddingF64) > 0:
		v = make([]float32, len(out.EmbeddingF64))
		for i, x := range out.EmbeddingF64 {
			v[i] = float32(x)
		}
	default:
		return nil, fmt.Errorf("empty embedding")
	}

	if c.Normalize {
		l2norm(v)
	}
	return v, nil
}

func Cosine(a, b []float32) float32 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var s float32
	for i := 0; i < n; i++ {
		s += a[i] * b[i]
	}
	return s
}

func l2norm(v []float32) {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	if s == 0 {
		return
	}
	inv := 1 / float32(math.Sqrt(s))
	for i := range v {
		v[i] *= inv
	}
}
