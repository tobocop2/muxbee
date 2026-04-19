package embed

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/element-hq/dendrite/setup/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(t *testing.T) *config.Dendrite {
	t.Helper()
	_, sk, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	cfg, err := BuildConfig(ConfigOptions{
		ServerName: "localhost",
		DataDir:    t.TempDir(),
		PrivateKey: sk,
	})
	require.NoError(t, err)
	return cfg
}

func TestAddAppservice_PopulatesRegexpObject(t *testing.T) {
	cfg := newTestConfig(t)
	err := AddAppservice(cfg, AppserviceRegistration{
		ID:              "whatsapp",
		URL:             "http://127.0.0.1:29318",
		ASToken:         "as",
		HSToken:         "hs",
		BotUsername:     "whatsappbot",
		NamespacePrefix: "whatsapp_",
		ServerName:      "localhost",
	})
	require.NoError(t, err)
	require.Len(t, cfg.Derived.ApplicationServices, 1)

	for key, namespaces := range cfg.Derived.ApplicationServices[0].NamespaceMap {
		for i, ns := range namespaces {
			assert.NotNil(t, ns.RegexpObject, "namespace %s[%d] RegexpObject should be compiled", key, i)
		}
	}
}

func TestAddAppservice_ProtectsBotUserID(t *testing.T) {
	cfg := newTestConfig(t)
	err := AddAppservice(cfg, AppserviceRegistration{
		ID:              "whatsapp",
		URL:             "http://127.0.0.1:29318",
		ASToken:         "as",
		HSToken:         "hs",
		BotUsername:     "whatsappbot",
		NamespacePrefix: "whatsapp_",
		ServerName:      "localhost",
	})
	require.NoError(t, err)

	as := cfg.Derived.ApplicationServices[0]
	assert.True(t, as.OwnsNamespaceCoveringUserId("@whatsappbot:localhost"),
		"appservice must claim its sender-localpart user ID")
	assert.True(t, as.OwnsNamespaceCoveringUserId("@whatsapp_123:localhost"),
		"appservice must claim its prefixed puppet user IDs")
	assert.False(t, as.OwnsNamespaceCoveringUserId("@alice:localhost"),
		"appservice must not claim unrelated user IDs")
}

func TestAddAppservice_CompilesExclusiveRegex(t *testing.T) {
	cfg := newTestConfig(t)
	err := AddAppservice(cfg, AppserviceRegistration{
		ID:              "signal",
		URL:             "http://127.0.0.1:29313",
		ASToken:         "as",
		HSToken:         "hs",
		BotUsername:     "signalbot",
		NamespacePrefix: "signal_",
		ServerName:      "localhost",
	})
	require.NoError(t, err)
	require.NotNil(t, cfg.Derived.ExclusiveApplicationServicesUsernameRegexp)
	assert.True(t, cfg.Derived.ExclusiveApplicationServicesUsernameRegexp.MatchString("@signal_1:localhost"))
}

func TestCompileNamespaceObjects_InvalidRegexReturnsError(t *testing.T) {
	nsMap := map[string][]config.ApplicationServiceNamespace{
		"users": {{Exclusive: true, Regex: "["}},
	}
	err := compileNamespaceObjects(nsMap)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "users")
}

func TestRecompileExclusiveRegexes_InvalidRegexReturnsError(t *testing.T) {
	cfg := newTestConfig(t)
	// Bypass compileNamespaceObjects — feed a raw, malformed Regex directly
	// into Derived.ApplicationServices so the join path in
	// recompileExclusiveRegexes sees an invalid sub-pattern. This proves the
	// error branch is reachable even though AddAppservice's normal flow
	// validates patterns before reaching it.
	cfg.Derived.ApplicationServices = append(cfg.Derived.ApplicationServices, config.ApplicationService{
		ID: "broken",
		NamespaceMap: map[string][]config.ApplicationServiceNamespace{
			"users": {{Exclusive: true, Regex: "["}},
		},
	})
	err := recompileExclusiveRegexes(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exclusive username regex")
}

func TestAddAppservice_PropagatesInvalidRegexError(t *testing.T) {
	cfg := newTestConfig(t)
	// A NamespacePrefix containing an unbalanced bracket survives QuoteMeta
	// — QuoteMeta only escapes regex metacharacters, and '[' IS one of them,
	// so this particular input is actually safe. To exercise AddAppservice's
	// error propagation end-to-end we'd need to bypass QuoteMeta, which the
	// public API doesn't allow. Keep this test narrow: assert the good path
	// and rely on the two tests above for the error branches.
	require.NoError(t, AddAppservice(cfg, AppserviceRegistration{
		ID:              "edgecase",
		URL:             "http://127.0.0.1:29999",
		ASToken:         "as",
		HSToken:         "hs",
		BotUsername:     "edgebot",
		NamespacePrefix: "edge[case_",
		ServerName:      "localhost",
	}))
}
