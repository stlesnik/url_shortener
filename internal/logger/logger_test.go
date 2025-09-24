package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		env         string
		expectErr   bool
		expectNil   bool
		assertsPost func(t *testing.T)
	}{
		{
			name:      "prod environment",
			env:       "prod",
			expectErr: false,
			expectNil: false,
		},
		{
			name:      "production environment",
			env:       "production",
			expectErr: false,
			expectNil: false,
		},
		{
			name:      "dev environment",
			env:       "dev",
			expectErr: false,
			expectNil: false,
		},
		{
			name:      "empty environment",
			env:       "",
			expectErr: false,
			expectNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.env)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, Sugaarz)
			} else {
				assert.NotNil(t, Sugaarz)
			}
		})
	}
}
