package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var ChromaClient chroma.Client

func main() {

	ctx := context.Background()

	chromaEndpoint, ok := os.LookupEnv("CHROMA_ENDPOINT")
	if !ok {
		chromaEndpoint = "http://localhost:8000"
	}

	ChromaClient, err := chroma.NewHTTPClient(
		chroma.WithBaseURL(chromaEndpoint),
	)

	if err != nil {
		panic("Could not connect to ChromaDB")
	}

	fmt.Printf("Successfully connect to ChromaDB: %s\n", ChromaClient)

	collection, err := ChromaClient.GetOrCreateCollection(ctx, "test")
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		se.Router.POST("/api/embed", func(e *core.RequestEvent) error {

			type CreateTextAsEmbeddingRequest struct {
				Text string `json:"text"`
			}
			data := CreateTextAsEmbeddingRequest{}

			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			err := collection.Add(
				context.Background(),
				chroma.WithIDGenerator(chroma.NewULIDGenerator()),
				chroma.WithTexts(data.Text),
			)
			if err != nil {
				return e.InternalServerError("Failed to create embedding. Cringe", err)
			}

			return e.JSON(http.StatusOK, map[string]any{
				"message": "Saved to ChromaDB",
			})
		},
		)

		se.Router.POST("/api/similar", func(e *core.RequestEvent) error {
			type SimilarSearchRequest struct {
				Text  string `json:"text"`
				Limit int    `json:"limit"`
			}
			data := SimilarSearchRequest{}

			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}
			qr, err := collection.Query(context.Background(), chroma.WithQueryTexts(data.Text), chroma.WithNResults(data.Limit))
			if err != nil {
				return e.InternalServerError("Error querying collection: %s\n", err)
			}
			fmt.Println(qr.GetDocumentsGroups())

			return e.JSON(http.StatusOK, map[string]any{
				"similar_content": qr.GetDocumentsGroups(),
			})

		})

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
