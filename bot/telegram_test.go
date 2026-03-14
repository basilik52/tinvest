package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGroupID(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"-1001234567890", -1001234567890, false},
		{"1234567890", 1234567890, false},
		{"0", 0, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseGroupID(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestClient_GetGroupID(t *testing.T) {
	c := &Client{
		groupID: -1001234567890,
	}

	assert.Equal(t, int64(-1001234567890), c.GetGroupID())
}

func TestClient_IsGroupMessage(t *testing.T) {
	c := &Client{
		groupID: -1001234567890,
	}

	tests := []struct {
		name     string
		message  interface{}
		expected bool
	}{
		{
			name:     "nil message",
			message:  nil,
			expected: false,
		},
		{
			name:     "different group",
			message:  map[string]interface{}{"Chat": map[string]interface{}{"ID": int64(-1009876543210)}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message == nil {
				assert.False(t, c.IsGroupMessage(nil))
			}
		})
	}
}
