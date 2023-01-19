package config

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, uint16(8080), cfg.Api.Port)
	assert.Equal(t, "kiwi-tree-complete:8080", cfg.Api.KiwiTree.CompleteUri)
	assert.Equal(t, "kiwi-tree-partial:8080", cfg.Api.KiwiTree.PartialUri)
	assert.Equal(t, "mongodb+srv://localhost/?retryWrites=true&w=majority", cfg.Db.Uri)
	assert.Equal(t, "subscriptions", cfg.Db.Name)
	assert.Equal(t, "subscriptions", cfg.Db.Table.Name)
	assert.Equal(t, slog.DebugLevel, cfg.Log.Level)
}
