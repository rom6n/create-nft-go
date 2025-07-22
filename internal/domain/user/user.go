package user

type User struct {
	ID    int64  `bson:"_id" json:"id"`
	Level int32  `bson:"level" json:"level"`
	Role  string `bson:"role" json:"role"`
}

func NewUser(ID int64, level int32, role string) User {
	return User{
		ID:    ID,
		Level: level,
		Role:  role,
	}
}
