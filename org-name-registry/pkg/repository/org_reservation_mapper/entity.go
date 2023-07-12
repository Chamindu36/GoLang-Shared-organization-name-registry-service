import "time"

// Owner object struct
type Owner struct {
	Id        string    `db:"ID"`
	Email     string    `db:"EMAIL"`
	CreatedAt time.Time `db:"CREATED_AT"`
	UpdatedAt time.Time `db:"UPDATED_AT"`
}

// OrgReservation object struct
type OrgReservation struct {
	Uuid             string    `db:"UUID"`
	OrganizationName string    `db:"ORGANIZATION_NAME"`
	OwnerId          string    `db:"OWNER_ID"`
	CreatedAt        time.Time `db:"CREATED_AT"`
	UpdatedAt        time.Time `db:"UPDATED_AT"`
}

// CloudReservation object struct
type CloudReservation struct {
	ReservationId  string    `db:"RESERVATION_ID"`
	ChoreoCloud    bool      `db:"CHOREO_CLOUD"`
	AsgardioCloud  bool      `db:"ASGARDIO_CLOUD"`
	BallerinaCloud bool      `db:"BALLERINA_CLOUD"`
	CreatedAt      time.Time `db:"CREATED_AT"`
	UpdatedAt      time.Time `db:"UPDATED_AT"`
}

// OrgReservationMapping object struct which maps the owner and reservation details and ownership of clouds
type OrgReservationMapping struct {
	OrgReservation OrgReservation
	Owner          Owner
	CloudMapping   CloudReservation
}

// OrgReservationPayload object to extract values from payload
type OrgReservationPayload struct {
	OrganizationName string `json:"orgName"`
	OwnerEmail       string `json:"ownerEmail"`
	CloudService     string `json:"cloudService"`
}

// OrgReservationDBMapping object to extract values from database query
type OrgReservationDBMapping struct {
	Uuid                 string    `db:"UUID"`
	OrganizationName     string    `db:"ORGANIZATION_NAME"`
	ReservationCreatedAt time.Time `db:"CREATED_AT"`
	ReservationUpdatedAt time.Time `db:"UPDATED_AT"`
	OwnerId              string    `db:"OWNER_ID"`
	Email                string    `db:"EMAIL"`
	OwnerCreatedAt       time.Time `db:"CREATED_AT"`
	OwnerUpdatedAt       time.Time `db:"UPDATED_AT"`
	ChoreoCloud          bool      `db:"CHOREO_CLOUD"`
	AsgardioCloud        bool      `db:"ASGARDIO_CLOUD"`
	BallerinaCloud       bool      `db:"BALLERINA_CLOUD"`
}

type OrganizationList []*OrgReservation

// MappingUpdatePayload object to update values from payload
type MappingUpdatePayload struct {
	OrganizationName string `json:"orgName"`
	NewOwnersEmail   string `json:"newEmail"`
}

// ReservationDeletePayload object to delete mapping or reservation
type ReservationDeletePayload struct {
	OrganizationName string `json:"orgName"`
	CloudService     string `json:"cloudService"`
}

// List of organization reservations to return
type OrgReservationNameList []string
