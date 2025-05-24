package routes

import (
	"context"
	"fmt"
	"net/http"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type SimilarSearchRequest struct {
	Text  string `json:"text"`
	Limit int    `json:"limit"`
}

func registerEmbeddingRoutes(_ *pocketbase.PocketBase, se *core.ServeEvent, collection chroma.Collection) {
	se.Router.POST("/api/embed", handleEmbeddingRequest(se, collection))
	se.Router.POST("/api/similar", handleSimilarRequest(se, collection))
}

func handleEmbeddingRequest(_ *core.ServeEvent, collection chroma.Collection) handlerFunc {
	return func(e *core.RequestEvent) error {
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
			return e.InternalServerError(fmt.Sprintf("Failed to create embedding. %s", err), err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"message": "Saved to ChromaDB",
		})
	}
}

func handleSimilarRequest(_ *core.ServeEvent, collection chroma.Collection) handlerFunc {
	return func(e *core.RequestEvent) error {
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
	}
}
