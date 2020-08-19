package application

import (
	"context"

	"github.com/jnikolaeva/eshop-common/uuid"
)

type userIDContextKeyType string

const userIDContextKey userIDContextKeyType = "userID"

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func GetUserID(ctx context.Context) *uuid.UUID {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok {
		return nil
	}

	return &userID
}
