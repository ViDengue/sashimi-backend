package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sashimi/backend/routes"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaOpenai "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

var ChromaClient chroma.Client

func main() {
	llm, err := openai.New(openai.WithModel("gpt-4.1-nano"))

	if err != nil {
		log.Fatalf("Error connecting to OpenAI: %s\n", err)
		return
	}

	ef, efErr := chromaOpenai.NewOpenAIEmbeddingFunction(os.Getenv("OPENAI_API_KEY"), chromaOpenai.WithModel(chromaOpenai.TextEmbedding3Large))

	if efErr != nil {
		fmt.Printf("Error creating OpenAI embedding function: %s \n", efErr)
	}

	ctx := context.Background()

	chromaEndpoint, ok := os.LookupEnv("CHROMA_ENDPOINT")
	if !ok {
		chromaEndpoint = "http://localhost:8000"
	}

	ChromaClient, err := chroma.NewHTTPClient(
		chroma.WithBaseURL(chromaEndpoint),
	)

	collection, err := ChromaClient.GetOrCreateCollection(ctx, "test", chroma.WithEmbeddingFunctionCreate(ef))

	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	app := pocketbase.New()

	registerRoutes(app, llm, collection)

	if err := app.Start(); err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		err := ChromaClient.Close()
		if err != nil {
			log.Fatalf("Error closing ChromaDB Client: %s\n", err)
		}
	}()
}

func registerRoutes(app *pocketbase.PocketBase, llm llms.Model, collection chroma.Collection) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		routes.RegisterRoutes(app, se, &llm, collection)

		return se.Next()
	})

}
