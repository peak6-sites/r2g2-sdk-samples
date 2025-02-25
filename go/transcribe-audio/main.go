package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const transcribeUrl = "https://api-proxy-prod.prod.gcp.minisme.ai/ai.audio.v0.Transcriber/Transcribe"

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("file open error:", err)
		return
	}

	text, err := transcribe(context.Background(), f)
	if err != nil {
		fmt.Println("transcribe error:", err)
		return
	}
	fmt.Println("Result:")
	fmt.Println(text)
}

func transcribe(ctx context.Context, f *os.File) (string, error) {
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("file read error: %w", err)
	}

	js, err := json.Marshal(map[string]interface{}{
		"inline_data": map[string]interface{}{
			"mime_type": "audio/mpeg",
			"data":      base64.StdEncoding.EncodeToString(data),
		},
	})
	if err != nil {
		return "", fmt.Errorf("request marshal error: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, transcribeUrl, bytes.NewReader(js))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("R2G2_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	body := make(map[string]string)
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("response unmarshal error: %w", err)
	}
	return body["text"], nil
}
