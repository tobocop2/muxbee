package bridges

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_ValidBridge(t *testing.T) {
	info := Get("whatsapp")
	require.NotNil(t, info)
	assert.Equal(t, "whatsapp", info.Name)
	assert.Equal(t, 29318, info.Port)
	assert.True(t, info.BridgeV2Ready)
}

func TestGet_InvalidBridge(t *testing.T) {
	info := Get("nonexistent")
	assert.Nil(t, info)
}

func TestList(t *testing.T) {
	all := List()
	assert.NotEmpty(t, all)

	// Verify sorted
	for i := 1; i < len(all); i++ {
		assert.Less(t, all[i-1].Name, all[i].Name)
	}
}

func TestNames(t *testing.T) {
	names := Names()
	assert.NotEmpty(t, names)
	assert.Contains(t, names, "whatsapp")
	assert.Contains(t, names, "signal")

	// Verify sorted
	for i := 1; i < len(names); i++ {
		assert.Less(t, names[i-1], names[i])
	}
}

func TestExists(t *testing.T) {
	assert.True(t, Exists("whatsapp"))
	assert.True(t, Exists("signal"))
	assert.False(t, Exists("nonexistent"))
}

func TestBridgeV2Ready(t *testing.T) {
	ready := BridgeV2Ready()
	assert.NotEmpty(t, ready)
	for _, b := range ready {
		assert.True(t, b.BridgeV2Ready, "bridge %s should be bridgev2 ready", b.Name)
	}
}

func TestExternal(t *testing.T) {
	ext := External()
	assert.NotEmpty(t, ext)
	for _, b := range ext {
		assert.False(t, b.BridgeV2Ready, "bridge %s should not be bridgev2 ready", b.Name)
	}
}

func TestBridgeInfo_BotUsername(t *testing.T) {
	info := BridgeInfo{Name: "whatsapp"}
	assert.Equal(t, "whatsappbot", info.BotUsername())
}

func TestBridgeInfo_NamespacePrefix(t *testing.T) {
	info := BridgeInfo{Name: "signal"}
	assert.Equal(t, "signal_", info.NamespacePrefix())
}

func TestBridgeInfo_HasNote(t *testing.T) {
	assert.False(t, BridgeInfo{}.HasNote())
	assert.True(t, BridgeInfo{Note: "requires cookies"}.HasNote())
}
