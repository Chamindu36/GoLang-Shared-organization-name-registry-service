package server

import (
	handlers "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/handlers"
	service "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/services/org_reservation_mapper"
	"net/http"
)

type Route struct {
	Path        string
	PathPrefix  string
	Queries     []QueryMatch
	Methods     []string
	HandlerFunc HandlerFunc
}

type QueryMatch struct {
	Name    string
	Pattern string
}

//var (
//	orgService *service.Service
//)

type Routes []Route

func Build(service *service.Service) Routes {
	//Initialize Handlers with services
	orgReservationHandler := handlers.NewOrgReservationHandler(service)
	orgMappingHandler := handlers.NewOrgMappingsHandler(service)
	return Routes{
		// Organization Reservation routes
		Route{Path: "/org-reservations", Methods: []string{http.MethodPost}, HandlerFunc: orgReservationHandler.AddReservation},
		Route{Path: "/org-reservations", Queries: []QueryMatch{{"name", ""}}, Methods: []string{http.MethodGet}, HandlerFunc: orgReservationHandler.CheckReservation},

		// Organization mapping routes
		Route{Path: "/owners/{ownersEmail}/org-mappings/", Methods: []string{http.MethodPut}, HandlerFunc: orgMappingHandler.UpdateMapping},
		Route{Path: "/owners/{ownersEmail}/org-mappings/", Methods: []string{http.MethodGet}, HandlerFunc: orgMappingHandler.GetOwnReservations},
		Route{Path: "/owners/{ownersEmail}/org-mappings/", Methods: []string{http.MethodDelete}, HandlerFunc: orgMappingHandler.DeleteOwnReservations},
	}
}
