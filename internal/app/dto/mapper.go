package dto

import (
	"encoding/json"
	"fmt"
)

// Convert converts a source object of type T to a destination object of type U.
func Convert[T, U any](src T) (U, error) {
	var dst U
	data, err := json.Marshal(src)
	if err != nil {
		return dst, fmt.Errorf("failed to marshal source: %w", err)
	}
	if err = json.Unmarshal(data, &dst); err != nil {
		return dst, fmt.Errorf("failed to unmarshal to destination: %w", err)
	}
	return dst, nil
}

// ConvertSlice converts a slice of type T to a slice of type U.
func ConvertSlice[T, U any](src []T) ([]U, error) {
	if src == nil {
		return nil, nil
	}
	dst := make([]U, len(src))
	for i, t := range src {
		converted, err := Convert[T, U](t)
		if err != nil {
			return nil, fmt.Errorf("failed to convert element at index %d: %w", i, err)
		}
		dst[i] = converted
	}
	return dst, nil
}
