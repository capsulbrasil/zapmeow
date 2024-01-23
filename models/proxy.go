package models

import "gorm.io/gorm"

type Proxy struct {
	gorm.Model
	Scheme  string
	Ip      string
	Port    string
	Ranking uint
	Using   uint
}
