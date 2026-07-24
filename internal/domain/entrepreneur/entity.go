package entrepreneur

import (
	"time"

	"github.com/google/uuid"
)

type Entrepreneur struct {
	ID                    uuid.UUID
	InnID                 uuid.UUID
	InnName               string
	LegalName             string
	RegistrationAuthority string
	RegistrationDate      string
	RegistrationNumber    string
	LegalForm             string
	IfutCodeID            *uuid.UUID
	IfutCodeName          string
	ActivityType          string
	DbibtCode             int32
	ActivityStatus        bool
	CharterFund           int32
	Founders              string
	Email                 string
	Phone                 string
	MhobtCode             string
	Address               string
	ActivityRegionID      int32
	ActivityRegion        string
	ActivitySubRegion     string
	DirectorName          string
	SqbApiError           *string
	CreatedAt             time.Time
}

type EntrepreneurFilter struct {
	LegalName      *string
	InnName        *string
	ActivityStatus *bool
	DirectorName   *string
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}
