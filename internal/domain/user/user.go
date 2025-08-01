package user

import "github.com/google/uuid"

type User struct {
	UUID    uuid.UUID `bson:"_id" json:"uuid"`
	ID      int64     `bson:"id" json:"id"`
	Level   int32     `bson:"level" json:"level"`
	Role    string    `bson:"role" json:"role"`
	NanoTon uint64    `bson:"nano_ton" json:"nano_ton"`
}

func NewUser(UUID uuid.UUID, ID int64, level int32, role string, nanoTon uint64) User {
	return User{
		UUID:    UUID,
		ID:      ID,
		Level:   level,
		Role:    role,
		NanoTon: nanoTon,
	}
}
