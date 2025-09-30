package user

import "time"

type User struct {
	Id               int64
	Username         string
	Name             string
	Email            string
	DocumentID       string //CMND
	Birthday         time.Time
	Gender           bool
	PermanentAddress string
	PhoneNumber      string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginMethodPassword struct {
	UserName string
	Password string
}
