package bridges

// BridgeInfo contains metadata about a messaging bridge.
type BridgeInfo struct {
	Name                   string `yaml:"-"`
	Description            string `yaml:"description"`
	Port                   int    `yaml:"port"`
	Note                   string `yaml:"note,omitempty"`
	RequiresAPICredentials bool   `yaml:"requires_api_credentials,omitempty"`
	LoginInstructions      string `yaml:"login_instructions"`
	BridgeV2Ready          bool   `yaml:"bridgev2_ready"`
}

// BotUsername returns the Matrix username for the bridge bot.
func (b BridgeInfo) BotUsername() string {
	return b.Name + "bot"
}

// NamespacePrefix returns the namespace prefix for bridge users/rooms.
func (b BridgeInfo) NamespacePrefix() string {
	return b.Name + "_"
}

// HasNote returns true if this bridge has a note.
func (b BridgeInfo) HasNote() bool {
	return b.Note != ""
}
