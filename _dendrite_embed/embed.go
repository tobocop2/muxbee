// Package embed provides a public API for embedding Dendrite as an in-process
// Matrix homeserver. This wraps Dendrite's internal packages to expose a clean
// interface for external consumers.
package embed

import (
	"context"
	"fmt"

	"github.com/element-hq/dendrite/appservice"
	"github.com/element-hq/dendrite/federationapi"
	"github.com/element-hq/dendrite/internal/caching"
	"github.com/element-hq/dendrite/internal/httputil"
	"github.com/element-hq/dendrite/internal/sqlutil"
	"github.com/element-hq/dendrite/roomserver"
	"github.com/element-hq/dendrite/setup"
	basepkg "github.com/element-hq/dendrite/setup/base"
	"github.com/element-hq/dendrite/setup/config"
	"github.com/element-hq/dendrite/setup/jetstream"
	"github.com/element-hq/dendrite/setup/process"
	"github.com/element-hq/dendrite/userapi"
	userapi_api "github.com/element-hq/dendrite/userapi/api"
	"github.com/matrix-org/gomatrixserverlib/spec"
)

// Server is an embedded Dendrite homeserver.
type Server struct {
	Config     *config.Dendrite
	ProcessCtx *process.ProcessContext
	UserAPI    userapi_api.UserInternalAPI
	Monolith   setup.Monolith
}

// Start boots the Dendrite monolith and begins serving HTTP on the given address.
func Start(cfg *config.Dendrite, listenAddr string) (*Server, error) {
	processCtx := process.NewProcessContext()

	cm := sqlutil.NewConnectionManager(processCtx, cfg.Global.DatabaseOptions)
	routers := httputil.NewRouters()
	natsInstance := &jetstream.NATSInstance{}
	caches := caching.NewRistrettoCache(
		cfg.Global.Cache.EstimatedMaxSize,
		cfg.Global.Cache.MaxAge,
		caching.DisableMetrics,
	)

	federationClient := basepkg.CreateFederationClient(cfg, nil)
	httpClient := basepkg.CreateClient(cfg, nil)

	rsAPI := roomserver.NewInternalAPI(processCtx, cfg, cm, natsInstance, caches, caching.DisableMetrics)
	fsAPI := federationapi.NewInternalAPI(
		processCtx, cfg, cm, natsInstance, federationClient, rsAPI, caches, nil, false,
	)
	keyRing := fsAPI.KeyRing()
	rsAPI.SetFederationAPI(fsAPI, keyRing)

	userAPI := userapi.NewInternalAPI(
		processCtx, cfg, cm, natsInstance, rsAPI, federationClient,
		caching.DisableMetrics, fsAPI.IsBlacklistedOrBackingOff,
	)

	asAPI := appservice.NewInternalAPI(processCtx, cfg, natsInstance, userAPI, rsAPI)
	rsAPI.SetAppserviceAPI(asAPI)
	rsAPI.SetUserAPI(userAPI)

	monolith := setup.Monolith{
		Config:        cfg,
		Client:        httpClient,
		FedClient:     federationClient,
		KeyRing:       keyRing,
		AppserviceAPI: asAPI,
		FederationAPI: fsAPI,
		RoomserverAPI: rsAPI,
		UserAPI:       userAPI,
	}

	monolith.AddAllPublicRoutes(processCtx, cfg, routers, cm, natsInstance, caches, caching.DisableMetrics)

	// Use Dendrite's own HTTP server setup (handles gorilla/mux routing correctly)
	httpAddr, err := config.HTTPAddress("http://" + listenAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing listen address %q: %w", listenAddr, err)
	}
	go basepkg.SetupAndServeHTTP(processCtx, cfg, routers, httpAddr, nil, nil)

	return &Server{
		Config:     cfg,
		ProcessCtx: processCtx,
		UserAPI:    userAPI,
		Monolith:   monolith,
	}, nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	if s.ProcessCtx != nil {
		s.ProcessCtx.ShutdownDendrite()
		s.ProcessCtx.WaitForComponentsToFinish()
	}
}

// CreateUser creates a Matrix user account directly via the UserAPI.
func (s *Server) CreateUser(ctx context.Context, localpart, password string, admin bool) error {
	accountType := userapi_api.AccountTypeUser
	if admin {
		accountType = userapi_api.AccountTypeAdmin
	}

	var res userapi_api.PerformAccountCreationResponse
	return s.UserAPI.PerformAccountCreation(ctx, &userapi_api.PerformAccountCreationRequest{
		AccountType: accountType,
		Localpart:   localpart,
		ServerName:  spec.ServerName(s.Config.Global.ServerName),
		Password:    password,
		OnConflict:  userapi_api.ConflictUpdate,
	}, &res)
}
