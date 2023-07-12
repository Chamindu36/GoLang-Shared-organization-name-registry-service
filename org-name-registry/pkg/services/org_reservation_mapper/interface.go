package org_reservation_mapper

//go:generate go run ${PROJECT_DIR}/tools/fake_gen.go

import (
	"context"
	entity "github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/repository/org_reservation_mapper"
)

type Interface interface {
	CreateNewOrgReservation(context context.Context, reservation *entity.OrgReservationPayload) (*entity.OrgReservationMapping, error)
	GetOrgReservationByOrgName(context context.Context, orgName string) (*entity.OrgReservation, error)
	GetOwnReservationWithNameAndMail(context context.Context, orgName string, ownerEmail string) (*entity.OrgReservationMapping, bool, error)
	UpdateOrgMapping(context context.Context, currentMail string, payload *entity.MappingUpdatePayload) (*entity.OrgReservationMapping, error)
	SearchReservationByName(context context.Context, orgName string) (string, error)
	GetReservationsOfOwner(context context.Context, email string) ([]string, error)
	DeleteReservationsOfOwner(context context.Context, email string, payload *entity.ReservationDeletePayload) error
}
