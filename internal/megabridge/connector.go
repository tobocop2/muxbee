package megabridge

import (
	"fmt"

	"maunium.net/go/mautrix/bridgev2"

	"go.mau.fi/mautrix-whatsapp/pkg/connector"
)

// NewConnector creates a NetworkConnector for the given bridge name.
func NewConnector(name string) (bridgev2.NetworkConnector, error) {
	switch name {
	case "whatsapp":
		return &connector.WhatsAppConnector{}, nil
	default:
		return nil, fmt.Errorf("no built-in connector for bridge: %s", name)
	}
}
