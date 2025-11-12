package main

import (
	"strings"
	"time"

	"identity-metadata-server/internal/shared"
	"identity-metadata-server/internal/tokenprovider"

	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/trivago/go-bootstrap/config"
	"github.com/trivago/go-bootstrap/httpserver"
	_ "github.com/trivago/go-bootstrap/logging"
	kubernetes "github.com/trivago/go-kubernetes/v4"
)

var (
	endpoints     map[string]func(*gin.Context)
	tokenProvider tokenprovider.TokenProvider
	knownTokens   *TokenCache
)

var (
	AccessTokenLifetime   = 10 * time.Minute
	IdentityTokenLifetime = 10 * time.Minute
)

const (
	tokenTimeFormat = time.RFC3339
)

func init() {
	endpoints = map[string]func(*gin.Context){
		"/computeMetadata/":                                                      HandleOk,
		"/computeMetadata/v1/":                                                   HandleListEndpoints,
		"/computeMetadata/v1/project/":                                           HandleListEndpoints,
		"/computeMetadata/v1/project/project-id":                                 HandleGetProjectId,
		"/computeMetadata/v1/project/numeric-project-id":                         HandleGetProjectNumber,
		"/computeMetadata/v1/universe/":                                          HandleListEndpoints,
		"/computeMetadata/v1/universe/universe-domain":                           HandleGetUniverse,
		"/computeMetadata/v1/instance/":                                          HandleListEndpoints,
		"/computeMetadata/v1/instance/service-accounts/":                         HandleGetServiceAccounts,
		"/computeMetadata/v1/instance/service-accounts/default/email":            HandleGetDefaultServiceAccount,
		"/computeMetadata/v1/instance/service-accounts/:serviceAccount/token":    HandleGetAccessToken,
		"/computeMetadata/v1/instance/service-accounts/:serviceAccount/identity": HandleGetIdentityToken,
		"/computeMetadata/v1/instance/service-accounts/:serviceAccount":          HandleGetServiceAccountInfo,
		"/computeMetadata/v1/instance/service-accounts/:serviceAccount/scopes":   HandleGetServiceAccountScopes,
	}
}

// initGinEndpoints will automatically populate the gin router with the
// handlers defined in the endpoints map.
func initGinEndpoints(router *gin.Engine) {
	initPrometheus(router)

	maxRequestDuration := viper.GetDuration("maxRequestDuration")
	router.Use(func(g *gin.Context) {
		shared.ForceMaxDuration(maxRequestDuration, g)
	})

	for path, handler := range endpoints {
		router.GET(path, handler)
	}

	// This is important for golang clients to work correctly
	// the golang library will do two check in parallel, using the result of the first one
	// to respond. The DNS check works always, however if this is not returing 200 with
	// the correct metadata flavor, the client will _sometimes_ fail.
	router.GET("/", HandleOk)
}

// initPrometheus will initialize the prometheus metrics for the server.
// We rely on the default prom-client default registry, but we add some
// custom labels to the metrics to distinguish between different instances.
func initPrometheus(router *gin.Engine) {
	registry := prometheus.NewRegistry()
	prometheus.DefaultGatherer = registry

	// Make sure the following labels are always present in the metrics
	// This is useful to distinguish between different instances of the server
	// and to know which mode is being used (host or kubernetes).
	prometheus.DefaultRegisterer = prometheus.WrapRegistererWith(prometheus.Labels{
		"node_name": shared.GetNodename(),
		"mode":      strings.ToLower(viper.GetString("mode")),
	}, registry)

	// These are the default collectors that we want to register.
	// They will collect metrics about the process and the Go runtime.
	// We cannot use the auto-generated collectors initialized by prometheus-client,
	// as we use a custom registry to add some labels to the metrics.
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

// initConfigDefaults sets the default values for the available configuration
// options.
func initConfigDefaults() {
	viper.SetDefault("port", 8080)
	viper.SetDefault("projectId", "trv-identity-server-testing")
	viper.SetDefault("projectNumber", "866597189115")
	viper.SetDefault("poolName", "kubernetes-pool")
	viper.SetDefault("providerName", "production")
	viper.SetDefault("mode", "kubernetes")
	viper.SetDefault("maxRequestDuration", 3*time.Second)
	viper.SetDefault("cache.serviceAccountTTL", 2*time.Minute)
	viper.SetDefault("cache.tokenCleanupInterval", time.Hour)
	viper.SetDefault("cache.tokenMinLifetime", 1*time.Minute)
	viper.SetDefault("kubernetes.kubeletHost", "https://127.0.0.1:10250")
	viper.SetDefault("kubernetes.kubeletCaPath", "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	viper.SetDefault("host.identityServer", "https://identity-server:443")
	viper.SetDefault("host.clientCert", "/etc/certs/machine/identity.pem")
	viper.SetDefault("host.clientKey", "/etc/certs/machine/identity.key")
	viper.SetDefault("host.cacert", "")
	viper.SetDefault("host.clientCertMinimumLifetime", time.Hour*24*10)
	viper.SetDefault("host.clientCertRefresh", time.Hour*24)
	viper.SetDefault("token.lifetime.access", 10*time.Minute)
	viper.SetDefault("token.lifetime.identity", 10*time.Minute)
}

func main() {
	initConfigDefaults()
	config.Read("AUTH", "config.yaml")

	srv, err := httpserver.NewWithConfig(httpserver.Config{
		Port:       viper.GetInt("port"),
		Health:     httpserver.AlwaysOk,
		Ready:      httpserver.AlwaysOk,
		InitRoutes: initGinEndpoints,
		DisableAccessLogFor: []string{
			"/healthz",
			"/readyz",
			"/metrics",
		},
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	mode := viper.GetString("mode")
	workloadIdentityAudience := shared.GetWorkloadIdentityAudience(
		viper.GetString("projectNumber"),
		viper.GetString("poolName"),
		viper.GetString("providerName"))

	AccessTokenLifetime = viper.GetDuration("token.lifetime.access")
	IdentityTokenLifetime = viper.GetDuration("token.lifetime.identity")

	switch strings.ToLower(viper.GetString("mode")) {
	case "host":
		log.Info().Msg("Using host mode")
		identityServerURL := viper.GetString("host.identityServer")
		caCertPath := viper.GetString("host.cacert")
		clientCertPath := viper.GetString("host.clientCert")
		clientKeyPath := viper.GetString("host.clientKey")
		minCertLifetime := viper.GetDuration("host.clientCertMinimumLifetime")
		refreshInterval := viper.GetDuration("host.clientCertRefresh")

		if refreshInterval > minCertLifetime {
			log.Fatal().Msg("The client cert refresh interval must be less than the minimum lifetime")
		}

		tokenProvider, err = tokenprovider.NewHostTokenProvider(workloadIdentityAudience, identityServerURL, caCertPath, clientCertPath, clientKeyPath, refreshInterval, minCertLifetime)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create host token provider")
		}

	case "kubernetes":
		log.Info().Msg("Using kubernetes mode")
		localCluster, err := kubernetes.NewClusterClient()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create local kubernetes client")
		}

		saCacheTTL := viper.GetDuration("cache.serviceAccountTTL")
		kubeletHost := viper.GetString("kubernetes.kubeletHost")
		kubletCaPath := viper.GetString("kubernetes.kubeletCaPath")

		if err := shared.RegisterRootCAFile(kubletCaPath); err != nil {
			log.Fatal().Err(err).Msg("failed to register kubelet CA file")
		}

		tokenProvider = tokenprovider.NewKubernetesTokenProvider(workloadIdentityAudience, localCluster, kubeletHost, saCacheTTL)

	default:
		log.Fatal().Str("mode", mode).Msg("Invalid mode. Must be either 'kubernetes' or 'host'")
	}

	// Init token cache
	gcInterval := viper.GetDuration("cache.tokenCleanupInterval")
	tokenMinLifetime := viper.GetDuration("cache.tokenMinLifetime")
	knownTokens = NewTokenCache(gcInterval, tokenMinLifetime)
	defer knownTokens.StopGC()

	// Start the server.
	// This is a blocking call
	httpserver.Listen(srv, nil)
}
