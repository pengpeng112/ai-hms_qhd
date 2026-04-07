package services

import "gorm.io/gorm"

func ensureTables(db *gorm.DB, entities ...interface{}) error {
	if db == nil {
		return nil
	}

	for _, entity := range entities {
		if db.Migrator().HasTable(entity) {
			continue
		}
		if err := db.AutoMigrate(entity); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchema(db *gorm.DB, entities ...interface{}) error {
	if db == nil {
		return nil
	}

	return db.AutoMigrate(entities...)
}
