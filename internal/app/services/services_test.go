package services

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateShortKey(t *testing.T) {
	tests := []struct {
		name    string
		longURL string
		want    string
	}{
		{
			"generate hash",
			"http://mbrgaoyhv.yandex",
			"_SGMGLQIsIM=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GenerateShortKey(tt.longURL))
		})
	}
}
