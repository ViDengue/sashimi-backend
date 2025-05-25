package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/tmc/langchaingo/llms"
)

type ChatRequest struct {
	Prompt string `json:"prompt"`
}

func registerChatRoutes(_ *pocketbase.PocketBase, se *core.ServeEvent, llm *llms.Model) {
	se.Router.POST("/api/chat", handleChatRequest(se, llm))
}

func handleChatRequest(_ *core.ServeEvent, llm *llms.Model) handlerFunc {
	return func(e *core.RequestEvent) error {
		ctx := context.Background()

		data := ChatRequest{}

		if err := e.BindBody(&data); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}

		/*
			r, w := io.Pipe()

			go func() {
				_, err := llms.GenerateFromSinglePrompt(ctx, *llm, data.Prompt, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
					w.Write(chunk)
					return nil
				}))

				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}

				w.Close()
			}()

			return e.Stream(http.StatusOK, "text/plain", r)
		*/

		prompt := "Answer the question in the following format: Goal, Why It Matters, Action Steps, Key Risks & Solutions, How to Measure Success. Question: " + data.Prompt

		completion, err := llms.GenerateFromSinglePrompt(ctx, *llm, prompt, llms.WithMaxTokens(1000))

		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"completion": completion,
		})
	}
}
