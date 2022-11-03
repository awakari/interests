package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	f, err := os.Open("defaults.yaml")
	require.Nil(t, err)
	defer f.Close()
	cfg, err := NewConfigFromYaml(f)
	assert.Nil(t, err)
	assert.Equal(t, uint16(8080), cfg.Api.Port)
	assert.Equal(t, "matchers:8080", cfg.Api.Matchers.Uri)
	assert.Equal(t, "mongodb+srv://localhost/?retryWrites=true&w=majority", cfg.Db.Uri)
	assert.Equal(t, "subscriptions-staging", cfg.Db.Name)
	assert.Equal(t, "subscriptions", cfg.Db.Table.Name)
}
