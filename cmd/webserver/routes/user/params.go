package user

type User struct {
	Name string `form:"name" binding:"required" json:"name"`
}
