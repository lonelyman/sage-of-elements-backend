package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/applogger"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, appLogger applogger.Logger) error {
	appLogger.Info("Running database migrations...")

	err := db.AutoMigrate(
		// --- System & Core Game Data ---
		&domain.GameConfig{},
		&domain.Element{},
		&domain.ElementalMatchup{}, // <-- ตารางแพ้ทางธาตุ
		&domain.Mastery{},
		&domain.Effect{},
		&domain.Recipe{},
		&domain.RecipeIngredient{},
		&domain.Spell{},
		&domain.SpellEffect{},

		// --- Player & Character Data ---
		&domain.Player{},
		&domain.PlayerAuth{},
		&domain.Character{},
		&domain.CharacterMastery{},
		&domain.DimensionalSealInventory{},
		&domain.CharacterJournalDiscovery{},
		&domain.Deck{},
		&domain.DeckSlot{},

		// --- Enemy Data ---
		&domain.Enemy{},
		&domain.EnemyAbility{},
		&domain.EnemyAI{},
		&domain.EnemyLoot{},

		// --- PvE Data ---
		&domain.Realm{},
		&domain.Chapter{},
		&domain.Stage{},
		&domain.StageEnemy{},

		// --- ✨⭐️ สิ่งที่หายไป อยู่ตรงนี้! ⭐️✨ ---
		// --- Combat Data ---
		&domain.CombatMatch{},
		&domain.Combatant{},
	)

	if err != nil {
		// ✨ Log Error ที่เกิดขึ้นจริงๆ! ✨
		appLogger.Error("Database migration failed", err)
		return err
	}

	appLogger.Success("Database migration completed.")
	return nil
}
