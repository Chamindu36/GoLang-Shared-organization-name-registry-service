package org_reservation_mapper

//go:generate go run ${PROJECT_DIR}/tools/fake_gen.go

import (
	"context"
)

type Interface interface {
	InsertNewOrgReservation(ctx context.Context, orgReservationMapping *OrgReservation) (*OrgReservation, error)
	GetOrgReservationByName(ctx context.Context, orgName string) (*OrgReservation, error)
	GetOrgReservationByNameAndEmail(ctx context.Context, orgName string, ownerEmail string) (*OrgReservationMapping, error)
	UpdateOrgReservationByName(ctx context.Context, name, ownerId string) error
	UpdateOrgReservationById(ctx context.Context, uuid, ownerId string) error
	UpdateCloudMappingById(ctx context.Context, cloud *CloudReservation) error
	GetOwnerByEmail(ctx context.Context, currentMail string) (*Owner, error)
	UpdateOwnerById(ctx context.Context, ownerId string, newMail string) error
	InsertNewOwner(ctx context.Context, entry *Owner) (*Owner, error)
	InsertNewCloudMappingEntry(ctx context.Context, cloudMapping *CloudReservation) (*CloudReservation, error)
	FindOrgReservationsByEmail(ctx context.Context, email string) ([]string, error)
	DeleteCloudMappingEntryById(ctx context.Context, id string) error
	DeleteOrgReservationById(ctx context.Context, uuid string) error
}

func OrgReservationMapperRepository() *sqlRepository {
	return &sqlRepository{}
}
