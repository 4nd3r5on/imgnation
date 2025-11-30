package usersRepo

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NewUserReqData struct {
	Username string   `bson:"username" json:"username"`
	Name     string   `bson:"name" json:"name"`
	Email    string   `bson:"email" json:"email"`
	Roles    []string `bson:"relos" json:"roles"` // allowed to control only for privileged tokens
	Password string   `bson:"password" json:"password"`
}

// Provided to repository
type NewUserRepoData struct {
	ID           uuid.UUID `bson:"id" json:"id"`
	Username     string    `bson:"username" json:"username"`
	Name         string    `bson:"name" json:"name"`
	Email        string    `bson:"email" json:"email"`
	Roles        []string  `bson:"relos" json:"roles"`
	PasswordHash []byte    `bson:"password_hash" json:"password_hash"`
}

type StoredUser struct {
	ID               uuid.UUID `bson:"id" json:"id"` // unique
	ProfilePicfileID uuid.UUID `bson:"profile_pic_file_id" json:"profile_pic_file_id"`
	Username         string    `bson:"username" json:"username"` // unique
	Name             string    `bson:"name" json:"name"`
	Email            string    `bson:"email" json:"email"`
	Roles            []string  `bson:"roles" json:"roles"` // default ["user"]
	PasswordHash     []byte    `bson:"password_hash" json:"password_hash"`

	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type UpdateUser struct {
	ProfilePicfileID *uuid.UUID `bson:"profile_pic_file_id" json:"profile_pic_file_id"`
	Username         *string    `bson:"username" json:"username"` // unique
	Name             *string    `bson:"name" json:"name"`
	Email            *string    `bson:"email" json:"email"`
	AddRoles         []string   `bson:"add_roles" json:"add_roles"`
	RemoveRoles      []string   `bson:"remove_roles" json:"remove_roles"`
	PasswordHash     []byte     `bson:"password_hash" json:"password_hash"`
}

type UserCredentials struct {
	Email        string `bson:"email" json:"email"`
	PasswordHash []byte `bson:"password" json:"password"`
}

type EmailIdPair struct {
	Email string    `bson:"email" json:"email"`
	ID    uuid.UUID `bson:"id" json:"id"`
}

type UsersRepo interface {
	CreateUser(ctx context.Context, userData NewUserRepoData) (id uuid.UUID, err error)
	RemoveUser(ctx context.Context, userID uuid.UUID) error
	CheckUsernameTaken(ctx context.Context, username string) (isTaken bool, err error)

	GetUser(ctx context.Context, id uuid.UUID) (user StoredUser, err error)
	GetUserByEmail(ctx context.Context, email string) (user StoredUser, err error)
	GetUserByUsername(ctx context.Context, username string) (user StoredUser, err error)
	GetIDsByEmails(ctx context.Context, emails []string) ([]EmailIdPair, error)

	UpdateUser(ctx context.Context, id uuid.UUID, update UpdateUser) error

	GetUsersCount(ctx context.Context) (int64, error)
	CreateIndexes(ctx context.Context) error
}
