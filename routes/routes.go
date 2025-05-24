package routes

import (
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/tmc/langchaingo/llms"
)

func RegisterRoutes(app *pocketbase.PocketBase, se *core.ServeEvent, llm *llms.Model, collection chroma.Collection) {
	registerChatRoutes(app, se, llm)
	registerEmbeddingRoutes(app, se, collection)
}
