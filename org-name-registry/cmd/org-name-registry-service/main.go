package main

import (
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/clients"
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/oidc"
	orgreservation "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/repository/org_reservation_mapper"
	server "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/server"
	service "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/services/org_reservation_mapper"
	"github.com/Chamindu36/organization-name-registry-service/pkg/config"
	"github.com/Chamindu36/organization-name-registry-service/pkg/logging"
	mysqlstore "github.com/Chamindu36/organization-name-registry-service/pkg/store/sql/mysql"
	"go.uber.org/zap"
	"log"
)

const (
	componentName             = "org-name-registry-service"
	dbConfigFilePath   string = "/home/chamindu/Desktop/Choreo/organization-name-registry-service/org-name-registry/kustomize/db_config.yaml"
	tomlConfigPath     string = "/home/chamindu/Desktop/Choreo/organization-name-registry-service/org-name-registry/kustomize/deployment.yaml"
	oidcConfigFilePath string = "/home/chamindu/Desktop/Choreo/organization-name-registry-service/org-name-registry/kustomize/oidc_config.yaml"
)

// main is the main function which initializes services,repositories and extract configs
func main() {

	// get configuration from config files
	cfg1, err := config.ReadFile(dbConfigFilePath)
	if err != nil {
		log.Fatalf("Cannot read config file: %s", err.Error())
	}

	//Extract IDP configurations
	cfg2, err2 := config.ReadFile(oidcConfigFilePath)
	if err2 != nil {
		log.Fatalf("Cannot read config file: %s", err2.Error())
	}
	oidcIdpConfig := &oidc.OidcIdpConfig{}
	err = cfg2.Unmarshal("oidc", oidcIdpConfig)
	if err != nil {
		log.Fatalf("Error loading oidc idp config: %s", err.Error())
	}

	// Validate the provided configuration
	err = oidcIdpConfig.Validate()
	if err != nil {
		log.Fatal(err)
	}

	oidcAuthenticator, err := clients.NewAuthClient(oidc.NewAuthenticator(oidcIdpConfig))
	if err != nil {
		log.Fatal(err)
	}

	//Extract logging middleware configurations
	cfg3, err := config.ReadFile(tomlConfigPath)
	if err != nil {
		log.Fatalf("Cannot read config file: %s", err.Error())
	}
	middlewareCfg := server.MiddlewareConfig{}
	cfg3.MustUnmarshal("middleware", &middlewareCfg)

	//Extract DB configurations
	dbConfig := &mysqlstore.Config{}
	cfg1.MustUnmarshal("db", dbConfig)
	mysqlDb := mysqlstore.MustOpen(dbConfig)

	// Initialize repository and service
	orgRepository := orgreservation.NewSql(mysqlDb)
	orgService := service.NewOrgReservationService(orgRepository)
	opt := server.Options{
		Port:             8081,
		MiddlewareConfig: middlewareCfg,
	}

	// Initialize logger
	var loggerConfig *zap.Config
	if cfg3.IsSet("logger") {
		loggerConfig = &zap.Config{}
		cfg3.MustUnmarshal("logger", loggerConfig)
	}

	logger, err := logging.NewNamedFromConfig(loggerConfig, componentName)
	if err != nil {
		log.Fatalf("Error building logger: %v", err.Error())
	}
	defer logger.Sync()

	// Initialize mux server with routes
	httpRouter := server.NewMuxRouter(oidcAuthenticator, server.Build(orgService), opt, logger)
	logger.Info("Routes built successfully")

	//Start and serve mux server
	httpRouter.Serve(opt)
}
