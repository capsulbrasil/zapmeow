package model

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	User       string
	RawAgent   uint8
	Device     uint16
	Integrator uint16
	Server     string
	QrCode     string
	Status     string
	WasSynced  bool
	InstanceID string
}
