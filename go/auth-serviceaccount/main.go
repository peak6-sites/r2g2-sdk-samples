package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// This example demonstrates how to authenticate using a Service Account and retrieve an access token that can be used
// to call the API.
// In this example, the service account credentials are stored in a JSON file that was previously retrieved by calling
// the `r2g2 iam service-account create` command. That JSON file is provided to the application using the `R2G2_CREDENTIALS`
// environment variable.
func main() {
	creds, err := loadCredentials()
	if err != nil {
		panic(err)
	}

	// The Go Oauth2 library implements the Client Credentials flow for you.
	// This will automatically retrieve an access token using the provided credentials and refresh it when necessary.
	oauthConfig := clientcredentials.Config{
		ClientID:     creds.ClientId,
		ClientSecret: creds.ClientSecret,
		TokenURL:     creds.TokenUri,
		AuthStyle:    oauth2.AuthStyleInParams,
		EndpointParams: url.Values{
			"audience": {creds.Audience},
		},
	}
	client := oauthConfig.Client(context.TODO())

	// List the first page of Assistants just to validate that the access token works.
	if err := listAssistants(client); err != nil {
		panic(err)
	}
}

type Credentials struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
	TokenUri     string `json:"token_uri"`
}

func loadCredentials() (Credentials, error) {
	bs, err := os.ReadFile(os.Getenv("R2G2_CREDENTIALS"))
	if err != nil {
		return Credentials{}, fmt.Errorf("ReadFile: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(bs, &creds); err != nil {
		return Credentials{}, fmt.Errorf("Unmarshal: %w", err)
	}
	return creds, nil
}

// listAssistants retrieves the first page of Assistants from the API and prints their IDs.
func listAssistants(client *http.Client) error {
	resp, err := client.Post("https://api-proxy-prod.prod.gcp.minisme.ai/ai.assistants.v0.Assistants/ListAssistants",
		"application/json", strings.NewReader("{}"))
	if err != nil {
		return fmt.Errorf("Post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var result struct {
		Assistants []struct {
			Id   string `json:"id"`
			Name string `json:"displayName"`
		} `json:"assistants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("Decode: %w", err)
	}
	for _, assistant := range result.Assistants {
		fmt.Println(assistant.Name, assistant.Id)
	}
	return nil
}
