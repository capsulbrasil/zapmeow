package model

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	User       string
	Agent      uint8
	Device     uint8
	Server     string
	AD         bool
	QrCode     string
	Status     string
	WasSynced  bool
	InstanceID string
}
