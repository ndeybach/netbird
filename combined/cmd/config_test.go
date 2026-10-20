package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigCopiesDisableDefaultPolicy(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`server:
  exposedAddress: "https://example.com:443"
  authSecret: "secret"
  disableDefaultPolicy: true
`)
	require.NoError(t, os.WriteFile(configPath, data, 0o600))

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	mgmtConfig, err := cfg.ToManagementConfig()
	require.NoError(t, err)
	require.True(t, mgmtConfig.DisableDefaultPolicy, "server.disableDefaultPolicy should reach the management config")
}
