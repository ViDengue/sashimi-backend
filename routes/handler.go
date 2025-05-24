package routes

import "github.com/pocketbase/pocketbase/core"

type handlerFunc func(*core.RequestEvent) error
