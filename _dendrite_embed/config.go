package embed

import (
	"crypto/ed25519"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/element-hq/dendrite/setup/config"
	"github.com/matrix-org/gomatrixserverlib/spec"
)

// ConfigOptions holds parameters for building a Dendrite config.
type ConfigOptions struct {
	ServerName  string
	DataDir     string
	PrivateKey  ed25519.PrivateKey
	Federation  bool
}

// BuildConfig creates a Dendrite config programmatically.
func BuildConfig(opts ConfigOptions) (*config.Dendrite, error) {
	cfg := &config.Dendrite{}
	cfg.Defaults(config.DefaultOpts{
		Generate:       true,
		SingleDatabase: true,
	})

	cfg.Global.ServerName = spec.ServerName(opts.ServerName)
	cfg.Global.PrivateKey = opts.PrivateKey
	cfg.Global.KeyID = "ed25519:beetrix"
	cfg.Global.DisableFederation = !opts.Federation

	dbPath := func(name string) config.DataSource {
		return config.DataSource(fmt.Sprintf("file:%s", filepath.Join(opts.DataDir, name+".db")))
	}

	cfg.UserAPI.AccountDatabase.ConnectionString = dbPath("account")
	cfg.MediaAPI.Database.ConnectionString = dbPath("mediaapi")
	cfg.SyncAPI.Database.ConnectionString = dbPath("syncapi")
	cfg.RoomServer.Database.ConnectionString = dbPath("roomserver")
	cfg.KeyServer.Database.ConnectionString = dbPath("keyserver")
	cfg.FederationAPI.Database.ConnectionString = dbPath("federationapi")
	cfg.RelayAPI.Database.ConnectionString = dbPath("relayapi")

	cfg.Global.JetStream.StoragePath = config.Path(filepath.Join(opts.DataDir, "jetstream"))
	cfg.MediaAPI.BasePath = config.Path(filepath.Join(opts.DataDir, "media"))
	cfg.MediaAPI.AbsBasePath = config.Path(filepath.Join(opts.DataDir, "media"))
	cfg.SyncAPI.Fulltext.Enabled = true
	cfg.SyncAPI.Fulltext.IndexPath = config.Path(filepath.Join(opts.DataDir, "searchindex"))

	cfg.ClientAPI.RegistrationDisabled = true

	if err := cfg.Derive(); err != nil {
		return nil, fmt.Errorf("deriving config: %w", err)
	}

	return cfg, nil
}

// AppserviceRegistration holds parameters for registering a bridge appservice.
type AppserviceRegistration struct {
	ID              string
	URL             string
	ASToken         string
	HSToken         string
	BotUsername     string
	NamespacePrefix string
	ServerName      string
}

// AddAppservice registers a bridge appservice with the Dendrite config.
// Must be called before Start().
func AddAppservice(cfg *config.Dendrite, reg AppserviceRegistration) error {
	quotedServer := regexp.QuoteMeta(reg.ServerName)
	as := config.ApplicationService{
		ID:              reg.ID,
		URL:             reg.URL,
		ASToken:         reg.ASToken,
		HSToken:         reg.HSToken,
		SenderLocalpart: reg.BotUsername,
		RateLimited:     false,
		Protocols:       []string{reg.ID},
		NamespaceMap: map[string][]config.ApplicationServiceNamespace{
			"users": {
				// Prefixed puppet userIDs (e.g. @whatsapp_123:localhost).
				{
					Exclusive: true,
					Regex:     fmt.Sprintf("@%s.*:%s", regexp.QuoteMeta(reg.NamespacePrefix), quotedServer),
				},
				// Bot sender userID (e.g. @whatsappbot:localhost). Dendrite's
				// native YAML loader appends this; we must match that behavior
				// so the bot name is reserved via an exclusive namespace.
				{
					Exclusive: true,
					Regex:     regexp.QuoteMeta(fmt.Sprintf("@%s:%s", reg.BotUsername, reg.ServerName)),
				},
			},
			"aliases": {
				{
					Exclusive: true,
					Regex:     fmt.Sprintf("#%s.*:%s", regexp.QuoteMeta(reg.NamespacePrefix), quotedServer),
				},
			},
		},
	}

	if err := compileNamespaceObjects(as.NamespaceMap); err != nil {
		return fmt.Errorf("appservice %q: %w", reg.ID, err)
	}
	cfg.Derived.ApplicationServices = append(cfg.Derived.ApplicationServices, as)
	recompileExclusiveRegexes(cfg)
	return nil
}

// compileNamespaceObjects populates each namespace's RegexpObject so Dendrite's
// per-namespace matching (OwnsNamespaceCoveringUserId, etc.) doesn't nil-panic.
// Returns an error if any regex fails to compile — surfacing this matters
// because a silent failure reproduces the exact nil-panic this code fixes.
func compileNamespaceObjects(nsMap map[string][]config.ApplicationServiceNamespace) error {
	for key, namespaces := range nsMap {
		for i := range namespaces {
			r, err := regexp.Compile(namespaces[i].Regex)
			if err != nil {
				return fmt.Errorf("namespace %q regex %q: %w", key, namespaces[i].Regex, err)
			}
			namespaces[i].RegexpObject = r
		}
	}
	return nil
}

func recompileExclusiveRegexes(cfg *config.Dendrite) {
	var userPatterns, aliasPatterns []string

	for _, as := range cfg.Derived.ApplicationServices {
		if users, ok := as.NamespaceMap["users"]; ok {
			for _, ns := range users {
				if ns.Exclusive {
					userPatterns = append(userPatterns, ns.Regex)
				}
			}
		}
		if aliases, ok := as.NamespaceMap["aliases"]; ok {
			for _, ns := range aliases {
				if ns.Exclusive {
					aliasPatterns = append(aliasPatterns, ns.Regex)
				}
			}
		}
	}

	if len(userPatterns) > 0 {
		combined := "(" + strings.Join(userPatterns, ")|(") + ")"
		cfg.Derived.ExclusiveApplicationServicesUsernameRegexp, _ = regexp.Compile(combined)
	}
	if len(aliasPatterns) > 0 {
		combined := "(" + strings.Join(aliasPatterns, ")|(") + ")"
		cfg.Derived.ExclusiveApplicationServicesAliasRegexp, _ = regexp.Compile(combined)
	}
}
