package usersRepo

import (
	"context"
	"errors"
	"time"

	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UsersRepository struct {
	collection *mongo.Collection
}

func NewUsersRepository(ctx context.Context, db *mongo.Database) (*UsersRepository, error) {
	collection := db.Collection("users")
	repo := &UsersRepository{
		collection: collection,
	}
	err := repo.CreateIndexes(ctx)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to create user repository indexes")
	}
	return repo, nil
}

// CreateUser creates a new user in the database
func (r *UsersRepository) CreateUser(ctx context.Context, userData NewUserRepoData) (uuid.UUID, error) {
	now := time.Now()

	storedUser := StoredUser{
		ID:               userData.ID,
		ProfilePicfileID: uuid.Nil, // Default empty UUID
		Username:         userData.Username,
		Name:             userData.Name,
		Email:            userData.Email,
		Roles:            userData.Roles,
		PasswordHash:     userData.PasswordHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Set default roles if empty
	if len(storedUser.Roles) == 0 {
		storedUser.Roles = []string{"user"}
	}

	_, err := r.collection.InsertOne(ctx, storedUser)
	if err != nil {
		// Check for duplicate key error
		if mongo.IsDuplicateKeyError(err) {
			return uuid.Nil, werr.Wrapf(xerr.ErrEntityExists, "username or email already exists")
		}
		return uuid.Nil, err
	}

	return userData.ID, nil
}

// RemoveUser removes a user by ID
func (r *UsersRepository) RemoveUser(ctx context.Context, userID uuid.UUID) error {
	filter := bson.M{"id": userID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return werr.Wrapf(xerr.ErrEntityNotFound, "user to delete wasn't found")
	}
	return nil
}

// CheckUsernameTaken checks if a username is already taken
func (r *UsersRepository) CheckUsernameTaken(ctx context.Context, username string) (bool, error) {
	filter := bson.M{"username": username}
	count, err := r.collection.CountDocuments(ctx, filter)
	return count > 0, err
}

func (r *UsersRepository) getUser(ctx context.Context, filter bson.M) (StoredUser, error) {
	var user StoredUser
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return StoredUser{}, werr.Wrapf(xerr.ErrEntityNotFound, "user wasn't found")
		}
		return StoredUser{}, err
	}
	return user, nil
}

// GetUser retrieves a user by ID
func (r *UsersRepository) GetUser(ctx context.Context, id uuid.UUID) (StoredUser, error) {
	return r.getUser(ctx, bson.M{"id": id})
}

// GetUserByEmail retrieves a user by E-Mail
func (r *UsersRepository) GetUserByEmail(ctx context.Context, email string) (StoredUser, error) {
	return r.getUser(ctx, bson.M{"email": email})
}

// GetUserByUsername retrieves a user by username
func (r *UsersRepository) GetUserByUsername(ctx context.Context, username string) (StoredUser, error) {
	return r.getUser(ctx, bson.M{"username": username})
}

func (r *UsersRepository) GetIDsByEmails(ctx context.Context, emails []string) ([]EmailIdPair, error) {
	if len(emails) == 0 {
		return []EmailIdPair{}, nil
	}
	filter := bson.M{"email": bson.M{"$in": emails}}
	opts := options.Find().SetProjection(bson.M{"email": 1, "id": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	results := make([]EmailIdPair, len(emails))
	var i int64 = 0
	for cursor.Next(ctx) {
		if err := cursor.Decode(&results[i]); err != nil {
			return nil, err
		}
		i++
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateUser updates user information
func (r *UsersRepository) UpdateUser(ctx context.Context, id uuid.UUID, update UpdateUser) error {
	filter := bson.M{"id": id}

	// Build update document
	updateDoc := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Add fields to update if they are provided
	if update.ProfilePicfileID != nil {
		updateDoc["$set"].(bson.M)["profile_pic_file_id"] = *update.ProfilePicfileID
	}
	if update.Username != nil {
		updateDoc["$set"].(bson.M)["username"] = *update.Username
	}
	if update.Name != nil {
		updateDoc["$set"].(bson.M)["name"] = *update.Name
	}
	if update.Email != nil {
		updateDoc["$set"].(bson.M)["email"] = *update.Email
	}
	if len(update.PasswordHash) > 0 {
		updateDoc["$set"].(bson.M)["password_hash"] = update.PasswordHash
	}

	// Handle role additions
	if len(update.AddRoles) > 0 {
		updateDoc["$addToSet"] = bson.M{
			"roles": bson.M{"$each": update.AddRoles},
		}
	}

	// Handle role removals
	if len(update.RemoveRoles) > 0 {
		if updateDoc["$pull"] == nil {
			updateDoc["$pull"] = bson.M{}
		}
		updateDoc["$pull"].(bson.M)["roles"] = bson.M{"$in": update.RemoveRoles}
	}

	result, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		// Check for duplicate key error
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("username or email already exists")
		}
		return err
	}

	if result.MatchedCount == 0 {
		return werr.Wrapf(xerr.ErrEntityNotFound, "user wasn't found")
	}

	return nil
}

// GetUsersCount returns the total number of users
func (r *UsersRepository) GetUsersCount(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// CreateIndexes creates necessary indexes for the users collection
func (r *UsersRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
