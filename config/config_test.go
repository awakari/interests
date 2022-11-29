package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, uint16(8080), cfg.Api.Port)
	assert.Equal(t, "localhost:8080", cfg.Api.Matchers.UriExcludesComplete)
	assert.Equal(t, "mongodb+srv://localhost/?retryWrites=true&w=majority", cfg.Db.Uri)
	assert.Equal(t, "subscriptions", cfg.Db.Name)
	assert.Equal(t, "subscriptions", cfg.Db.Table.Name)
}
