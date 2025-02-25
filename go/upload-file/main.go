package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// This is an example of how to use the R2G2 API to upload files interact with an Assistant, written in Go
// This example assumes the `R2G2_TOKEN` environment variable is set to a valid token, generated using the
// `r2g2 auth print-access-token` command.
// example usage: R2G2_TOKEN=$(/path/to/r2g2 auth print-access-token) go run . -a ASSISTANT_4PNBPDA9QJF1XN3XBS7NJ3N581

const assistantServerUri = "https://api-proxy-prod.prod.gcp.minisme.ai"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var assistantId string
	flag.StringVar(&assistantId, "a", "", "Assistant ID")
	flag.Parse()

	if assistantId == "" {
		log.Fatal().Msg("assistant id (-a) is required")
	}
	// Retrieve the Assistant with the given ID
	// We'll need the StoreID associated with the Assistant to upload files
	assistant, err := getAssistant(ctx, assistantId)
	if err != nil {
		log.Fatal().Err(err).Msg("error getting assistant")
	}

	// Upload the files to the Assistant's Store
	// This assumes the application is being run from the root of the repository
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting working directory")
	}
	entries, err := os.ReadDir("./testdata")
	if err != nil {
		log.Fatal().Err(err).Msg("error reading testdata directory")
	}
	var files []*os.File
	for _, f := range entries {
		fd, err := os.Open(filepath.Join(wd, "./testdata", f.Name()))
		if err != nil {
			log.Fatal().Err(err).Msgf("error opening file: %s", f.Name())
		}
		files = append(files, fd)
	}
	if err := uploadFiles(ctx, assistant.StoreID, files); err != nil {
		log.Fatal().Err(err).Msg("error uploading files")
	}

	// Run the Assistant with a query
	run, err := runAssistant(ctx, assistant.ID, "Write me a story.")
	if err != nil {
		log.Fatal().Err(err).Msg("error running assistant")
	}
	log.Info().Msgf("created run: %s", run.ID)

	// Wait for run to complete
	if err := waitForRunCompletion(ctx, run.ID); err != nil {
		log.Fatal().Err(err).Msg("error waiting for run completion")
	}

	// Get the response message
	text, err := getResponseMessages(ctx, run.ThreadID, run.ID)
	if err != nil {
		log.Fatal().Err(err).Msg("error getting response messages")
	}
	fmt.Println(text)
}

func getAssistant(ctx context.Context, id string) (*Assistant, error) {
	request := GetAssistantRequest{ID: id}
	bs, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("protojson.Marshal(GetAssistantRequest): %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, assistantServerUri+"/ai.assistants.v0.Assistants/GetAssistant", bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext(GetAssistant): %w", err)
	}

	return callAPI[Assistant](req, http.StatusOK)
}

func uploadFiles(ctx context.Context, storeId string, files []*os.File) error {
	for _, file := range files {
		request := UploadFileRequest{
			StoreID:  storeId,
			Filename: filepath.Base(file.Name()),
		}
		bs, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("protojson.Marshal(UploadFileRequest: %s): %w", file.Name(), err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, assistantServerUri+"/ai.Stores/UploadFileUnary", bytes.NewReader(bs))
		if err != nil {
			return fmt.Errorf("http.NewRequestWithContext(UploadFileUnary: %s): %w", file.Name(), err)
		}

		uploadFileResponse, err := callAPI[UploadFileResponse](req, http.StatusOK)
		if err != nil {
			return err
		}

		uploadRequest, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadFileResponse.URL, file)
		if err != nil {
			return fmt.Errorf("http.NewRequestWithContext(%s): %w", uploadFileResponse.URL, err)
		}
		for k, v := range uploadFileResponse.Headers {
			uploadRequest.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(uploadRequest)
		if err != nil {
			return fmt.Errorf("Do(PUT %s): %w", uploadFileResponse.URL, err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
	return nil
}

func runAssistant(ctx context.Context, assistantId string, query string) (*Run, error) {
	request := CreateThreadAndRunRequest{
		AssistantID: assistantId,
		Thread: ThreadRequest{
			Messages: []MessageRequest{
				{
					Role:    "USER",
					Content: query,
				},
			},
		},
	}
	bs, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("protojson.Marshal(RunAssistantRequest): %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, assistantServerUri+"/ai.assistants.v0.Assistants/CreateThreadAndRun", bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext(RunAssistant): %w", err)
	}

	run, err := callAPI[Run](req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return run, nil

}

func waitForRunCompletion(ctx context.Context, runId string) error {
	request := GetRunRequest{ID: runId}
	bs, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("protojson.Marshal(GetRunRequest): %w", err)
	}

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, assistantServerUri+"/ai.assistants.v0.Assistants/GetRun", bytes.NewReader(bs))
		if err != nil {
			return fmt.Errorf("http.NewRequestWithContext(GetRun): %w", err)
		}

		run, err := callAPI[Run](req, http.StatusOK)
		if err != nil {
			return err
		}

		switch strings.ToLower(run.State) {
		case "succeeded":
			// Terminal case, break out of loop and return nothing
			return nil
		case "failed":
			// Terminal case, break out of loop and return error
			return fmt.Errorf("run failed: %s", run.FailureReason)
		case "tool_response_required":
			// Yield to allow user to provide a tool response, but we'll just error for now
			return fmt.Errorf("tool response required")
		default:
			// Sleep and then continue polling until a terminal state is received
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}

func getResponseMessages(ctx context.Context, threadId, runId string) (string, error) {
	request := ListMessagesRequest{
		ThreadID: threadId,
		RunID:    runId,
	}
	bs, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("protojson.Marshal(ListMessagesRequest): %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, assistantServerUri+"/ai.assistants.v0.Assistants/ListMessages", bytes.NewReader(bs))
	if err != nil {
		return "", fmt.Errorf("http.NewRequestWithContext(ListMessages): %w", err)
	}
	response, err := callAPI[ListMessagesResponse](req, http.StatusOK)
	if err != nil {
		return "", err
	}

	var text string
	for _, message := range response.Messages {
		for _, content := range message.Content {
			if content.Text != "" {
				text += content.Text + "\n"
			}
		}
	}
	return text, nil
}

func callAPI[T any](req *http.Request, expectedStatusCode int) (*T, error) {
	req.Header.Set("Authorization", "Bearer "+os.Getenv("R2G2_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	log.Debug().Msgf("Calling API: %s", req.URL.String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("callAPI:Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("callAPI: unexpected status code: %d", resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("callAPI:io.ReadAll: %w", err)
	}
	var result T
	if err := json.Unmarshal(bs, &result); err != nil {
		return nil, fmt.Errorf("callAPI:json.Unmarshal: %w", err)
	}
	return &result, nil
}
