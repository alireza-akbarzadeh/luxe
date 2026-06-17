package main

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/database"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

func connectDatabase(cfg *config.Config) *gorm.DB {
	db, err := database.Connect(cfg)
	if err != nil {
		utils.Log.WithError(err).Fatal("failed to connect to database")
	}
	return db
}

func closeDatabase(db *gorm.DB) {
	database.Close(db)
}
