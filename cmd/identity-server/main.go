package main

import (
	"context"
	"crypto/tls"
	"time"

	"identity-metadata-server/internal/certificates"
	"identity-metadata-server/internal/shared"

	"github.com/Depado/ginprom"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/trivago/go-bootstrap/config"
	"github.com/trivago/go-bootstrap/httpserver"
	_ "github.com/trivago/go-bootstrap/logging"
)

// initPrometheus will initialize the prometheus metrics for the server.
func initPrometheus(router *gin.Engine) {
	// We create a new registry as using the default registry somehow
	// generates a go routine leak.
	registry := prometheus.NewRegistry()
	prometheus.DefaultGatherer = registry
	prometheus.DefaultRegisterer = registry

	prometheus.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	prometheus.MustRegister(collectors.NewGoCollector())

	// Initialize the ginprom middleware with the custom registry.
	// This will automatically register the metrics with the default registry.
	// The middleware will also add the client IP as a label to the metrics.
	prom := ginprom.New(
		ginprom.Engine(router),
		ginprom.CustomCounterLabels(
			[]string{"client_ip"},
			func(c *gin.Context) map[string]string {
				clientIP := c.ClientIP()
				if clientIP == "" {
					clientIP = "unknown"
				}
				return map[string]string{
					"client_ip": clientIP,
				}
			}),
	)
	router.Use(prom.Instrument())
}

func main() {
	viper.SetDefault("port", 8443)
	viper.SetDefault("maxRequestDuration", 5*time.Second)
	// If the server key changes, the JWKS must be re-registered with the workload identity provider
	viper.SetDefault("server.key", "server.pem")
	// If the keyName is changed, the JWKS must be re-registered with the workload identity provider
	viper.SetDefault("server.keyName", "trivago-identity-server-01")
	// The server.issuer must match the one used in the workload identity provider
	viper.SetDefault("server.issuer", "https://identity-server")
	// The service account bound to the identity server
	viper.SetDefault("server.identity", "identity-server@trv-identity-server-testing.iam.gserviceaccount.com")
	// This can be used to overwrite the hostname
	viper.SetDefault("server.hostname", "")
	// The idle timeout for HTTP2 connections.
	viper.SetDefault("server.idleTimeout", "620s")

	viper.SetDefault("server.workloadIdentity.projectNumber", "866597189115")
	viper.SetDefault("server.workloadIdentity.poolName", "integration-test")
	viper.SetDefault("server.workloadIdentity.providerName", "identity-server")
	viper.SetDefault("server.certAuthority.project", "trv-identity-server-testing")
	viper.SetDefault("server.certAuthority.region", "europe-west1")
	viper.SetDefault("server.certAuthority.poolName", "integration-test-ca-pool")
	viper.SetDefault("server.certAuthority.name", "identity-server-ca")
	viper.SetDefault("server.certAuthority.crlRefresh", "24h")
	viper.SetDefault("server.certAuthority.clientCertLifetime", "2160h")

	viper.SetDefault("tls.certificate", "/etc/certs/tls.crt")
	viper.SetDefault("tls.key", "/etc/certs/tls.key")
	viper.SetDefault("tls.reload", time.Hour*24)

	viper.SetDefault("profile", false)

	config.Read("IDS", "config.yaml")

	if err := initJWKS(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize JWKS")
		return
	}

	// Configure mTLS certificate verification
	// Note: We don't require the client to present a certificate.
	// Endpoints that require a client certificate will need to check for it.

	project := viper.GetString("server.certAuthority.project")
	region := viper.GetString("server.certAuthority.region")
	poolName := viper.GetString("server.certAuthority.poolName")
	caName := viper.GetString("server.certAuthority.name")
	clientCanBeVerified := true

	clientCertLifetime := viper.GetDuration("server.certAuthority.clientCertLifetime")
	switch {
	case clientCertLifetime <= 0:
		log.Fatal().Msg("Client certificate lifetime is negative or 0. This is not allowed.")
	case clientCertLifetime < time.Hour*24:
		log.Warn().Msg("Client certificate lifetime is less than 24h. This is not recommended.")
	case clientCertLifetime > time.Hour*24*90:
		log.Warn().Msg("Client certificate lifetime is more than 90d. This is not recommended.")
	}

	clientRootCAs, clientRootCAPool, err := InitClientRootCA(project, region, poolName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load client root CA. Switching to JWKS mode")
		// We won't be able to verify client certificates.
		// As of this all mTLS based endpoints will be disabled.
		clientCanBeVerified = false
	}

	var revocationList *CertificateRevocationList
	if clientCanBeVerified {
		revocationList = NewCertificateRevocationList(
			clientRootCAs,
			project,
			region,
			poolName,
			caName,
			viper.GetDuration("server.certAuthority.crlRefresh"),
		)

		// Update the CRL and start the timer
		if err := revocationList.Update(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to update revoked certificates. Switching to JWKS mode")
			// We won't be able to detect revoked certificates.
			// As of this all mTLS based endpoints will be disabled.
			clientCanBeVerified = false
		}

		// Start the timer for automatic updates
		revocationList.StartUpdateTimer()

		// Ensure the timer goroutine is stopped when the server shuts down
		defer revocationList.StopUpdateTimer()
	}

	caConfig := certificates.GCPCertificateAuthorityConfig{
		ProjectID:            project,
		Location:             region,
		CertificatePool:      poolName,
		CertificateAuthority: caName,
	}

	// Configure the server
	config := httpserver.Config{
		Port:        viper.GetInt("port"),
		PathTLSCert: viper.GetString("tls.certificate"),
		PathTLSKey:  viper.GetString("tls.key"),
		Health:      httpserver.AlwaysOk,
		Ready:       httpserver.AlwaysOk,
		InitRoutes: func(router *gin.Engine) {
			initPrometheus(router)

			maxRequestDuration := viper.GetDuration("maxRequestDuration")
			router.Use(func(g *gin.Context) {
				shared.ForceMaxDuration(maxRequestDuration, g)
			})

			if viper.GetBool("profile") {
				pprof.Register(router, "/debug/pprof")
			}

			router.GET("/jwks.json", HandleJWKSRequest)

			if clientCanBeVerified {
				router.GET("/token", func(c *gin.Context) { HandleTokenRequest(c, revocationList) })
				router.GET("/identity", func(c *gin.Context) { HandleIdentityRequest(c, revocationList) })
				router.POST("/refreshCrl", func(c *gin.Context) { HandleRefreshRequest(c, revocationList) })
				router.POST("/renew", func(c *gin.Context) { HandleRenewRequest(c, revocationList, caConfig) })
			}
		},
		DisableAccessLogFor: []string{
			"/healthz",
			"/readyz",
			"/metrics",
		},
	}

	srv, err := httpserver.NewWithConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	// Prevent open HTTP2 connections from piling up.
	srv.IdleTimeout = viper.GetDuration("server.idleTimeout")

	if clientCanBeVerified {
		// We have some endpoint that don't require mTLS (like healthcheck)
		// So we do not make the client certificate mandatory and check for the certificate
		// in code. However, we still want the handshake to fail early if a wrong certificate is presented.
		srv.TLSConfig.ClientAuth = tls.VerifyClientCertIfGiven
		srv.TLSConfig.ClientCAs = clientRootCAPool
	}

	httpserver.Listen(srv, nil)
}
