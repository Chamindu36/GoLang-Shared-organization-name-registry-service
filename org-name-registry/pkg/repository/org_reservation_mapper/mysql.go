package org_reservation_mapper

import (
	"context"
	"database/sql"
	"github.com/Chamindu36/organization-name-registry-service/pkg/logging"
	sqlstore "github.com/Chamindu36/organization-name-registry-service/pkg/store/sql"
)

type sqlRepository struct {
	querier sqlstore.Querier
}

func NewSql(querier sqlstore.Querier) *sqlRepository {
	return &sqlRepository{querier: querier}
}

//Queries to execute on db instance
const InsertNewOwnerQuery = "INSERT INTO OWNER(ID, EMAIL) VALUES(:ID, :EMAIL)"
const InsertNewOrgReservationQuery = "INSERT INTO ORGANIZATION_RESERVATION(UUID, ORGANIZATION_NAME, OWNER_ID)" +
	" VALUES(:UUID, :ORGANIZATION_NAME, :OWNER_ID)"
const InsertNewCloudMappingQuery = "INSERT INTO MAPPED_CLOUDS (RESERVATION_ID, CHOREO_CLOUD, ASGARDIO_CLOUD, " +
	"BALLERINA_CLOUD) VALUES (:RESERVATION_ID, :CHOREO_CLOUD, :ASGARDIO_CLOUD, :BALLERINA_CLOUD)"
const GetOrgReservationByOrgNameQuery = "SELECT UUID,ORGANIZATION_NAME,OWNER_ID FROM ORGANIZATION_RESERVATION WHERE" +
	" ORGANIZATION_NAME = ?"
const GetOrgReservationByOwnerEmailAndOrgNameQuery = "SELECT org.UUID,org.ORGANIZATION_NAME,org.CREATED_AT," +
	"org.UPDATED_AT,org.OWNER_ID,o.EMAIL,c.CHOREO_CLOUD,c.ASGARDIO_CLOUD,c.BALLERINA_CLOUD FROM ORGANIZATION_RESERVATION " +
	"AS org, OWNER AS o, MAPPED_CLOUDS as c WHERE org.OWNER_ID = o.ID AND org.UUID = c.RESERVATION_ID AND " +
	"org.ORGANIZATION_NAME = ? AND o.EMAIL = ?"
const GetOwnerByEmailQuery = "SELECT ID, EMAIL FROM OWNER WHERE EMAIL= ?"
const UpdateOwnerByIdQuery = "UPDATE OWNER SET EMAIL = :EMAIL WHERE ID = :ID"
const UpdateOrgReservationByIdQuery = "UPDATE ORGANIZATION_RESERVATION SET OWNER_ID = :OWNER_ID WHERE UUID = :UUID"
const UpdateOrgReservationByNameQuery = "UPDATE ORGANIZATION_RESERVATION SET OWNER_ID = :OWNER_ID WHERE ORGANIZATION_NAME" +
	" = :ORGANIZATION_NAME"
const UpdateOrgCloudMappingByIdQuery = "UPDATE MAPPED_CLOUDS SET CHOREO_CLOUD = :CHOREO_CLOUD, ASGARDIO_CLOUD = " +
	":ASGARDIO_CLOUD, BALLERINA_CLOUD = :BALLERINA_CLOUD WHERE RESERVATION_ID = :RESERVATION_ID"
const SelectOrgReservationsByEmailQuery = "SELECT org.ORGANIZATION_NAME FROM ORGANIZATION_RESERVATION AS org, " +
	"OWNER AS o WHERE org.OWNER_ID = o.ID AND o.EMAIL = ?"
const DeleteCloudMappingByIdQuery = "DELETE FROM MAPPED_CLOUDS WHERE RESERVATION_ID = :RESERVATION_ID"
const DeleteOrgReservationByIdQuery = "DELETE FROM ORGANIZATION_RESERVATION WHERE UUID = :UUID"

// InsertNewOrgReservation method is used to enter a new organization reservation entry with owner's details to database
// @param ctx context of the request
// @param OrgReservation organization reservation object with entry details
// @return *OrgReservation if db entry is entered successfully
// @return error if any error occurred
func (s sqlRepository) InsertNewOrgReservation(ctx context.Context, orgReservation *OrgReservation) (*OrgReservation, error) {

	logging.NewDefaultLogger().Info("InsertNewOrgReservation method reached")
	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(InsertNewOrgReservationQuery, orgReservation)
	if err != nil {
		return &OrgReservation{}, err
	}

	saved := *orgReservation
	return &saved, err
}

// GetOrgReservationByName method is used to retrieves the reservation entries when a org_name is provided
// @param ctx context of the request
// @param orgName org_nam to be used in the query
// @return *OrgReservation if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) GetOrgReservationByName(ctx context.Context, orgName string) (*OrgReservation, error) {
	logging.NewDefaultLogger().Info("GetOrgReservationByName method reached")
	found := &OrgReservation{}
	err := sqlstore.FromContext(ctx, s.querier).QueryRowx(GetOrgReservationByOrgNameQuery, orgName).StructScan(found)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return found, nil
	}
}

// GetOrgReservationByNameAndEmail method is used to retrieves the reservation entries when a org_name and owner's_mail are provided
// @param ctx context of the request
// @param orgName org_nam to be used in the query
// @param ownerEmail owner's_mail to be used in the query
// @return *OrgReservationMapping if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) GetOrgReservationByNameAndEmail(ctx context.Context, orgName string, ownerEmail string) (*OrgReservationMapping, error) {
	logging.NewDefaultLogger().Info("GetOrgReservationByNameAndEmail method reached")
	found := &OrgReservationDBMapping{}
	err := sqlstore.FromContext(ctx, s.querier).QueryRowx(GetOrgReservationByOwnerEmailAndOrgNameQuery, orgName, ownerEmail).StructScan(found)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		// Mapping from DB to payload object
		record := &OrgReservationMapping{
			OrgReservation: OrgReservation{
				Uuid:             found.Uuid,
				OrganizationName: found.OrganizationName,
				OwnerId:          found.OwnerId,
				CreatedAt:        found.ReservationCreatedAt,
				UpdatedAt:        found.ReservationUpdatedAt,
			},
			Owner: Owner{
				Id:        found.OwnerId,
				Email:     found.Email,
				CreatedAt: found.OwnerCreatedAt,
				UpdatedAt: found.OwnerUpdatedAt,
			},
			CloudMapping: CloudReservation{
				ReservationId:  found.Uuid,
				ChoreoCloud:    found.ChoreoCloud,
				AsgardioCloud:  found.AsgardioCloud,
				BallerinaCloud: found.BallerinaCloud,
			},
		}
		return record, nil
	}
}

// GetOwnerByEmail method is used to retrieves the owner entries when the owner's_mail is provided
// @param ctx context of the request
// @param currentMail owner's_mail to be used in the query
// @return *Owner if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) GetOwnerByEmail(ctx context.Context, currentMail string) (*Owner, error) {
	logging.NewDefaultLogger().Info("GetOwnerByEmail method reached")
	found := &Owner{}
	err := sqlstore.FromContext(ctx, s.querier).QueryRowx(GetOwnerByEmailQuery, currentMail).StructScan(found)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return found, nil
	}
}

// UpdateOwnerById method is used to update the owner entries when the owner's id is provided
// @param ctx context of the request
// @param ownerId Id of the existing owner entry
// @param newMail new email of the owner
// @return error if any error occurred
func (s sqlRepository) UpdateOwnerById(ctx context.Context, ownerId, newMail string) error {
	logging.NewDefaultLogger().Info("UpdateOwnerById method reached")

	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(UpdateOwnerByIdQuery, map[string]interface{}{
		"EMAIL": newMail,
		"ID":    ownerId,
	})
	return err
}

// InsertNewOwner method is used to create a new owner
// @param ctx context of the request
// @param entry owner entry to be added as a new owner
// @return *Owner if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) InsertNewOwner(ctx context.Context, entry *Owner) (*Owner, error) {
	logging.NewDefaultLogger().Info("InsertNewOwner method reached")
	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(InsertNewOwnerQuery, entry)
	if err != nil {
		return &Owner{}, err
	}
	saved := *entry
	return &saved, err
}

// UpdateOrgReservationById method is used to update the org name reservation entries when the uuid of the reservation is provided
// @param ctx context of the request
// @param uuid unique id of the existing reservation entry
// @param ownerId Id of the new owner whom to be pass the ownership
// @return error if any error occurred
func (s sqlRepository) UpdateOrgReservationById(ctx context.Context, uuid, ownerId string) error {
	logging.NewDefaultLogger().Info("UpdateOrgReservationById method reached")

	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(UpdateOrgReservationByIdQuery, map[string]interface{}{
		"OWNER_ID": ownerId,
		"UUID":     uuid,
	})
	return err
}

// UpdateOrgReservationByName method is used to update the org name reservation entries when the org name of the reservation is provided
// @param ctx context of the request
// @param name org name of the existing reservation entry
// @param ownerId Id of the new owner whom to be pass the ownership
// @return error if any error occurred
func (s sqlRepository) UpdateOrgReservationByName(ctx context.Context, name, ownerId string) error {
	logging.NewDefaultLogger().Info("UpdateOrgReservationById method reached")

	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(UpdateOrgReservationByNameQuery, map[string]interface{}{
		"OWNER_ID":          ownerId,
		"ORGANIZATION_NAME": name,
	})
	return err
}

// FindOrgReservationsByEmail method is used to get all the org name reservations of an owner when the owner's email is provided
// @param ctx context of the request
// @param email owner's_mail to be used in the query
//	@return org names list reserved by the user
// @return error if any error occurred
func (s sqlRepository) FindOrgReservationsByEmail(ctx context.Context, email string) ([]string, error) {
	logging.NewDefaultLogger().Info("FindOrgReservationsByEmail method reached")
	var reservationNames []string
	err := sqlstore.FromContext(ctx, s.querier).Select(&reservationNames, SelectOrgReservationsByEmailQuery, email)
	if err != nil {
		return nil, err
	}
	return reservationNames, nil
}

// UpdateCloudMappingById method is used to update the cloud mapping when new cloud service reserved an existing org_name by the owner
// @param ctx context of the request
// @param cloud cloud reservation object to be used in update query
// @return error if any error occurred
func (s sqlRepository) UpdateCloudMappingById(ctx context.Context, cloud *CloudReservation) error {
	logging.NewDefaultLogger().Info("UpdateCloudMappingById method reached")

	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(UpdateOrgCloudMappingByIdQuery, map[string]interface{}{
		"RESERVATION_ID":  cloud.ReservationId,
		"CHOREO_CLOUD":    cloud.ChoreoCloud,
		"ASGARDIO_CLOUD":  cloud.AsgardioCloud,
		"BALLERINA_CLOUD": cloud.BallerinaCloud,
	})
	return err
}

// InsertNewCloudMappingEntry method is used to create a cloud mapping entry record for a given org_name reservation
// @param ctx context of the request
// @param cloudMapping mapping entry to be added
// @return *CloudReservation if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) InsertNewCloudMappingEntry(ctx context.Context, cloudMapping *CloudReservation) (*CloudReservation, error) {
	logging.NewDefaultLogger().Info("InsertNewOwner method reached")
	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(InsertNewCloudMappingQuery, cloudMapping)
	if err != nil {
		return &CloudReservation{}, err
	}
	saved := *cloudMapping
	return &saved, err
}

// DeleteCloudMappingEntryById method is used to delete a cloud mapping entry record when reservation id is provided
// @param ctx context of the request
// @param id reservation id of the mapping to be deleted
// @return *CloudReservation if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) DeleteCloudMappingEntryById(ctx context.Context, id string) error {
	logging.NewDefaultLogger().Info("DeleteCloudMappingEntryById method reached")
	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(DeleteCloudMappingByIdQuery, map[string]interface{}{
		"RESERVATION_ID": id,
	})

	return err
}

// DeleteCloudMappingEntryById method is used to delete a cloud mapping entry record when reservation id is provided
// @param ctx context of the request
// @param id reservation id of the mapping to be deleted
// @return *CloudReservation if db query is executed successfully
// @return error if any error occurred
func (s sqlRepository) DeleteOrgReservationById(ctx context.Context, uuid string) error {
	logging.NewDefaultLogger().Info("DeleteCloudMappingEntryById method reached")
	_, err := sqlstore.FromContext(ctx, s.querier).NamedExec(DeleteOrgReservationByIdQuery, map[string]interface{}{
		"UUID": uuid,
	})

	return err
}

var _ Interface = (*sqlRepository)(nil)
