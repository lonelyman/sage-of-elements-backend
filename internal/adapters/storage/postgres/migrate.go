package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/applogger"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, appLogger applogger.Logger) error {
	appLogger.Info("Running database migrations...")

	err := db.AutoMigrate(
		&domain.GameConfig{},
		// Core Game Data
		&domain.Element{},
		&domain.Mastery{},
		&domain.Effect{},
		&domain.Recipe{},
		&domain.RecipeIngredient{},
		&domain.Spell{},
		&domain.SpellEffect{},

		// Player Data
		&domain.Player{},
		&domain.PlayerAuth{},
		&domain.Character{},
		&domain.CharacterMastery{},
		&domain.DimensionalSealInventory{},
		&domain.CharacterJournalDiscovery{},
	)

	if err != nil {
		// ✨ Log Error ที่เกิดขึ้นจริงๆ! ✨
		appLogger.Error("Database migration failed", err)
		return err
	}

	appLogger.Success("Database migration completed.")
	return nil
}
