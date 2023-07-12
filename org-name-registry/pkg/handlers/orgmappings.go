package handlers

import (
	"encoding/json"
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/errors"
	repo "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/repository/org_reservation_mapper"
	service "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/services/org_reservation_mapper"
	"github.com/gorilla/mux"
	"net/http"
)

// organization reservation handler struct
type orgMappingsHandler struct {
	service service.Interface
}

// Handler interface
type MappingsHandlerInterface interface {
	UpdateMapping(response http.ResponseWriter, request *http.Request)
	GetOwnReservations(response http.ResponseWriter, request *http.Request)
	DeleteOwnReservations(response http.ResponseWriter, request *http.Request)
}

var (
	mappingService *service.Service
)

// NewOrgReservationHandler method will initialize the OrgReservation interface
// @param service Initialized service interface
// @return *orgReservationHandler
func NewOrgMappingsHandler(service *service.Service) *orgMappingsHandler {
	mappingService = service
	return &orgMappingsHandler{
		service: mappingService,
	}
}

// AddReservation will add a new organization name reservation after checking validations
// @param response response object
// @param request request object
func (*orgMappingsHandler) UpdateMapping(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	params := mux.Vars(request)
	currentMail := params["ownersEmail"]
	var payload repo.MappingUpdatePayload

	// Mapping request payload into a put payload structure
	error := json.NewDecoder(request.Body).Decode(&payload)
	if error != nil {
		errors.JsonError(response, error.Error(), http.StatusBadRequest)
		return
	}

	result, err := mappingService.UpdateOrgMapping(ctx, currentMail, &payload)
	if err != nil {
		errors.JsonError(response, err.Error(), http.StatusBadRequest)
		return

	}
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(result)
}

// GetOwnReservations will retrieves org name reservations which are reserved under the owner
// @param response response object
// @param request request object
func (*orgMappingsHandler) GetOwnReservations(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	params := mux.Vars(request)
	mail := params["ownersEmail"]

	result, err := reservationService.GetReservationsOfOwner(ctx, mail)
	if err != nil {
		errors.JsonError(response, err.Error(), http.StatusNotFound)
		return

	}
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(result)
}

// DeleteOwnReservations will remove org name reservation or reservation mapping of a user
// @param response response object
// @param request request object
func (*orgMappingsHandler) DeleteOwnReservations(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	params := mux.Vars(request)
	mail := params["ownersEmail"]

	var payload repo.ReservationDeletePayload

	// Mapping request payload into a put payload structure
	error := json.NewDecoder(request.Body).Decode(&payload)
	if error != nil {
		errors.JsonError(response, error.Error(), http.StatusBadRequest)
		return
	}
	err := reservationService.DeleteReservationsOfOwner(ctx, mail, &payload)
	if err != nil {
		errors.JsonError(response, err.Error(), http.StatusNotFound)
		return

	}
	response.WriteHeader(http.StatusOK)
}
