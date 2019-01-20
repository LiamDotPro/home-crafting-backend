package main

import (
	"encoding/gob"
	"fmt"
	"github.com/LiamDotPro/Go-Multitenancy/tenants"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/wader/gormstore"
	"os"
	"time"
)

var Connection *gorm.DB
var Store *gormstore.Store

func startDatabaseServices() {

	// Database Connection string
	db, err := gorm.Open(os.Getenv("dialect"), os.Getenv("connectionString"))

	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}

	// Turn logging for the database on.
	db.LogMode(true)

	// Make Master connection available globally.
	Connection = db

	// Now Setup store - Tenant Store
	// Password is passed as byte key method
	Store = gormstore.NewOptions(db, gormstore.Options{
		TableName:       "sessions",
		SkipCreateTable: false,
	}, []byte(os.Getenv("sessionsPassword")))

	// Register session types for consuming in sessions
	gob.Register(HostProfile{})
	gob.Register(CustomerProfile{})

	// Always attempt to migrate changes to the master tenant schema
	if err := migrateTables(); err != nil {
		fmt.Print("There was an error while trying to migrate the tenant tables..")
		os.Exit(1)
	}

	// Makes quit Available
	quit := make(chan struct{})

	// Every hour remove dead sessions.
	go Store.PeriodicCleanup(1*time.Hour, quit)
}

// Migrates the application tables.
func migrateTables() error {

	if err := Connection.AutoMigrate(&tenants.TenantConnectionInformation{}).Error; err != nil {
		return err
	}

	if err := Connection.AutoMigrate(&TenantSubscriptionInformation{}).Error; err != nil {
		return err
	}

	if err := Connection.AutoMigrate(&TenantSubscriptionType{}).Error; err != nil {
		return err
	}

	if err := Connection.AutoMigrate(&MasterUser{}).Error; err != nil {
		return err
	}

	if err := Connection.AutoMigrate(&User{}).Error; err != nil {
		return err
	}

	return nil
}
