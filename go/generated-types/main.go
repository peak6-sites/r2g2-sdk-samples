package main

import (
	"context"
	"crypto/tls"
	assistantspb "github.com/peak6-sites/r2g2-apis/gen/go/ai/assistants/v0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"os"
)

func main() {
	conn, err := grpc.NewClient("api-proxy-prod.prod.gcp.minisme.ai:443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		panic(err)
	}
	client := assistantspb.NewAssistantsClient(conn)
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+os.Getenv("R2G2_TOKEN"))
	a, err := client.GetAssistant(ctx, &assistantspb.GetAssistantRequest{
		Id: "default",
	})
	if err != nil {
		panic(err)
	}
	println(a.String())
}
