package handlers

import (
	"encoding/json"
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/errors"
	repo "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/repository/org_reservation_mapper"
	service "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/services/org_reservation_mapper"
	"net/http"
)

// organization reservation handler struct
type orgReservationHandler struct {
	service service.Interface
}

// Handler interface
type ReservationHandlerInterface interface {
	AddReservation(response http.ResponseWriter, request *http.Request)
	CheckReservation(response http.ResponseWriter, request *http.Request)
}

var (
	reservationService *service.Service
)

// NewOrgReservationHandler method will initialize the OrgReservation interface
// @param service Initialized service interface
// @return *orgReservationHandler
func NewOrgReservationHandler(service *service.Service) *orgReservationHandler {
	reservationService = service
	return &orgReservationHandler{
		service: reservationService,
	}
}

// AddReservation will add a new organization name reservation after checking validations
// @param response response object
// @param request request object
func (*orgReservationHandler) AddReservation(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	var payload repo.OrgReservationPayload
	error := json.NewDecoder(request.Body).Decode(&payload)
	if error != nil {
		errors.JsonError(response, error.Error(), http.StatusBadRequest)
		return
	}

	result, err := reservationService.CreateNewOrgReservation(ctx, &payload)
	if err != nil {
		errors.JsonError(response, err.Error(), http.StatusConflict)
		return

	}
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(result)
}

// CheckReservation will retrieves org names which are reserved under the given email or org name which is checked if existed
// @param response response object
// @param request request object
func (*orgReservationHandler) CheckReservation(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	reservationName := request.FormValue("name")
	if reservationName == "" {
		errors.WriteHttp(response, errors.NewInvalidRequest("Missing query parameter: name", nil))
		return
	}

	result, err := reservationService.SearchReservationByName(ctx, reservationName)
	if err != nil {
		errors.JsonError(response, err.Error(), http.StatusNotFound)
		return

	}
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(result)
}
