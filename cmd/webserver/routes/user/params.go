package user

type User struct {
	Name string `form:"name" binding:"required" json:"name"`
	ID   int32  `form:"id" validate:"gte=0" json:"id"`
	Age  int32  `form:"age" validate:"gte=0,lte=130" json:"age"`
}
