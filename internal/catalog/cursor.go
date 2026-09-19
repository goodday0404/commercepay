package catalog

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type productCursorPayload struct {
	Version   int       `json:"v"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

func encodeProductCursor(cursor ProductCursor) (string, error) {
	payload := productCursorPayload{
		Version:   currentCursorVersion,
		CreatedAt: cursor.CreatedAt,
		ID:        cursor.ID.String(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeProductCursor(value string) (ProductCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ProductCursor{}, fmt.Errorf("decoe product cursor: %w", err)
	}

	var payload productCursorPayload

	if err := json.Unmarshal(data, &payload); err != nil {
		return ProductCursor{}, fmt.Errorf("unmarshal product cursor payload %w", err)
	}

	if payload.Version != currentCursorVersion {
		return ProductCursor{}, errors.New("unsupported cursor version")
	}

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		return ProductCursor{}, fmt.Errorf("parse product cursor id: %w", err)
	}

	if payload.CreatedAt.IsZero() {
		return ProductCursor{}, errors.New("invalid cursor timestamp")
	}

	return ProductCursor{
		CreatedAt: payload.CreatedAt,
		ID:        id,
	}, nil
}
