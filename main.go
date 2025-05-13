package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	defaultef "github.com/amikos-tech/chroma-go/pkg/embeddings/default_ef"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var ChromaClient chroma.Client
var embedderEndpoint = "http://localhost:5005/embed"

func main() {

	chromaEndpoint, ok := os.LookupEnv("CHROMA_ENDPOINT")
	if !ok {
		chromaEndpoint = "localhost:8000"
	}

	ChromaClient, err := chroma.NewHTTPClient(
		chroma.WithBaseURL(chromaEndpoint),
	)

	if err != nil {
		panic("Could not connect to ChromaDB")
	}

	fmt.Printf("Successfully connect to ChromaDB: %s\n", ChromaClient)
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		se.Router.POST("/api/embed", CreateTextAsEmbedding)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := ChromaClient.Close()
		if err != nil {
			log.Fatalf("Error closing ChromaDB Client: %s\n", err)
		}
	}()
}

type CreateTextAsEmbeddingRequest struct {
	text string `json:"text"`
}

func CreateTextAsEmbedding(e *core.RequestEvent) error {
	data := CreateTextAsEmbeddingRequest{}

	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Failed to read request data", err)
	}

	resp, err := CreateEmbeddingFromText(data.text)
	if err != nil {
		return e.InternalServerError("Failed to create embedding. Cringe", err)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"message":   "Saved to ChromaDB",
		"embedding": resp,
	})
}

func GetSimilarText(e *core.RequestEvent) error {
	return e.JSON(http.StatusOK, map[string]any{})
}

func CreateEmbeddingFromText(text string) ([]embeddings.Embedding, error) {
	ef, closeef, efErr := defaultef.NewDefaultEmbeddingFunction()

	defer func() {
		err := closeef()
		if err != nil {
			fmt.Printf("Error closing default embedding function: %s \n", err)
		}
	}()
	if efErr != nil {
		fmt.Printf("Error creating OpenAI embedding function: %s \n", efErr)
	}

	resp, reqErr := ef.EmbedDocuments(context.Background(), []string{text})
	if reqErr != nil {
		fmt.Printf("Error embedding documents: %s \n", reqErr)
	}

	return resp, nil

}
