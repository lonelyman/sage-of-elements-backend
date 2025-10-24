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

/*
	func seedSpells(tx *gorm.DB) error {
		log.Println("Seeding/Updating spells...")
		spells := []domain.Spell{
			// --- S (Solidity) ---
			{ID: 1, Name: "EarthSlam", TargetType: domain.TargetTypeEnemy, ElementID: 1, MasteryID: 1, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Earth Slam", "th": "ปฐพีถล่ม"},
				Descriptions: datatypes.JSONMap{"en": "Deals Solidity damage and grants S Stance.", "th": "สร้างความเสียหายปฐพีและมอบสถานะ S"},
				Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 60}, {EffectID: 200, DurationInTurns: 2}}},
			{ID: 2, Name: "StoneFortress", TargetType: domain.TargetTypeSelf, ElementID: 1, MasteryID: 2, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Stone Fortress", "th": "ปราการศิลา"},
				Descriptions: datatypes.JSONMap{"en": "Creates a shield and grants S Stance.", "th": "สร้างโล่ป้องกันและมอบสถานะ S"},
				Effects:      []*domain.SpellEffect{{EffectID: 2, BaseValue: 450}, {EffectID: 200, DurationInTurns: 2}}},
			{ID: 3, Name: "IronBody", TargetType: domain.TargetTypeSelf, ElementID: 1, MasteryID: 3, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Iron Body", "th": "กายาเหล็ก"},
				Descriptions: datatypes.JSONMap{"en": "Greatly reduces incoming damage for a few turns.", "th": "ลดความเสียหายที่ได้รับอย่างมากชั่วขณะ"},
				Effects:      []*domain.SpellEffect{{EffectID: 105, DurationInTurns: 3}}}, // Assuming 105 is DEF_UP buff
			{ID: 4, Name: "GravityWave", TargetType: domain.TargetTypeEnemy, ElementID: 1, MasteryID: 4, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Gravity Wave", "th": "คลื่นโน้มถ่วง"},
				Descriptions: datatypes.JSONMap{"en": "Slows the target's initiative.", "th": "ลดค่าความคิดริเริ่มของเป้าหมาย"},
				Effects:      []*domain.SpellEffect{{EffectID: 301, BaseValue: -20, DurationInTurns: 2}}}, // EffectID 301 is DEBUFF_SLOW

			// --- L (Liquidity) ---
			{ID: 5, Name: "WaterWhip", TargetType: domain.TargetTypeEnemy, ElementID: 2, MasteryID: 1, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Water Whip", "th": "แส้วารี"},
				Descriptions: datatypes.JSONMap{"en": "Deals Liquidity damage and drains MP.", "th": "สร้างความเสียหายวารีและดูดกลืน MP"},
				Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 50}, {EffectID: 5, BaseValue: 15}}}, // EffectID 5 is MP_DRAIN?
			{ID: 6, Name: "HealingTide", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 2, APCost: 2, MPCost: 15, // Changed target to SELF
				DisplayNames: datatypes.JSONMap{"en": "Healing Tide", "th": "กระแสชโลม"},
				Descriptions: datatypes.JSONMap{"en": "Applies HP Regeneration and grants L Stance.", "th": "มอบผลฟื้นฟู HP ต่อเนื่องและมอบสถานะ L"},
				Effects:      []*domain.SpellEffect{{EffectID: 100, BaseValue: 50, DurationInTurns: 3}, {EffectID: 201, DurationInTurns: 3}}},
			{ID: 7, Name: "FocusWill", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 3, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Focus Will", "th": "รวบรวมจิต"},
				Descriptions: datatypes.JSONMap{"en": "Applies MP Regeneration.", "th": "มอบผลฟื้นฟู MP ต่อเนื่อง"},
				Effects:      []*domain.SpellEffect{{EffectID: 101, BaseValue: 10, DurationInTurns: 3}}},
			{ID: 8, Name: "Cleanse", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 4, APCost: 1, MPCost: 10, // Changed target to SELF
				DisplayNames: datatypes.JSONMap{"en": "Cleanse", "th": "ชำระล้าง"},
				Descriptions: datatypes.JSONMap{"en": "Removes debuffs and restores HP.", "th": "ลบล้างสถานะผิดปกติและฟื้นฟู HP"},
				Effects:      []*domain.SpellEffect{{EffectID: 7}, {EffectID: 3, BaseValue: 100}}}, // EffectID 7 is CLEANSE_DEBUFF?

			// --- G (Gale) ---
			{ID: 9, Name: "WindBlade", TargetType: domain.TargetTypeEnemy, ElementID: 3, MasteryID: 1, APCost: 1, MPCost: 10,
				DisplayNames: datatypes.JSONMap{"en": "Wind Blade", "th": "คมมีดวายุ"},
				Descriptions: datatypes.JSONMap{"en": "Deals Gale damage.", "th": "สร้างความเสียหายวายุ"},
				Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 40}}},
			{ID: 10, Name: "Mirage", TargetType: domain.TargetTypeSelf, ElementID: 3, MasteryID: 2, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Mirage", "th": "ภาพลวงตา"},
				Descriptions: datatypes.JSONMap{"en": "Increases Evasion for one turn.", "th": "เพิ่มอัตราหลบหลีกชั่วขณะ"},
				Effects:      []*domain.SpellEffect{{EffectID: 102, BaseValue: 50, DurationInTurns: 1}}}, // EffectID 102 is BUFF_EVASION?
			{ID: 11, Name: "TempoSurge", TargetType: domain.TargetTypeSelf, ElementID: 3, MasteryID: 3, APCost: 1, MPCost: 10,
				DisplayNames: datatypes.JSONMap{"en": "Tempo Surge", "th": "เร่งจังหวะ"},
				Descriptions: datatypes.JSONMap{"en": "Grants extra AP next turn and G Stance.", "th": "ได้รับ AP เพิ่มในเทิร์นถัดไปและมอบสถานะ G"},
				Effects:      []*domain.SpellEffect{{EffectID: 6, BaseValue: 1}, {EffectID: 202, DurationInTurns: 1}}}, // EffectID 6 is GAIN_AP_NEXT_TURN?
			{ID: 12, Name: "HinderingWind", TargetType: domain.TargetTypeEnemy, ElementID: 3, MasteryID: 4, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Hindering Wind", "th": "ลมเหนี่ยวรั้ง"},
				Descriptions: datatypes.JSONMap{"en": "Greatly slows the target's initiative.", "th": "ลดค่าความคิดริเริ่มของเป้าหมายอย่างมาก"},
				Effects:      []*domain.SpellEffect{{EffectID: 301, BaseValue: -30, DurationInTurns: 2}}},

			// --- P (Plasma) ---
			{ID: 13, Name: "PureBeam", TargetType: domain.TargetTypeEnemy, ElementID: 4, MasteryID: 1, APCost: 2, MPCost: 20,
				DisplayNames: datatypes.JSONMap{"en": "Pure Beam", "th": "ลำแสงบริสุทธิ์"},
				Descriptions: datatypes.JSONMap{"en": "Deals Plasma damage ignoring defenses.", "th": "สร้างความเสียหายพลาสมาโดยไม่สนพลังป้องกัน"},
				Effects:      []*domain.SpellEffect{{EffectID: 4, BaseValue: 80}}}, // EffectID 4 is PURE_DAMAGE?
			{ID: 14, Name: "EnergyThorns", TargetType: domain.TargetTypeSelf, ElementID: 4, MasteryID: 2, APCost: 2, MPCost: 15, // Changed target to SELF
				DisplayNames: datatypes.JSONMap{"en": "Energy Thorns", "th": "หนามพลังงาน"},
				Descriptions: datatypes.JSONMap{"en": "Creates a shield that damages attackers.", "th": "สร้างโล่ที่สะท้อนความเสียหายแก่ผู้โจมตี"},
				Effects:      []*domain.SpellEffect{{EffectID: 2, BaseValue: 250}, {EffectID: 104, BaseValue: 25}}}, // EffectID 104 is BUFF_THORNS?
			{ID: 15, Name: "AwakenPower", TargetType: domain.TargetTypeSelf, ElementID: 4, MasteryID: 3, APCost: 1, MPCost: 10,
				DisplayNames: datatypes.JSONMap{"en": "Awaken Power", "th": "ปลุกพลัง"},
				Descriptions: datatypes.JSONMap{"en": "Greatly increases damage next turn and grants P Stance.", "th": "เพิ่มความเสียหายอย่างมากในเทิร์นถัดไปและมอบสถานะ P"},
				Effects:      []*domain.SpellEffect{{EffectID: 103, BaseValue: 30, DurationInTurns: 1}, {EffectID: 203, DurationInTurns: 1}}}, // EffectID 103 is BUFF_DAMAGE_UP?
			{ID: 16, Name: "ExposeWeakness", TargetType: domain.TargetTypeEnemy, ElementID: 4, MasteryID: 4, APCost: 2, MPCost: 15,
				DisplayNames: datatypes.JSONMap{"en": "Expose Weakness", "th": "เปิดจุดอ่อน"},
				Descriptions: datatypes.JSONMap{"en": "Makes the target take increased damage.", "th": "ทำให้เป้าหมายได้รับความเสียหายเพิ่มขึ้น"},
				Effects:      []*domain.SpellEffect{{EffectID: 302, BaseValue: 20, DurationInTurns: 2}}}, // EffectID 302 is DEBUFF_VULNERABLE?

			// --- Tier 1 Spells ---
			{ID: 17, Name: "MudPrison", TargetType: domain.TargetTypeEnemy, ElementID: 5, MasteryID: 4, APCost: 2, MPCost: 25,
				DisplayNames: datatypes.JSONMap{"en": "Mud Prison", "th": "คุกโคลน"},
				Descriptions: datatypes.JSONMap{"en": "Stuns the target for one turn.", "th": "ทำให้เป้าหมายติดสถานะมึนงง 1 เทิร์น"},
				Effects:      []*domain.SpellEffect{{EffectID: 303, DurationInTurns: 1}}}, // EffectID 303 is DEBUFF_STUN?
			{ID: 18, Name: "EnergySap", TargetType: domain.TargetTypeEnemy, ElementID: 5, MasteryID: 4, APCost: 3, MPCost: 30,
				DisplayNames: datatypes.JSONMap{"en": "Energy Sap", "th": "ดูดกลืนพลัง"},
				Descriptions: datatypes.JSONMap{"en": "Drains a large amount of MP from the target.", "th": "ดูดกลืน MP จำนวนมากจากเป้าหมาย"},
				Effects:      []*domain.SpellEffect{{EffectID: 5, BaseValue: 25}}},
			{ID: 21, Name: "MoltenMeteor", TargetType: domain.TargetTypeEnemy, ElementID: 7, MasteryID: 1, APCost: 2, MPCost: 30,
				DisplayNames: datatypes.JSONMap{"en": "Molten Meteor", "th": "อุกกาบาตหลอมเหลว"},
				Descriptions: datatypes.JSONMap{"en": "Deals massive damage and applies Burn.", "th": "สร้างความเสียหายรุนแรงและติดสถานะเผาไหม้"},
				Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 75}, {EffectID: 306, BaseValue: 20, DurationInTurns: 2}}}, // EffectID 306 is DOT_BURN?
		}
		tx.Exec("DELETE FROM spell_effects")
		tx.Exec("DELETE FROM spells")
		return tx.Create(&spells).Error
	}
*/
func seedSpells(tx *gorm.DB) error {
	log.Println("Seeding/Updating spells (Rebalanced for Patch 1)...")
	spells := []domain.Spell{
		// --- S (Solidity) - เน้นป้องกัน, Debuff เบื้องต้น ---
		{ID: 1, Name: "EarthSlam", TargetType: domain.TargetTypeEnemy, ElementID: 1, MasteryID: 1, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Earth Slam", "th": "ปฐพีทุบ"}, // เปลี่ยนชื่อนิดหน่อย
			Descriptions: datatypes.JSONMap{"en": "Deals Solidity damage and grants S Stance.", "th": "สร้างความเสียหายปฐพีและมอบสถานะ S"},
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 55}, {EffectID: 200, DurationInTurns: 2}}}, // ลด Damage ลงเล็กน้อย
		{ID: 2, Name: "StoneSkin", TargetType: domain.TargetTypeSelf, ElementID: 1, MasteryID: 2, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Fortress
			DisplayNames: datatypes.JSONMap{"en": "Stone Skin", "th": "ผิวศิลา"},
			Descriptions: datatypes.JSONMap{"en": "Creates a small shield and grants S Stance.", "th": "สร้างโล่ป้องกันเล็กน้อยและมอบสถานะ S"},
			Effects:      []*domain.SpellEffect{{EffectID: 2, BaseValue: 200}, {EffectID: 200, DurationInTurns: 2}}}, // ลดค่า Shield ลงเยอะมาก!
		{ID: 3, Name: "Reinforce", TargetType: domain.TargetTypeSelf, ElementID: 1, MasteryID: 3, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก IronBody
			DisplayNames: datatypes.JSONMap{"en": "Reinforce", "th": "เสริมกำลัง"},
			Descriptions: datatypes.JSONMap{"en": "Slightly reduces incoming damage for a few turns.", "th": "ลดความเสียหายที่ได้รับเล็กน้อยชั่วขณะ"},
			Effects:      []*domain.SpellEffect{{EffectID: 110, BaseValue: 15, DurationInTurns: 3}}}, // เปลี่ยนเป็น DEF_UP มาตรฐาน (สมมติ ID 110 คือ DEF+) ให้ค่าเป็น % หรือค่าคงที่ (อันนี้สมมติ +15 Def)
		{ID: 4, Name: "Tremor", TargetType: domain.TargetTypeEnemy, ElementID: 1, MasteryID: 4, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก GravityWave
			DisplayNames: datatypes.JSONMap{"en": "Tremor", "th": "สะเทือน"},
			Descriptions: datatypes.JSONMap{"en": "Slightly slows the target's initiative.", "th": "ลดค่าความคิดริเริ่มของเป้าหมายเล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 301, BaseValue: -15, DurationInTurns: 2}}}, // ลดผล Slow ลง

		// --- L (Liquidity) - เน้นฟื้นฟูพื้นฐาน ---
		{ID: 5, Name: "AquaShot", TargetType: domain.TargetTypeEnemy, ElementID: 2, MasteryID: 1, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Whip, เอา MP Drain ออกก่อน
			DisplayNames: datatypes.JSONMap{"en": "Aqua Shot", "th": "กระสุนวารี"},
			Descriptions: datatypes.JSONMap{"en": "Deals Liquidity damage.", "th": "สร้างความเสียหายวารี"},
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 50}}}, // เหลือแค่ Damage
		{ID: 6, Name: "SoothingMist", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 2, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Tide
			DisplayNames: datatypes.JSONMap{"en": "Soothing Mist", "th": "หมอกบรรเทา"},
			Descriptions: datatypes.JSONMap{"en": "Applies minor HP Regeneration and grants L Stance.", "th": "มอบผลฟื้นฟู HP เล็กน้อยต่อเนื่องและมอบสถานะ L"},
			Effects:      []*domain.SpellEffect{{EffectID: 100, BaseValue: 25, DurationInTurns: 3}, {EffectID: 201, DurationInTurns: 3}}}, // ลดค่า Heal per Turn ลง
		{ID: 7, Name: "Meditate", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 3, APCost: 1, MPCost: 0, // เปลี่ยนชื่อ, ลด AP, ไม่ใช้ MP!
			DisplayNames: datatypes.JSONMap{"en": "Meditate", "th": "ทำสมาธิ"},
			Descriptions: datatypes.JSONMap{"en": "Restores a small amount of MP over time.", "th": "ฟื้นฟู MP เล็กน้อยต่อเนื่อง"},
			Effects:      []*domain.SpellEffect{{EffectID: 101, BaseValue: 10, DurationInTurns: 3}}}, // MP Regen เท่าเดิม แต่ Cost ถูกลงมาก
		{ID: 8, Name: "MinorHeal", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 4, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Cleanse, เพิ่ม Cost, เอา Cleanse ออก
			DisplayNames: datatypes.JSONMap{"en": "Minor Heal", "th": "ฟื้นฟูเล็กน้อย"},
			Descriptions: datatypes.JSONMap{"en": "Restores a small amount of HP.", "th": "ฟื้นฟู HP เล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 3, BaseValue: 75}}}, // ลดค่า Heal ลง, เพิ่ม Cost กลับไปเป็นปกติ

		// --- G (Gale) - เน้นความเร็ว, ก่อกวนเบาๆ ---
		{ID: 9, Name: "WindSlash", TargetType: domain.TargetTypeEnemy, ElementID: 3, MasteryID: 1, APCost: 1, MPCost: 10, // เปลี่ยนชื่อนิดหน่อย
			DisplayNames: datatypes.JSONMap{"en": "Wind Slash", "th": "ดาบลม"},
			Descriptions: datatypes.JSONMap{"en": "Deals Gale damage quickly.", "th": "สร้างความเสียหายวายุอย่างรวดเร็ว"},
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 35}}}, // Damage เท่าเดิม (เพราะ Cost ถูก)
		{ID: 10, Name: "Blur", TargetType: domain.TargetTypeSelf, ElementID: 3, MasteryID: 2, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Mirage
			DisplayNames: datatypes.JSONMap{"en": "Blur", "th": "พร่ามัว"},
			Descriptions: datatypes.JSONMap{"en": "Slightly increases Evasion for one turn.", "th": "เพิ่มอัตราหลบหลีกเล็กน้อยชั่วขณะ"},
			Effects:      []*domain.SpellEffect{{EffectID: 102, BaseValue: 40, DurationInTurns: 1}}}, // ลด % Evasion ลง
		{ID: 11, Name: "SwiftStep", TargetType: domain.TargetTypeSelf, ElementID: 3, MasteryID: 3, APCost: 1, MPCost: 10, // เปลี่ยนชื่อจาก TempoSurge, เอา AP Gain ออก!
			DisplayNames: datatypes.JSONMap{"en": "Swift Step", "th": "ก้าววายุ"},
			Descriptions: datatypes.JSONMap{"en": "Grants G Stance.", "th": "มอบสถานะ G"},
			Effects:      []*domain.SpellEffect{{EffectID: 202, DurationInTurns: 2}}}, // เหลือแค่ Stance (อาจจะเพิ่มระยะเวลา Stance)
		{ID: 12, Name: "Gust", TargetType: domain.TargetTypeEnemy, ElementID: 3, MasteryID: 4, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Hindering Wind
			DisplayNames: datatypes.JSONMap{"en": "Gust", "th": "ลมกระโชก"},
			Descriptions: datatypes.JSONMap{"en": "Moderately slows the target's initiative.", "th": "ลดค่าความคิดริเริ่มของเป้าหมายปานกลาง"},
			Effects:      []*domain.SpellEffect{{EffectID: 301, BaseValue: -25, DurationInTurns: 2}}}, // ลดผล Slow ลงเล็กน้อย

		// --- P (Plasma) - เน้นดาเมจพื้นฐาน, บัฟ/ดีบัฟเบาๆ ---
		{ID: 13, Name: "PlasmaBolt", TargetType: domain.TargetTypeEnemy, ElementID: 4, MasteryID: 1, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก PureBeam, เอา Ignore Def ออก, ลด MP Cost
			DisplayNames: datatypes.JSONMap{"en": "Plasma Bolt", "th": "กระสุนพลาสมา"},
			Descriptions: datatypes.JSONMap{"en": "Deals Plasma damage.", "th": "สร้างความเสียหายพลาสมา"},
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 65}}}, // Damage ปกติ (อาจจะแรงกว่าธาตุอื่นนิดหน่อย)
		{ID: 14, Name: "StaticField", TargetType: domain.TargetTypeSelf, ElementID: 4, MasteryID: 2, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Thorns
			DisplayNames: datatypes.JSONMap{"en": "Static Field", "th": "สนามไฟฟ้าสถิต"},
			Descriptions: datatypes.JSONMap{"en": "Creates a weak shield that slightly damages attackers.", "th": "สร้างโล่เบาบางที่สะท้อนความเสียหายเล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 2, BaseValue: 150}, {EffectID: 104, BaseValue: 10}}}, // ลดทั้ง Shield และ Thorns Dmg
		{ID: 15, Name: "Empower", TargetType: domain.TargetTypeSelf, ElementID: 4, MasteryID: 3, APCost: 1, MPCost: 10, // เปลี่ยนชื่อจาก Awaken Power
			DisplayNames: datatypes.JSONMap{"en": "Empower", "th": "เสริมพลัง"},
			Descriptions: datatypes.JSONMap{"en": "Slightly increases damage next turn and grants P Stance.", "th": "เพิ่มความเสียหายเล็กน้อยในเทิร์นถัดไปและมอบสถานะ P"},
			Effects:      []*domain.SpellEffect{{EffectID: 103, BaseValue: 15, DurationInTurns: 1}, {EffectID: 203, DurationInTurns: 1}}}, // ลด % Damage Buff ลงเยอะ! (สมมติ ID 103 คือ ATK+)
		{ID: 16, Name: "Analyze", TargetType: domain.TargetTypeEnemy, ElementID: 4, MasteryID: 4, APCost: 2, MPCost: 15, // เปลี่ยนชื่อจาก Expose Weakness
			DisplayNames: datatypes.JSONMap{"en": "Analyze", "th": "วิเคราะห์"},
			Descriptions: datatypes.JSONMap{"en": "Makes the target take slightly increased damage.", "th": "ทำให้เป้าหมายได้รับความเสียหายเพิ่มขึ้นเล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 302, BaseValue: 10, DurationInTurns: 2}}}, // ลด % Vulnerable ลง (สมมติ ID 302 คือ Vulnerable)

		// --- Tier 1 Spells - ทำให้เบาลง ให้พอเห็นความต่าง แต่ไม่โกง ---
		{ID: 17, Name: "EntanglingRoots", TargetType: domain.TargetTypeEnemy, ElementID: 5, MasteryID: 4, APCost: 2, MPCost: 20, // เปลี่ยนชื่อจาก MudPrison, เปลี่ยน Stun เป็น Slow แรงๆ แทน, ลด Cost
			DisplayNames: datatypes.JSONMap{"en": "Entangling Roots", "th": "รากไม้พันธนาการ"},
			Descriptions: datatypes.JSONMap{"en": "Greatly slows the target for a short duration.", "th": "ลดค่าความคิดริเริ่มเป้าหมายอย่างมากชั่วขณะ"},
			Effects:      []*domain.SpellEffect{{EffectID: 301, BaseValue: -40, DurationInTurns: 1}}}, // เปลี่ยนเป็น Slow แรงๆ แค่ 1 เทิร์น
		{ID: 18, Name: "ManaBurn", TargetType: domain.TargetTypeEnemy, ElementID: 5, MasteryID: 4, APCost: 2, MPCost: 25, // เปลี่ยนชื่อจาก EnergySap, ลด Cost, เปลี่ยนเป็น MP Damage แทน Drain?
			DisplayNames: datatypes.JSONMap{"en": "Mana Burn", "th": "เผาผลาญมานา"},
			Descriptions: datatypes.JSONMap{"en": "Damages the target's MP.", "th": "สร้างความเสียหายแก่ MP ของเป้าหมาย"},
			Effects:      []*domain.SpellEffect{{EffectID: 5, BaseValue: 30}}}, // เพิ่ม BaseValue, ลด Cost (สมมติ ID 5 คือ MP Damage)
		{ID: 21, Name: "Fireball", TargetType: domain.TargetTypeEnemy, ElementID: 7, MasteryID: 1, APCost: 2, MPCost: 25, // เปลี่ยนชื่อจาก Meteor, ลด Cost, ลด Burn Dmg
			DisplayNames: datatypes.JSONMap{"en": "Fireball", "th": "ลูกไฟ"},
			Descriptions: datatypes.JSONMap{"en": "Deals significant damage and applies a minor Burn.", "th": "สร้างความเสียหายรุนแรงและติดสถานะเผาไหม้เล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 70}, {EffectID: 306, BaseValue: 10, DurationInTurns: 2}}}, // ลด Burn Dmg ลง
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
		{Key: "STAT_MP_BASE", Value: "200", Description: "ค่า MP พื้นฐานเริ่มต้น"},
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

		// --- ⭐️ "กฎ" ของโหมดการร่าย 6 บรรทัดที่ถูกต้อง! ⭐️ ---
		// (Overcharge: จ่ายแพง 2 เท่า, ได้พลัง 1.5 เท่า)
		{Key: "CAST_MODE_OVERCHARGE_AP_MOD", Value: "2.0", Description: "ตัวคูณ AP ที่ใช้ในโหมด Overcharge (x2.0)"},
		{Key: "CAST_MODE_OVERCHARGE_MP_MOD", Value: "2.0", Description: "ตัวคูณ MP ที่ใช้ในโหมด Overcharge (x2.0)"},
		{Key: "CAST_MODE_OVERCHARGE_POWER_MOD", Value: "1.5", Description: "ตัวคูณ พลัง (Damage/Shield) ที่เพิ่มขึ้นในโหมด Overcharge (x1.5)"},

		// (Charge: จ่ายแพง 1.5 เท่า, ได้พลัง 1.2 เท่า)
		{Key: "CAST_MODE_CHARGE_AP_MOD", Value: "1.5", Description: "ตัวคูณ AP ที่ใช้ในโหมด Charge (x1.5)"},
		{Key: "CAST_MODE_CHARGE_MP_MOD", Value: "1.5", Description: "ตัวคูณ MP ที่ใช้ในโหมด Charge (x1.5)"},
		{Key: "CAST_MODE_CHARGE_POWER_MOD", Value: "1.2", Description: "ตัวคูณ พลัง (Damage/Shield) ที่เพิ่มขึ้นในโหมด Charge (x1.2)"},
		// -----------------------------------------------------------

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
	// ENEMY 1: TRAINING GOLEM (POTENCY) - (ตัวนี้ AI ดีอยู่แล้ว... แต่เราเพิ่มท่าให้มันเท่ขึ้น!)
	// ========================================================================
	golemP := domain.Enemy{ID: 1, Name: "TRAINING_GOLEM_POTENCY", DisplayNames: datatypes.JSON(`{"en": "Potency Golem", "th": "โกเลมพลังงาน"}`), ElementID: 4, Level: 1, MaxHP: 250, Initiative: 40, MaxEndurance: 100}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemP)
	abilitiesP := []domain.EnemyAbility{
		{ID: 1, EnemyID: 1, Name: "P_PUNCH", DisplayNames: datatypes.JSON(`{"en": "Punch", "th": "หมัดตรง"}`), APCost: 1, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 20}]`)},
		{ID: 2, EnemyID: 1, Name: "P_TREMOR", DisplayNames: datatypes.JSON(`{"en": "Tremor", "th": "คลื่นพลัง"}`), APCost: 3, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 40}, {"effect_id": 301, "value": -10, "duration": 2}]`)},
		{ID: 9, EnemyID: 1, Name: "P_OVERCHARGE", DisplayNames: datatypes.JSON(`{"en": "Overcharge", "th": "ปลุกพลัง"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 103, "target": "SELF", "value": 30, "duration": 2}]`)}, // ⭐️ ท่าใหม่!
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesP)
	tx.Where("enemy_id = ?", 1).Delete(&domain.EnemyAI{})
	aiRulesP := []domain.EnemyAI{
		{EnemyID: 1, Priority: 1, Condition: domain.AIConditionTurnIs, ConditionValue: 2, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesP[1].ID},       // ท่าไม้ตาย (Tremor)
		{EnemyID: 1, Priority: 10, Condition: domain.AIConditionSelfHPBelow, ConditionValue: 0.6, Action: domain.AIActionUseAbility, Target: "SELF", AbilityToUseID: &abilitiesP[2].ID}, // ท่าประจำธาตุ (Overcharge)
		{EnemyID: 1, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesP[0].ID},                         // ท่าดาเมจ (Punch)
	}
	tx.Create(&aiRulesP)

	// ========================================================================
	// ENEMY 2: TRAINING GOLEM (SOLIDITY) - (แก้ไข AI ที่ "พัง"!)
	// ========================================================================
	golemS := domain.Enemy{ID: 2, Name: "TRAINING_GOLEM_SOLIDITY", DisplayNames: datatypes.JSON(`{"en": "Solidity Golem", "th": "โกเลมศิลา"}`), ElementID: 1, Level: 1, MaxHP: 300, Initiative: 35, MaxEndurance: 120}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemS)
	abilitiesS := []domain.EnemyAbility{
		{ID: 3, EnemyID: 2, Name: "S_SLAP", DisplayNames: datatypes.JSON(`{"en": "Slap", "th": "ตบ"}`), APCost: 1, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 15}]`)},
		{ID: 4, EnemyID: 2, Name: "S_HARDEN", DisplayNames: datatypes.JSON(`{"en": "Harden", "th": "กายาหิน"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 110, "target": "SELF", "duration": 2}]`)},
		{ID: 10, EnemyID: 2, Name: "S_QUAKE", DisplayNames: datatypes.JSON(`{"en": "Quake", "th": "ดินไหว"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 30}]`)}, // ⭐️ ท่าใหม่!
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesS)
	tx.Where("enemy_id = ?", 2).Delete(&domain.EnemyAI{})
	aiRulesS := []domain.EnemyAI{
		{EnemyID: 2, Priority: 1, Condition: domain.AIConditionTurnIs, ConditionValue: 3, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesS[2].ID},       // ท่าไม้ตาย (Quake)
		{EnemyID: 2, Priority: 10, Condition: domain.AIConditionSelfHPBelow, ConditionValue: 0.5, Action: domain.AIActionUseAbility, Target: "SELF", AbilityToUseID: &abilitiesS[1].ID}, // ท่าประจำธาตุ (Harden)
		{EnemyID: 2, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesS[0].ID},                         // ท่าดาเมจ (Slap)
	}
	tx.Create(&aiRulesS)

	// ========================================================================
	// ENEMY 3: TRAINING GOLEM (LIQUIDITY) - (แก้ไข AI ที่ "พัง"!)
	// ========================================================================
	golemL := domain.Enemy{ID: 3, Name: "TRAINING_GOLEM_LIQUIDITY", DisplayNames: datatypes.JSON(`{"en": "Liquidity Golem", "th": "โกเลมวารี"}`), ElementID: 2, Level: 1, MaxHP: 220, Initiative: 45, MaxEndurance: 80}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemL)
	abilitiesL := []domain.EnemyAbility{
		{ID: 5, EnemyID: 3, Name: "L_SPLASH", DisplayNames: datatypes.JSON(`{"en": "Splash", "th": "สาดน้ำ"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 25}]`)},
		{ID: 6, EnemyID: 3, Name: "L_REGEN", DisplayNames: datatypes.JSON(`{"en": "Regenerate", "th": "ฟื้นฟู"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 100, "target": "SELF", "value": 20, "duration": 3}]`)},
		{ID: 11, EnemyID: 3, Name: "L_DROWN", DisplayNames: datatypes.JSON(`{"en": "Drown", "th": "กระแสน้ำ"}`), APCost: 3, EffectsJSON: datatypes.JSON(`[{"effect_id": 302, "target": "PLAYER", "value": 20, "duration": 2}]`)}, // ⭐️ ท่าใหม่! (Vulnerable)
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesL)
	tx.Where("enemy_id = ?", 3).Delete(&domain.EnemyAI{})
	aiRulesL := []domain.EnemyAI{
		{EnemyID: 3, Priority: 1, Condition: domain.AIConditionTurnIs, ConditionValue: 2, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesL[2].ID},       // ท่าไม้ตาย (Drown)
		{EnemyID: 3, Priority: 10, Condition: domain.AIConditionSelfHPBelow, ConditionValue: 0.5, Action: domain.AIActionUseAbility, Target: "SELF", AbilityToUseID: &abilitiesL[1].ID}, // ท่าประจำธาตุ (Regen)
		{EnemyID: 3, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesL[0].ID},                         // ท่าดาเมจ (Splash)
	}
	tx.Create(&aiRulesL)

	// ========================================================================
	// ENEMY 4: TRAINING GOLEM (TEMPO) - (เพิ่มท่าให้ครบ 3!)
	// ========================================================================
	golemG := domain.Enemy{ID: 4, Name: "TRAINING_GOLEM_TEMPO", DisplayNames: datatypes.JSON(`{"en": "Tempo Golem", "th": "โกเลมวายุ"}`), ElementID: 3, Level: 1, MaxHP: 200, Initiative: 55, MaxEndurance: 70}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemG)
	abilitiesG := []domain.EnemyAbility{
		{ID: 7, EnemyID: 4, Name: "G_WIND_SLASH", DisplayNames: datatypes.JSON(`{"en": "Wind Slash", "th": "คมลม"}`), APCost: 1, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 25}]`)},
		{ID: 8, EnemyID: 4, Name: "G_EVADE", DisplayNames: datatypes.JSON(`{"en": "Evade", "th": "หลบหลีก"}`), APCost: 2, EffectsJSON: datatypes.JSON(`[{"effect_id": 102, "target": "SELF", "value": 50, "duration": 1}]`)},
		{ID: 12, EnemyID: 4, Name: "G_TORNADO", DisplayNames: datatypes.JSON(`{"en": "Tornado", "th": "พายุหมุน"}`), APCost: 3, EffectsJSON: datatypes.JSON(`[{"effect_id": 1, "value": 50}]`)}, // ⭐️ ท่าใหม่!
	}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&abilitiesG)
	tx.Where("enemy_id = ?", 4).Delete(&domain.EnemyAI{})
	aiRulesG := []domain.EnemyAI{
		{EnemyID: 4, Priority: 1, Condition: domain.AIConditionTurnIs, ConditionValue: 3, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesG[2].ID}, // ท่าไม้ตาย (Tornado)
		{EnemyID: 4, Priority: 10, Condition: domain.AIConditionTurnIs, ConditionValue: 2, Action: domain.AIActionUseAbility, Target: "SELF", AbilityToUseID: &abilitiesG[1].ID},  // ท่าประจำธาตุ (Evade)
		{EnemyID: 4, Priority: 99, Condition: domain.AIConditionAlways, Action: domain.AIActionUseAbility, Target: "PLAYER", AbilityToUseID: &abilitiesG[0].ID},                   // ท่าดาเมจ (Wind Slash)
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
