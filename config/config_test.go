package config

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("API_PORT", "55555")
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, uint16(55555), cfg.Api.Port)
	assert.Equal(t, "mongodb://localhost:27017/?retryWrites=true&w=majority", cfg.Db.Uri)
	assert.Equal(t, "subscriptions", cfg.Db.Name)
	assert.Equal(t, "subscriptions", cfg.Db.Table.Name)
	assert.Equal(t, int(slog.LevelDebug), cfg.Log.Level)
}
