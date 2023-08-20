package user

import "weezel/example-gin/pkg/generated/sqlc"

type HandlerController struct {
	querier sqlc.Querier
}

func NewHandlerController(db sqlc.Querier) *HandlerController {
	return &HandlerController{querier: db}
}
