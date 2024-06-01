package user

import (
	"context"

	"weezel/example-gin/pkg/generated/sqlc"

	"github.com/gin-gonic/gin"
)

type HandlerController struct {
	querier         sqlc.Querier
	userRouterGroup *gin.RouterGroup
}

func NewHandlerController(userRouterGroup *gin.RouterGroup, db sqlc.Querier) *HandlerController {
	return &HandlerController{
		userRouterGroup: userRouterGroup,
		querier:         db,
	}
}

// MockHandlerController introduces method calls that can be implemented on test case basis
type MockHandlerController struct {
	mAdduser    func(context.Context, sqlc.AddUserParams) (int32, error)
	mDeleteUser func(context.Context, string) (*sqlc.HomepageSchemaUser, error)
	mGetUser    func(context.Context, string) (*sqlc.HomepageSchemaUser, error)
	mListUsers  func(context.Context) ([]*sqlc.HomepageSchemaUser, error)
}

func (h MockHandlerController) AddUser(ctx context.Context, arg sqlc.AddUserParams) (int32, error) {
	return h.mAdduser(ctx, arg)
}

func (h MockHandlerController) DeleteUser(ctx context.Context, name string) (*sqlc.HomepageSchemaUser, error) {
	return h.mDeleteUser(ctx, name)
}

func (h MockHandlerController) GetUser(ctx context.Context, name string) (*sqlc.HomepageSchemaUser, error) {
	return h.mGetUser(ctx, name)
}

func (h MockHandlerController) ListUsers(ctx context.Context) ([]*sqlc.HomepageSchemaUser, error) {
	return h.mListUsers(ctx)
}
