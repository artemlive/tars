package utils

import (
	"encoding/json"
	"fmt"
)

func DecodeEvent[T any](rawEvent interface{}) (T, error) {
	var event T

	// Convert rawEvent (map[string]interface{}) to JSON bytes
	jsonBytes, err := json.Marshal(rawEvent)
	if err != nil {
		return event, fmt.Errorf("failed to marshal raw event: %w", err)
	}

	// Unmarshal JSON bytes into the target type
	if err := json.Unmarshal(jsonBytes, &event); err != nil {
		return event, fmt.Errorf("failed to decode event: %w", err)
	}

	return event, nil
}
