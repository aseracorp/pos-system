package main

import (
	"time"

	"gorm.io/gorm"
)

type ProductType struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;" json:"id"`
	Name      string    `gorm:"uniqueIndex; index; type:mediumtext not null" json:"name"`
	Title     string    `gorm:"type:mediumtext" json:"title"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli" json:"-"`
}

type Product struct {
	gorm.Model
	ID            uint64    `gorm:"primaryKey; autoIncrement; not_null;"`
	Name          string    `gorm:"uniqueIndex; index; type:mediumtext not null"`
	Price         float64   `gorm:"not_null;"`
	Discontinued  uint8     `gorm:"not_null; default:0"`
	ProductTypeID uint64    `gorm:"not_null; default:0"`
	SoldOut       uint8     `gorm:"not_null; default:0"`
	CreatedAt     time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time `gorm:"autoCreateTime:milli"`
	// Type          string    `gorm:"not_null; default:'food'"`
	ProductType ProductType
}

type Order struct {
	gorm.Model
	ID            uint64         `gorm:"primaryKey; autoIncrement; not_null;"`
	Cancelled     uint8          `gorm:"not_null; default 0"`
	CreatedAt     time.Time      `gorm:"autoCreateTime:milli"`
	OrderProducts []OrderProduct `gorm:"foreignKey:OrderID;references:ID;"`
	Products      []Product      `gorm:"-"`
}

type OrderProduct struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"`
	ProductID uint64    `gorm:"not_null;"` // index;
	OrderID   uint64    `gorm:"not_null;"` // index;
	Fulfilled uint8     `gorm:"not_null; default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	Product   Product
}

type User struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"`
	Username  string    `gorm:"uniqueIndex; index; not_null;"`
	Password  string    `gorm:"not_null;"`
	StationID uint8     `gorm:"not_null; default 0"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli"`
	Station   Station   `gorm:"one2one:stations"`
}

type Station struct {
	gorm.Model
	ID              uint64           `gorm:"primaryKey; autoIncrement; not_null;" json:"id"`
	Name            string           `gorm:"uniqueIndex; index; not_null;" json:"name"`
	CreatedAt       time.Time        `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoCreateTime:milli" json:"updated_at"`
	StationProducts []StationProduct `gorm:"one2many:station_products"`
}

type StationProduct struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;" json:"id"`
	StationID uint64    `gorm:"index; not_null;" json:"station_id"`
	ProductID uint64    `gorm:"index; not_null;" json:"product_id"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli" json:"updated_at"`
	Product   Product   `gorm:"one2one:products"`
	Station   Station   `gorm:"one2one:stations"`
}

// https://gorm.io/docs/has_one.html#Override-Foreign-Key

const (
	RoleSales   uint8 = 0
	RoleStation uint8 = 1
)
