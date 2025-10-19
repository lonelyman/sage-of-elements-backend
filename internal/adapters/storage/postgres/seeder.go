package postgres

import (
	"log"
	"sage-of-elements-backend/internal/domain"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seed ทำหน้าที่เติมและอัปเดตข้อมูล Master Data ทั้งหมด
func Seed(db *gorm.DB) error {
	log.Println("Database seeding process started...")
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedMasteries(tx); err != nil {
			return err
		}
		if err := seedElements(tx); err != nil {
			return err
		}
		if err := seedEffects(tx); err != nil {
			return err
		} // ⭐️ ต้องมาก่อน Spells และ Enemies
		if err := seedRecipes(tx); err != nil {
			return err
		}
		if err := seedSpells(tx); err != nil {
			return err
		} // ⭐️ ต้องมาก่อน Seeder ที่ใช้ SpellEffect
		if err := seedGameConfigs(tx); err != nil {
			return err
		}
		if err := seedEnemies(tx); err != nil {
			return err
		}
		if err := seedElementalMatchups(tx); err != nil {
			return err
		}
		if err := seedPveContent(tx); err != nil {
			return err
		} // ⭐️ เพิ่มเข้ามา

		log.Println("✅ Database seeding process finished successfully.")
		return nil
	})
}

func seedMasteries(tx *gorm.DB) error {

	log.Println("Seeding/Updating masteries...")
	masteries := []domain.Mastery{
		{ID: 1, Name: "Force", DisplayNames: datatypes.JSONMap{"en": "Force", "th": "ศาสตร์โจมตี"}, Descriptions: datatypes.JSONMap{"en": "The art of destruction.", "th": "ศาสตร์ที่มุ่งเน้นการทำลายล้าง"}},
		{ID: 2, Name: "Resilience", DisplayNames: datatypes.JSONMap{"en": "Resilience", "th": "ศาสตร์ป้องกัน"}, Descriptions: datatypes.JSONMap{"en": "The art of endurance.", "th": "ศาสตร์ที่มุ่งเน้นการอดทนและปกป้อง"}},
		{ID: 3, Name: "Efficacy", DisplayNames: datatypes.JSONMap{"en": "Efficacy", "th": "ศาสตร์เสริมพลัง"}, Descriptions: datatypes.JSONMap{"en": "The art of enhancement.", "th": "ศาสตร์ที่มุ่งเน้นการเสริมความสามารถ"}},
		{ID: 4, Name: "Command", DisplayNames: datatypes.JSONMap{"en": "Command", "th": "ศาสตร์สนับสนุน"}, Descriptions: datatypes.JSONMap{"en": "The art of control.", "th": "ศาสตร์ที่มุ่งเน้นการควบคุมและก่อกวน"}},
	}
	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&masteries).Error; err != nil {
		return err
	}
	return nil
}

func seedElements(tx *gorm.DB) error {
	log.Println("Seeding/Updating elements...")
	elements := []domain.Element{
		{ID: 1, Name: "Solidity", DisplayNames: datatypes.JSONMap{"en": "Solidity", "th": "สถิตยภาพ"}, Tier: 0},
		{ID: 2, Name: "Liquidity", DisplayNames: datatypes.JSONMap{"en": "Liquidity", "th": "สภาพคล่อง"}, Tier: 0},
		{ID: 3, Name: "Tempo", DisplayNames: datatypes.JSONMap{"en": "Tempo", "th": "จังหวะ"}, Tier: 0},
		{ID: 4, Name: "Potency", DisplayNames: datatypes.JSONMap{"en": "Potency", "th": "พลังงาน"}, Tier: 0},
		{ID: 5, Name: "Viscosity", DisplayNames: datatypes.JSONMap{"en": "Viscosity", "th": "หนืด"}, Tier: 1},
		{ID: 6, Name: "Obscurity", DisplayNames: datatypes.JSONMap{"en": "Obscurity", "th": "คลุมเครือ"}, Tier: 1},
		{ID: 7, Name: "Magma", DisplayNames: datatypes.JSONMap{"en": "Magma", "th": "แม็กม่า"}, Tier: 1},
		{ID: 8, Name: "Ionization", DisplayNames: datatypes.JSONMap{"en": "Ionization", "th": "ไอออน"}, Tier: 1},
		{ID: 9, Name: "Reactivity", DisplayNames: datatypes.JSONMap{"en": "Reactivity", "th": "กรด"}, Tier: 1},
		{ID: 10, Name: "Volatility", DisplayNames: datatypes.JSONMap{"en": "Volatility", "th": "ผันผวน"}, Tier: 1},
		{ID: 11, Name: "Adamantite", DisplayNames: datatypes.JSONMap{"en": "Adamantite", "th": "อาดามันไทต์"}, Tier: 1},
		{ID: 12, Name: "Elixir", DisplayNames: datatypes.JSONMap{"en": "Elixir", "th": "ยาอายุวัฒนะ"}, Tier: 1},
		{ID: 13, Name: "Aether", DisplayNames: datatypes.JSONMap{"en": "Aether", "th": "อีเธอร์"}, Tier: 1},
		{ID: 14, Name: "Sunfire", DisplayNames: datatypes.JSONMap{"en": "Sunfire", "th": "แก่นสุริยะ"}, Tier: 1},
		{ID: 15, Name: "Chaos", DisplayNames: datatypes.JSONMap{"en": "Chaos", "th": "ความโกลาหล"}, Tier: 1},
	}
	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&elements).Error; err != nil {
		return err
	}
	return nil
}

func seedEffects(tx *gorm.DB) error {
	log.Println("Seeding/Updating effects...")
	effects := []domain.Effect{
		// Basic Actions (1-99)
		{ID: 1, Name: "DAMAGE", Type: domain.EffectTypeDamage},
		{ID: 2, Name: "SHIELD", Type: domain.EffectTypeShield},
		{ID: 3, Name: "HEAL", Type: domain.EffectTypeHeal},
		{ID: 4, Name: "TRUE_DAMAGE", Type: domain.EffectTypeTrueDamage},
		{ID: 5, Name: "DRAIN_MP", Type: domain.EffectTypeResource},
		{ID: 6, Name: "GAIN_AP", Type: domain.EffectTypeResource},
		{ID: 7, Name: "CLEANSE", Type: domain.EffectTypeUtility},
		// Buffs (100-199)
		{ID: 100, Name: "BUFF_HP_REGEN", Type: domain.EffectTypeBuff},
		{ID: 101, Name: "BUFF_MP_REGEN", Type: domain.EffectTypeBuff},
		{ID: 102, Name: "BUFF_EVASION", Type: domain.EffectTypeBuff},
		{ID: 103, Name: "BUFF_DMG_UP", Type: domain.EffectTypeBuff},
		{ID: 104, Name: "BUFF_RETALIATION", Type: domain.EffectTypeBuff},
		{ID: 105, Name: "BUFF_MAX_HP", Type: domain.EffectTypeBuff},
		{ID: 106, Name: "BUFF_CC_RESIST", Type: domain.EffectTypeBuff},
		{ID: 108, Name: "BUFF_PENETRATION", Type: domain.EffectTypeBuff},
		{ID: 110, Name: "BUFF_DEFENSE_UP", Type: domain.EffectTypeBuff}, // ⭐️ เพิ่ม Effect ที่ขาดไป!
		// Synergy Buffs (200-299)
		{ID: 200, Name: "SYNERGY_GRANT_STANCE_S", Type: domain.EffectTypeSynergyBuff},
		{ID: 201, Name: "SYNERGY_GRANT_FLOW_L", Type: domain.EffectTypeSynergyBuff},
		{ID: 202, Name: "SYNERGY_GRANT_TEMPO_G", Type: domain.EffectTypeSynergyBuff},
		{ID: 203, Name: "SYNERGY_GRANT_OVERCHARGE_P", Type: domain.EffectTypeSynergyBuff},
		// Debuffs (300-399)
		{ID: 300, Name: "DEBUFF_REDUCE_ARMOR", Type: domain.EffectTypeDebuff},
		{ID: 301, Name: "DEBUFF_SLOW", Type: domain.EffectTypeDebuffCC},
		{ID: 302, Name: "DEBUFF_VULNERABLE", Type: domain.EffectTypeDebuff},
		{ID: 303, Name: "DEBUFF_ROOT", Type: domain.EffectTypeDebuffCC},
		{ID: 304, Name: "DEBUFF_AP_DRAIN", Type: domain.EffectTypeDebuff},
		{ID: 305, Name: "DEBUFF_STUN", Type: domain.EffectTypeDebuffHardCC},
		{ID: 306, Name: "DEBUFF_IGNITE", Type: domain.EffectTypeDebuffDOT},
		{ID: 308, Name: "DEBUFF_CORROSION", Type: domain.EffectTypeDebuffDOT},
	}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, UpdateAll: true}).Create(&effects).Error
}

func seedRecipes(tx *gorm.DB) error {
	// --- 4. เพาะเมล็ดพันธุ์ "Recipes" ---
	log.Println("Seeding/Updating recipes...")
	recipes := []domain.Recipe{
		{OutputElementID: 5, BaseMPCost: 25, Ingredients: []*domain.RecipeIngredient{{InputElementID: 1, Quantity: 1}, {InputElementID: 2, Quantity: 1}}},
		{OutputElementID: 6, BaseMPCost: 25, Ingredients: []*domain.RecipeIngredient{{InputElementID: 1, Quantity: 1}, {InputElementID: 3, Quantity: 1}}},
		{OutputElementID: 7, BaseMPCost: 30, Ingredients: []*domain.RecipeIngredient{{InputElementID: 1, Quantity: 1}, {InputElementID: 4, Quantity: 1}}},
		{OutputElementID: 8, BaseMPCost: 30, Ingredients: []*domain.RecipeIngredient{{InputElementID: 2, Quantity: 1}, {InputElementID: 3, Quantity: 1}}},
		{OutputElementID: 9, BaseMPCost: 30, Ingredients: []*domain.RecipeIngredient{{InputElementID: 2, Quantity: 1}, {InputElementID: 4, Quantity: 1}}},
		{OutputElementID: 10, BaseMPCost: 35, Ingredients: []*domain.RecipeIngredient{{InputElementID: 3, Quantity: 1}, {InputElementID: 4, Quantity: 1}}},
		{OutputElementID: 11, BaseMPCost: 40, Ingredients: []*domain.RecipeIngredient{{InputElementID: 1, Quantity: 2}}},
		{OutputElementID: 12, BaseMPCost: 40, Ingredients: []*domain.RecipeIngredient{{InputElementID: 2, Quantity: 2}}},
		{OutputElementID: 13, BaseMPCost: 40, Ingredients: []*domain.RecipeIngredient{{InputElementID: 3, Quantity: 2}}},
		{OutputElementID: 14, BaseMPCost: 50, Ingredients: []*domain.RecipeIngredient{{InputElementID: 4, Quantity: 2}}},
		{OutputElementID: 15, BaseMPCost: 60, Ingredients: []*domain.RecipeIngredient{{InputElementID: 1, Quantity: 1}, {InputElementID: 2, Quantity: 1}, {InputElementID: 3, Quantity: 1}, {InputElementID: 4, Quantity: 1}}},
	}
	tx.Exec("DELETE FROM recipe_ingredients")
	tx.Exec("DELETE FROM recipes")
	if err := tx.Create(&recipes).Error; err != nil {
		return err
	}
	return nil
}

func seedSpells(tx *gorm.DB) error {
	log.Println("Seeding/Updating spells...")
	spells := []domain.Spell{
		// S
		{ID: 1, Name: "EarthSlam", ElementID: 1, MasteryID: 1, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 1, BaseValue: 60}, {EffectID: 200, DurationInTurns: 2}}},
		{ID: 2, Name: "StoneFortress", ElementID: 1, MasteryID: 2, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 2, BaseValue: 450}, {EffectID: 200, DurationInTurns: 2}}},
		{ID: 3, Name: "IronBody", ElementID: 1, MasteryID: 3, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 105, DurationInTurns: 3}}},
		{ID: 4, Name: "GravityWave", ElementID: 1, MasteryID: 4, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 301, BaseValue: -20, DurationInTurns: 2}}},
		// L
		{ID: 5, Name: "WaterWhip", ElementID: 2, MasteryID: 1, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 1, BaseValue: 50}, {EffectID: 5, BaseValue: 15}}},
		{ID: 6, Name: "HealingTide", ElementID: 2, MasteryID: 2, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 100, BaseValue: 50, DurationInTurns: 3}, {EffectID: 201, DurationInTurns: 3}}},
		{ID: 7, Name: "FocusWill", ElementID: 2, MasteryID: 3, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 101, BaseValue: 10, DurationInTurns: 3}}},
		{ID: 8, Name: "Cleanse", ElementID: 2, MasteryID: 4, APCost: 1, MPCost: 10, Effects: []*domain.SpellEffect{{EffectID: 7}, {EffectID: 3, BaseValue: 100}}},
		// G
		{ID: 9, Name: "WindBlade", ElementID: 3, MasteryID: 1, APCost: 1, MPCost: 10, Effects: []*domain.SpellEffect{{EffectID: 1, BaseValue: 40}}},
		{ID: 10, Name: "Mirage", ElementID: 3, MasteryID: 2, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 102, BaseValue: 50, DurationInTurns: 1}}},
		{ID: 11, Name: "TempoSurge", ElementID: 3, MasteryID: 3, APCost: 1, MPCost: 10, Effects: []*domain.SpellEffect{{EffectID: 6, BaseValue: 1}, {EffectID: 202, DurationInTurns: 1}}},
		{ID: 12, Name: "HinderingWind", ElementID: 3, MasteryID: 4, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 301, BaseValue: -30, DurationInTurns: 2}}},
		// P
		{ID: 13, Name: "PureBeam", ElementID: 4, MasteryID: 1, APCost: 2, MPCost: 20, Effects: []*domain.SpellEffect{{EffectID: 4, BaseValue: 80}}},
		{ID: 14, Name: "EnergyThorns", ElementID: 4, MasteryID: 2, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 2, BaseValue: 250}, {EffectID: 104, BaseValue: 25}}},
		{ID: 15, Name: "AwakenPower", ElementID: 4, MasteryID: 3, APCost: 1, MPCost: 10, Effects: []*domain.SpellEffect{{EffectID: 103, BaseValue: 30, DurationInTurns: 1}, {EffectID: 203, DurationInTurns: 1}}},
		{ID: 16, Name: "ExposeWeakness", ElementID: 4, MasteryID: 4, APCost: 2, MPCost: 15, Effects: []*domain.SpellEffect{{EffectID: 302, BaseValue: 20, DurationInTurns: 2}}},
		// Tier 1 Spells (ปรับสมดุลแล้ว)
		{ID: 17, Name: "MudPrison", ElementID: 5, MasteryID: 4, APCost: 2, MPCost: 25, Effects: []*domain.SpellEffect{{EffectID: 303, DurationInTurns: 1}}},
		{ID: 18, Name: "EnergySap", ElementID: 5, MasteryID: 4, APCost: 3, MPCost: 30, Effects: []*domain.SpellEffect{{EffectID: 5, BaseValue: 25}}},
		{ID: 21, Name: "MoltenMeteor", ElementID: 7, MasteryID: 1, APCost: 2, MPCost: 30, Effects: []*domain.SpellEffect{{EffectID: 1, BaseValue: 75}, {EffectID: 306, BaseValue: 20, DurationInTurns: 2}}},
	}
	tx.Exec("DELETE FROM spell_effects")
	tx.Exec("DELETE FROM spells")
	return tx.Create(&spells).Error
}

func seedGameConfigs(tx *gorm.DB) error {
	log.Println("Seeding/Updating game_configs...")

	configs := []domain.GameConfig{
		// ========================================================================
		// General Game Rules
		// ========================================================================
		{Key: "GAME_VERSION", Value: "1.0.0", Description: "The current version of the game data."},

		// ========================================================================
		// Character Stat Calculation Rules
		// ========================================================================
		{Key: "STAT_HP_BASE", Value: "900", Description: "ค่า HP พื้นฐานเริ่มต้นของทุกตัวละคร"},
		{Key: "STAT_HP_PER_TALENT_S", Value: "30", Description: "ค่า HP ที่จะได้รับต่อ 1 แต้ม Talent S"},
		{Key: "STAT_MP_BASE", Value: "100", Description: "ค่า MP พื้นฐานเริ่มต้น"},
		{Key: "STAT_MP_PER_TALENT_L", Value: "2", Description: "ค่า MP ที่จะได้รับต่อ 1 แต้ม Talent L"},
		{Key: "STAT_INITIATIVE_BASE", Value: "50", Description: "ค่าความเร็วพื้นฐาน"},
		{Key: "STAT_INITIATIVE_PER_TALENT_G", Value: "1", Description: "ค่าความเร็วที่จะได้รับต่อ 1 แต้ม Talent G"},
		{Key: "STAT_ENDURANCE_BASE", Value: "100", Description: "ค่าความอดทนพื้นฐาน"},
		{Key: "STAT_ENDURANCE_PER_TALENT_S", Value: "5", Description: "ค่าความอดทนที่จะได้รับต่อ 1 แต้ม Talent S"},

		// ========================================================================
		// Character Creation Talent Allocation
		// ========================================================================
		{Key: "TALENT_PRIMARY_ALLOCATION", Value: "90", Description: "แต้มพรสวรรค์ที่จะมอบให้ธาตุหลัก"},
		{Key: "TALENT_AFFINITY_ALLOCATION", Value: "4", Description: "แต้มพรสวรรค์ที่จะมอบให้ธาตุคู่หู"},
		{Key: "TALENT_SECONDARY_ALLOCATION", Value: "3", Description: "แต้มพรสวรรค์ที่จะมอบให้ธาตุรอง"},

		// ========================================================================
		// Combat System Rules
		// ========================================================================
		{Key: "COMBAT_AP_PER_TURN", Value: "3", Description: "จำนวน AP ที่ผู้เล่นได้รับเมื่อเริ่มเทิร์น"},
		{Key: "COMBAT_BASE_AP_CAP", Value: "6", Description: "ขีดจำกัด AP สูงสุดพื้นฐานที่สามารถสะสมได้"},

		// ========================================================================
		// Resource Regeneration Rules
		// ========================================================================
		{Key: "PASSIVE_MP_REGEN_PER_MINUTE", Value: "5", Description: "จำนวน MP ที่ฟื้นฟูอัตโนมัติทุกๆ 1 นาที (นอกการต่อสู้)"},
	}

	// ใช้ OnConflict เพื่อให้สามารถรัน Seeder นี้ซ้ำได้โดยไม่ Error
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "description"}),
	}).Create(&configs).Error; err != nil {
		return err
	}

	return nil
}

func seedEnemies(tx *gorm.DB) error {
	log.Println("Seeding/Updating enemies and their AI...")

	// ========================================================================
	// ENEMY 1: TRAINING GOLEM (POTENCY) - สำหรับผู้เล่นสาย S
	// ========================================================================
	golemP := domain.Enemy{
		ID:           1,
		Name:         "TRAINING_GOLEM_POTENCY",
		DisplayNames: datatypes.JSON(`{"en": "Potency Training Golem", "th": "โกเลมพลังงานฝึกหัด"}`), // ⭐️ ชื่อใหม่!
		ElementID:    4,                                                                              // Potency
		Level:        1,
		MaxHP:        250,
		Initiative:   40,
		MaxEndurance: 100,
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemP)
	abilitiesP := []domain.EnemyAbility{
		{ID: 1, EnemyID: 1, Name: "P_PUNCH", DisplayNames: datatypes.JSON(`{"en": "Punch", "th": "หมัดตรง"}`), APCost: 1, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 20}]`)},
		{ID: 2, EnemyID: 1, Name: "P_TREMOR", DisplayNames: datatypes.JSON(`{"en": "Tremor", "th": "คลื่นพลัง"}`), APCost: 3, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 40}, {"effect_id": 301, "value": -10, "duration": 2}]`)},
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesP)
	tx.Where("enemy_id = ?", 1).Delete(&domain.EnemyAI{})
	aiRulesP := []domain.EnemyAI{
		{EnemyID: 1, Priority: 1, Condition: domain.AIConditionTurnIs, ConditionValue: 2, Action: domain.AIActionUseAbility, Target: domain.AITargetPlayer, AbilityToUseID: &abilitiesP[1].ID},
		{EnemyID: 1, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: domain.AITargetPlayer, AbilityToUseID: &abilitiesP[0].ID},
	}
	tx.Create(&aiRulesP)

	// ========================================================================
	// ENEMY 2: TRAINING GOLEM (SOLIDITY) - สำหรับผู้เล่นสาย L
	// ========================================================================
	golemS := domain.Enemy{
		ID:           2,
		Name:         "TRAINING_GOLEM_SOLIDITY",
		DisplayNames: datatypes.JSON(`{"en": "Solidity Training Golem", "th": "โกเลมศิลาฝึกหัด"}`), // ⭐️ ชื่อใหม่!
		ElementID:    1,                                                                            // Solidity
		Level:        1,
		MaxHP:        300,
		Initiative:   35,
		MaxEndurance: 120,
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemS)
	abilitiesS := []domain.EnemyAbility{
		{ID: 3, EnemyID: 2, Name: "S_SLAP", DisplayNames: datatypes.JSON(`{"en": "Slap", "th": "ตบ"}`), APCost: 1, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 15}]`)},
		{ID: 4, EnemyID: 2, Name: "S_HARDEN", DisplayNames: datatypes.JSON(`{"en": "Harden", "th": "กายาหิน"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 110, "target": "SELF", "duration": 2}]`)}, // สมมติ 110 คือ BUFF_DEFENSE
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesS)
	tx.Where("enemy_id = ?", 2).Delete(&domain.EnemyAI{})
	aiRulesS := []domain.EnemyAI{
		{EnemyID: 2, Priority: 1, Condition: domain.AIConditionSelfHPBelow, ConditionValue: 0.5, Action: domain.AIActionUseAbility, Target: domain.AITargetSelf, AbilityToUseID: &abilitiesS[1].ID},
		{EnemyID: 2, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: domain.AITargetPlayer, AbilityToUseID: &abilitiesS[0].ID},
	}
	tx.Create(&aiRulesS)

	// ========================================================================
	// ENEMY 3: TRAINING GOLEM (LIQUIDITY) - สำหรับผู้เล่นสาย G
	// ========================================================================
	golemL := domain.Enemy{
		ID:           3,
		Name:         "TRAINING_GOLEM_LIQUIDITY",
		DisplayNames: datatypes.JSON(`{"en": "Liquidity Training Golem", "th": "โกเลมวารีฝึกหัด"}`), // ⭐️ ชื่อใหม่!
		ElementID:    2,                                                                             // Liquidity
		Level:        1,
		MaxHP:        220,
		Initiative:   45,
		MaxEndurance: 80,
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemL)
	abilitiesL := []domain.EnemyAbility{
		{ID: 5, EnemyID: 3, Name: "L_SPLASH", DisplayNames: datatypes.JSON(`{"en": "Splash", "th": "สาดน้ำ"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 25}]`)},
		{ID: 6, EnemyID: 3, Name: "L_REGEN", DisplayNames: datatypes.JSON(`{"en": "Regenerate", "th": "ฟื้นฟู"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 100, "target": "SELF", "value": 20, "duration": 3}]`)},
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesL)
	tx.Where("enemy_id = ?", 3).Delete(&domain.EnemyAI{})
	aiRulesL := []domain.EnemyAI{
		{EnemyID: 3, Priority: 1, Condition: domain.AIConditionSelfHPBelow, ConditionValue: 0.5, Action: domain.AIActionUseAbility, Target: domain.AITargetSelf, AbilityToUseID: &abilitiesL[1].ID},
		{EnemyID: 3, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: domain.AITargetPlayer, AbilityToUseID: &abilitiesL[0].ID},
	}
	tx.Create(&aiRulesL)

	// ========================================================================
	// ENEMY 4: TRAINING GOLEM (TEMPO) - สำหรับผู้เล่นสาย P
	// ========================================================================
	golemG := domain.Enemy{
		ID:           4,
		Name:         "TRAINING_GOLEM_TEMPO",
		DisplayNames: datatypes.JSON(`{"en": "Tempo Training Golem", "th": "โกเลมวายุฝึกหัด"}`), // ⭐️ ชื่อใหม่!
		ElementID:    3,                                                                         // Tempo
		Level:        1,
		MaxHP:        200,
		Initiative:   55,
		MaxEndurance: 70,
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemG)
	abilitiesG := []domain.EnemyAbility{
		{ID: 7, EnemyID: 4, Name: "G_WIND_SLASH", DisplayNames: datatypes.JSON(`{"en": "Wind Slash", "th": "คมลม"}`), APCost: 1, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 25}]`)},
		{ID: 8, EnemyID: 4, Name: "G_EVADE", DisplayNames: datatypes.JSON(`{"en": "Evade", "th": "หลบหลีก"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 102, "target": "SELF", "value": 50, "duration": 1}]`)},
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesG)
	tx.Where("enemy_id = ?", 4).Delete(&domain.EnemyAI{})
	aiRulesG := []domain.EnemyAI{
		{EnemyID: 4, Priority: 1, Condition: domain.AIConditionTurnIs, ConditionValue: 2, Action: domain.AIActionUseAbility, Target: domain.AITargetSelf, AbilityToUseID: &abilitiesG[1].ID},
		{EnemyID: 4, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: domain.AITargetPlayer, AbilityToUseID: &abilitiesG[0].ID},
	}
	tx.Create(&aiRulesG)

	return nil
}

func seedElementalMatchups(tx *gorm.DB) error {
	log.Println("Seeding/Updating elemental matchups...")

	matchups := []domain.ElementalMatchup{
		// --- S (ID:1) Attacking ---
		{AttackingElementID: 1, DefendingElementID: 1, Modifier: 1.0}, // S vs S
		{AttackingElementID: 1, DefendingElementID: 2, Modifier: 0.8}, // S vs L (แพ้ทางเล็กน้อย)
		{AttackingElementID: 1, DefendingElementID: 3, Modifier: 1.0}, // S vs G
		{AttackingElementID: 1, DefendingElementID: 4, Modifier: 1.5}, // S vs P (ชนะทางขาดลอย!)

		// --- L (ID:2) Attacking ---
		{AttackingElementID: 2, DefendingElementID: 1, Modifier: 1.3}, // L vs S (ชนะทาง)
		{AttackingElementID: 2, DefendingElementID: 2, Modifier: 1.0}, // L vs L
		{AttackingElementID: 2, DefendingElementID: 3, Modifier: 0.7}, // L vs G (แพ้ทางหนัก)
		{AttackingElementID: 2, DefendingElementID: 4, Modifier: 1.0}, // L vs P

		// --- G (ID:3) Attacking ---
		{AttackingElementID: 3, DefendingElementID: 1, Modifier: 1.0}, // G vs S
		{AttackingElementID: 3, DefendingElementID: 2, Modifier: 1.2}, // G vs L (ชนะทางแบบคุมเกม)
		{AttackingElementID: 3, DefendingElementID: 3, Modifier: 1.0}, // G vs G
		{AttackingElementID: 3, DefendingElementID: 4, Modifier: 0.8}, // G vs P (แพ้ทางเล็กน้อย)

		// --- P (ID:4) Attacking ---
		{AttackingElementID: 4, DefendingElementID: 1, Modifier: 0.7}, // P vs S (แพ้ทางหนัก)
		{AttackingElementID: 4, DefendingElementID: 2, Modifier: 1.0}, // P vs L
		{AttackingElementID: 4, DefendingElementID: 3, Modifier: 1.4}, // P vs G (ชนะทางรุนแรง)
		{AttackingElementID: 4, DefendingElementID: 4, Modifier: 1.0}, // P vs P
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "attacking_element_id"}, {Name: "defending_element_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"modifier"}),
	}).Create(&matchups).Error
}

func seedPveContent(tx *gorm.DB) error {
	log.Println("Seeding/Updating PvE content...")
	mainStoryRealm := domain.Realm{ID: 1, Name: "MAIN_STORY", DisplayNames: datatypes.JSON(`{"en": "Main Story", "th": "เนื้อเรื่องหลัก"}`)}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&mainStoryRealm)
	chapter1 := domain.Chapter{ID: 1, RealmID: 1, ChapterNumber: 1, Name: "THE_AWAKENING", DisplayNames: datatypes.JSON(`{"en": "The Awakening", "th": "การตื่นรู้"}`)}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&chapter1)
	stages := []domain.Stage{
		{ID: 1, ChapterID: 1, StageNumber: 1, Name: "FORGOTTEN_PATH", DisplayNames: datatypes.JSON(`{"en": "A Forgotten Path", "th": "เส้นทางที่ถูกลืม"}`), StageType: domain.StageTypeStory},
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&stages).Error
}
