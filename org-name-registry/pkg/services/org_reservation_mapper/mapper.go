package org_reservation_mapper

import (
	"context"
	errors "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/errors"
	entity "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/repository/org_reservation_mapper"
	"github.com/Chamindu36/organization-name-registry-service/pkg/logging"
	"github.com/google/uuid"
	"regexp"
)

type Service struct {
	orgNameMapper entity.Interface
}

var (
	repo entity.Interface
)

const (
	ChoreoKey    string = "CH"
	AsgardioKey  string = "AG"
	BallerinaKey string = "BL"
)

// NewOrgReservationService method will initialize the Service interface
// @param service repository to used with the service
// @return *Service
func NewOrgReservationService(repository entity.Interface) *Service {
	repo = repository
	return &Service{
		repository,
	}
}

// CreateNewOrgReservation method is used to create a reservation after the checks
// If the owner tries with the same org name, the existing name will be returned, but if someone tries with an existing name error will be returned
// @param context context of the request
// @param payload payload of the request
// @param ownerEmail owner's_mail to be used in the query
// @return *OrgReservationMapping if the entry creation is permitted or the owner checks reservation which is already owned
// @return error if any error occurred or reservation is already taken
func (s Service) CreateNewOrgReservation(context context.Context, payload *entity.OrgReservationPayload) (*entity.OrgReservationMapping, error) {

	logging.NewDefaultLogger().Info("CreateNewOrgReservation method reached")
	// validate the email
	if !validateEmail(payload.OwnerEmail) {
		logging.NewDefaultLogger().Warnf("Email is invalid")
		return nil, errors.Newf(errors.Error_INVALID_REQUEST, nil, "Email is invalid")
	}

	// validate the provided orgName
	if !validateOrgName(payload.OrganizationName) {
		logging.NewDefaultLogger().Warnf("Organization reservation should contain only alpha numeric characters")
		return nil, errors.Newf(errors.Error_INVALID_REQUEST, nil, "Organization reservation should contain only alpha numeric characters")
	}

	// check whether owner has a reservation with same org_name
	ownReservation, owned, error := s.GetOwnReservationWithNameAndMail(context, payload.OrganizationName, payload.OwnerEmail)
	if error != nil {
		return nil, error
	}

	var cloudMappingEntry entity.CloudReservation
	if ownReservation != nil {
		cloudMappingEntry = ownReservation.CloudMapping
	}
	// Check the cloud service that reservation was initiated
	newCloudMapping := cloudMappingEntry
	switch payload.CloudService {
	case ChoreoKey:
		newCloudMapping.ChoreoCloud = true
	case AsgardioKey:
		newCloudMapping.AsgardioCloud = true
	case BallerinaKey:
		newCloudMapping.BallerinaCloud = true
	default:
	}
	if owned {
		// If nothing is changed
		if newCloudMapping == cloudMappingEntry {
			return ownReservation, nil
		} else {
			// Update reservation with new cloud mapping
			err := repo.UpdateCloudMappingById(context, &newCloudMapping)
			if err != nil {
				return nil, err
			}
			ownReservation.CloudMapping = newCloudMapping
			return ownReservation, nil
		}
	}

	_, err := s.GetOrgReservationByOrgName(context, payload.OrganizationName)
	// if there is an exiting reservation with that name and the owner is different
	if !errors.Is(err, errors.Error_DUPLICATE_RESERVATION) {

		var ownerObj entity.Owner
		// check for existing user
		found, errr := repo.GetOwnerByEmail(context, payload.OwnerEmail)
		if errr != nil {
			return nil, errr
		}
		if found == nil { // Create a new user if not found one
			ownerId := uuid.NewString()
			ownerObj.Id = ownerId
			ownerObj.Email = payload.OwnerEmail
			_, err := repo.InsertNewOwner(context, &ownerObj)
			if err != nil {
				logging.NewDefaultLogger().Errorf(err.Error())
				return nil, err
			}

		} else {
			ownerObj = *found
		}

		//create org reservation entry
		var reservationObj entity.OrgReservation
		reservationObj.Uuid = uuid.NewString()
		reservationObj.OrganizationName = payload.OrganizationName
		reservationObj.OwnerId = ownerObj.Id

		_, err := repo.InsertNewOrgReservation(context, &reservationObj)
		if err != nil {
			logging.NewDefaultLogger().Errorf(err.Error())
			return nil, err
		}

		// create cloud mapping entry

		newCloudMapping.ReservationId = reservationObj.Uuid
		_, err1 := repo.InsertNewCloudMappingEntry(context, &newCloudMapping)
		if err1 != nil {
			logging.NewDefaultLogger().Errorf(err1.Error())
			return nil, err1
		}

		logging.NewDefaultLogger().Info("Database entries created successfully")
		return &entity.OrgReservationMapping{
			Owner:          ownerObj,
			OrgReservation: reservationObj,
			CloudMapping:   newCloudMapping,
		}, nil
	} else if err != nil {
		return nil, err
	}
	return nil, err
}

// GetOrgReservationByOrgName method is used to retrieve an existing org reservation with a given owner
// @param context context of the request
// @param orgName org_name to be used as the reservation
// @param ownerEmail owner's_mail to be used in the query
// @return *OrgReservation if a reservation with the given name found
// @return error if any error occurred or no reservation entries found with given name
func (s Service) GetOrgReservationByOrgName(context context.Context, orgName string) (*entity.OrgReservation, error) {
	logging.NewDefaultLogger().Info("GetOrgReservationByOrgName method reached")
	found, err := repo.GetOrgReservationByName(context, orgName)
	if err != nil {
		return nil, err
	}
	// If reservation is existed with the provided org name
	if found != nil {
		logging.FromContext(context).With("orgName", orgName).
			Warnf("The org name is reserved already")
		return nil, errors.Newf(errors.Error_DUPLICATE_RESERVATION, err, "%s is already reserved by another user", orgName)
	}
	return found, nil
}

// GetOwnReservationWithNameAndMail method is used to retrieve own organization reservation using owner's mail and org name
// @param context context of the request
// @param orgName org_name to be used as the reservation
// @param ownerEmail owner's email address
// @return *OrgReservation if a reservation with the given name found for the owner
// @return error if any error occurred
func (s Service) GetOwnReservationWithNameAndMail(context context.Context, orgName string, ownerEmail string) (*entity.OrgReservationMapping, bool, error) {
	logging.NewDefaultLogger().Info("GetOwnReservationWithNameAndMail method reached")
	// Validate the email
	if !validateEmail(ownerEmail) {
		logging.NewDefaultLogger().Warnf("Email is invalid")
		return nil, false, errors.Newf(errors.Error_INVALID_REQUEST, nil, "Email is invalid")
	}
	found, err := repo.GetOrgReservationByNameAndEmail(context, orgName, ownerEmail)
	if err != nil {
		return nil, false, err
	}

	//If the owner has reservation from the given org_name
	if found != nil {
		return found, true, nil
	}
	return nil, false, nil
}

// UpdateOrgMapping method is used to change the ownership of org name reservation
// If the logged in user does not have a reservation under provide name, this will throw and error, but if there is a reservation owner can change the ownership
// If the new owner is already in the table the ownership will be changed to that user, if the new user is not in the registry new owner will created and assign the ownership
// @param context context of the request
// @param currentMail current owner's mail
// @param payload payload of the update request
// @return *OrgReservationMapping if the entry update is permitted
// @return error if any error occurred or owner has no reservation under provide name
func (s Service) UpdateOrgMapping(context context.Context, currentMail string, payload *entity.MappingUpdatePayload) (*entity.OrgReservationMapping, error) {

	logging.NewDefaultLogger().Info("UpdateOrgMapping method reached")
	// validate the provided orgName
	if !validateOrgName(payload.OrganizationName) {
		logging.NewDefaultLogger().Warnf("Organization reservation should contain only alpha numeric characters")
		return nil, errors.Newf(errors.Error_INVALID_REQUEST, nil, "Organization reservation should contain only alpha numeric characters")
	}

	// Validate the email
	if !validateEmail(payload.NewOwnersEmail) {
		logging.NewDefaultLogger().Warnf("Email is invalid")
		return nil, errors.Newf(errors.Error_INVALID_REQUEST, nil, "Email is invalid")
	}

	//check whether the owner has an existing reservation mapping
	ownReservation, recordFound, errr := s.GetOwnReservationWithNameAndMail(context, payload.OrganizationName, currentMail)

	if errr != nil {
		return nil, errr
	}
	// If reservation with the given name and ownership is not existed
	if !recordFound {
		return nil, errors.Newf(errors.Error_NOT_FOUND, errr, "Cannot find a reservation mapping with %s reservation name", payload.OrganizationName)
	}

	found, err := repo.GetOwnerByEmail(context, payload.NewOwnersEmail)
	if err != nil {
		return nil, err
	}

	var newOwner entity.Owner
	if found != nil { // if new owner is already in the registry

		//Get the owner with new email from DB
		result, err := repo.GetOwnerByEmail(context, payload.NewOwnersEmail)
		if err != nil {
			return nil, err
		}
		newOwner = *result
	} else { // if new owner is not in the registry

		//Create new owner object with new email
		ownerEntry := &entity.Owner{
			Id:    uuid.NewString(),
			Email: payload.NewOwnersEmail,
		}
		result, error := repo.InsertNewOwner(context, ownerEntry)
		if error != nil {
			return nil, error
		}
		newOwner = *result
	}
	//Update org reservation with new ownership
	err1 := repo.UpdateOrgReservationById(context, ownReservation.OrgReservation.Uuid, newOwner.Id)
	if err1 != nil {
		return nil, err1
	}
	// Get updated reservation from DB
	updatedEntry, err2 := repo.GetOrgReservationByName(context, payload.OrganizationName)
	if err2 != nil {
		return nil, err2
	}
	return &entity.OrgReservationMapping{
		OrgReservation: *updatedEntry,
		Owner:          newOwner,
		CloudMapping:   ownReservation.CloudMapping,
	}, nil
}

// SearchReservationByName method is used to retrieve an existing org reservation when name is provided
// @param context context of the request
// @param orgName org_name of the reservation to check
// @return string if a reservation with the given name found
// @return error if any error occurred or no reservation entries found with given name
func (s Service) SearchReservationByName(context context.Context, orgName string) (string, error) {
	logging.NewDefaultLogger().Info("SearchReservationByName method reached")
	found, err := repo.GetOrgReservationByName(context, orgName)
	if err != nil {
		return "", err
	}
	if found != nil {
		return found.OrganizationName, nil
	} else {
		logging.FromContext(context).With("orgName", orgName).
			Warnf("The org name is not found in the registry")
		return "", errors.Newf(errors.Error_NOT_FOUND, err, "%s is not found in the registry", orgName)
	}
}

// GetReservationsOfOwner method is used to retrieve an existing org reservation with a given owner
// @param context context of the request
// @param ownerEmail owner's_mail to be used in the query
// @return []string list of reservations owned by owner
// @return error if any error occurred or no reservation entries found with given name
func (s Service) GetReservationsOfOwner(context context.Context, email string) ([]string, error) {
	logging.NewDefaultLogger().Info("GetReservationsOfOwner method reached")
	found, err := repo.FindOrgReservationsByEmail(context, email)
	if err != nil {
		return nil, err
	}
	if found != nil {
		return found, nil

	} else {
		logging.NewDefaultLogger().Error("No reservations are not found in the registry")
		return nil, errors.Newf(errors.Error_NOT_FOUND, err, "No reservations are not found in the registry")
	}
}

// DeleteReservationsOfOwner method is used to remove cloud mapping or org reservation of a user
// @param context context of the request
// @param ownerEmail owner's_mail to be used in the query
// @param payload payload of the delete request
// @return error if any error occurred or no reservation entries found with given name
func (s Service) DeleteReservationsOfOwner(context context.Context, email string, payload *entity.ReservationDeletePayload) error {
	logging.NewDefaultLogger().Info("CreateNewOrgReservation method reached")
	// validate the email
	if !validateEmail(email) {
		logging.NewDefaultLogger().Warnf("Email is invalid")
		errors.Newf(errors.Error_INVALID_REQUEST, nil, "Email is invalid")
	}
	// check whether owner has a reservation with same org_name
	ownReservation, owned, error := s.GetOwnReservationWithNameAndMail(context, payload.OrganizationName, email)
	if error != nil {
		return error
	}
	if !owned {
		logging.FromContext(context).With("orgName", payload.OrganizationName).
			Warnf("The org name is not found in the registry with user's ownership")
		return errors.Newf(errors.Error_NOT_FOUND, nil, "%s is not found in the registry with your ownership", payload.OrganizationName)
	}

	// Check the cloud service that reservation was initiated
	newCloudMapping := ownReservation.CloudMapping
	switch payload.CloudService {
	case ChoreoKey:
		newCloudMapping.ChoreoCloud = false
	case AsgardioKey:
		newCloudMapping.AsgardioCloud = false
	case BallerinaKey:
		newCloudMapping.BallerinaCloud = false
	default:
	}
	if newCloudMapping.BallerinaCloud == false && newCloudMapping.ChoreoCloud == false && newCloudMapping.AsgardioCloud == false {
		// When all the cloud reservations are fall remove the cloud mapping and org reservation
		err := repo.DeleteCloudMappingEntryById(context, newCloudMapping.ReservationId)
		if err != nil {
			return err
		}
		// If org name is immutable
		//err1 := repo.DeleteOrgReservationById(context, newCloudMapping.ReservationId)
		//if err1 != nil {
		//	return err1
		//}
		logging.NewDefaultLogger().Info("Database entries deleted successfully")
	} else {
		err := repo.UpdateCloudMappingById(context, &newCloudMapping)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateEmail method will validate the email address format
// @param email email address to be validated
// @return bool true if the email is in correct format, false if the email is an invalid one
func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

// validateOrgName method will validate the email address format
// @param email org_name pattern to be validated
// @return bool true if the org_name is in correct format, false if the org_name is an invalid one
func validateOrgName(orgName string) bool {
	Re := regexp.MustCompile(`^[a-zA-Z0-9_]*$`)
	return Re.MatchString(orgName)
}
