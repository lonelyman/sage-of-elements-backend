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
		if err := seedGameConfig(tx); err != nil {
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
	log.Println("Seeding/Updating effects with new 1000-based ID structure...")
	effects := []domain.Effect{
		// === หมวด 1000: Direct Effects (กระทำโดยตรง) ===
		// --- 1100-1199: HP/MP/Resource Manipulation ---
		{ID: 1101, Name: "DAMAGE", Type: domain.EffectTypeDamage},      // 💥 สร้างความเสียหาย HP
		{ID: 1102, Name: "SHIELD", Type: domain.EffectTypeShield},      // 🛡️ สร้างโล่ (เลือดชั่วคราว)
		{ID: 1103, Name: "HEAL", Type: domain.EffectTypeHeal},          // ❤️ ฟื้นฟู HP
		{ID: 1104, Name: "MP_DAMAGE", Type: domain.EffectTypeResource}, // 💧 สร้างความเสียหาย MP

		// === หมวด 2000: Buffs (เสริมพลัง - ติดตัวเป้าหมาย) ===
		// --- 2100-2199: Regeneration Buffs ---
		{ID: 2101, Name: "BUFF_HP_REGEN", Type: domain.EffectTypeBuff}, // 💖 ฟื้นฟู HP ต่อเนื่อง
		{ID: 2102, Name: "BUFF_MP_REGEN", Type: domain.EffectTypeBuff}, // 💙 ฟื้นฟู MP ต่อเนื่อง
		// --- 2200-2299: Combat Stat Buffs ---
		{ID: 2201, Name: "BUFF_EVASION", Type: domain.EffectTypeBuff},     // 💨 เพิ่มโอกาสหลบหลีก
		{ID: 2202, Name: "BUFF_DMG_UP", Type: domain.EffectTypeBuff},      // 🔥 เพิ่มความเสียหายที่ทำ
		{ID: 2203, Name: "BUFF_RETALIATION", Type: domain.EffectTypeBuff}, // ✨ สะท้อนความเสียหาย
		{ID: 2204, Name: "BUFF_DEFENSE_UP", Type: domain.EffectTypeBuff},  // 💪 ลดความเสียหาย HP ที่ได้รับ

		// === หมวด 3000: Synergy Buffs (เสริมพลัง - เฉพาะทาง) ===
		// --- 3100-3199: Stance Buffs ---
		{ID: 3101, Name: "STANCE_S", Type: domain.EffectTypeSynergyBuff}, // 🌟 สถานะเสริมพลัง S
		{ID: 3102, Name: "STANCE_L", Type: domain.EffectTypeSynergyBuff}, // 🌟 สถานะเสริมพลัง L
		{ID: 3103, Name: "STANCE_G", Type: domain.EffectTypeSynergyBuff}, // 🌟 สถานะเสริมพลัง G
		{ID: 3104, Name: "STANCE_P", Type: domain.EffectTypeSynergyBuff}, // 🌟 สถานะเสริมพลัง P

		// === หมวด 4000: Debuffs (ลดทอน - ติดตัวเป้าหมาย) ===
		// --- 4100-4199: Stat Debuffs ---
		{ID: 4101, Name: "DEBUFF_SLOW", Type: domain.EffectTypeDebuffCC},     // 🐢 ลดค่า Initiative
		{ID: 4102, Name: "DEBUFF_VULNERABLE", Type: domain.EffectTypeDebuff}, // 🎯 ทำให้ได้รับความเสียหายแรงขึ้น
		// --- 4200-4299: Damage Over Time (DoT) Debuffs ---
		{ID: 4201, Name: "DEBUFF_IGNITE", Type: domain.EffectTypeDebuffDOT}, // 🔥 สร้างความเสียหายต่อเนื่อง (เผาไหม้)

		// === หมวด 5000+: Reserved for Future Expansion ===
		// (เช่น 5000=Utility, 6000=Crowd Control, etc.)
	}
	// ใช้ OnConflict เหมือนเดิม เพื่อให้รันซ้ำได้
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
	log.Println("Seeding/Updating spells (Updated with new 1000-based Effect IDs)...")
	spells := []domain.Spell{
		// --- S (Solidity) - เน้นป้องกัน, Debuff เบื้องต้น ---
		{ID: 1, Name: "EarthSlam", TargetType: domain.TargetTypeEnemy, ElementID: 1, MasteryID: 1, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Earth Slam", "th": "ปฐพีทุบ"},
			Descriptions: datatypes.JSONMap{"en": "Deals Solidity damage and grants S Stance.", "th": "สร้างความเสียหายปฐพีและมอบสถานะ S"},
			Effects:      []*domain.SpellEffect{{EffectID: 1101, BaseValue: 55}, {EffectID: 3101, DurationInTurns: 2}}},
		{ID: 2, Name: "StoneSkin", TargetType: domain.TargetTypeSelf, ElementID: 1, MasteryID: 2, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Stone Skin", "th": "ผิวศิลา"},
			Descriptions: datatypes.JSONMap{"en": "Creates a small shield and grants S Stance.", "th": "สร้างโล่ป้องกันเล็กน้อยและมอบสถานะ S"},
			Effects:      []*domain.SpellEffect{{EffectID: 1102, BaseValue: 200}, {EffectID: 3101, DurationInTurns: 2}}},
		{ID: 3, Name: "Reinforce", TargetType: domain.TargetTypeSelf, ElementID: 1, MasteryID: 3, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Reinforce", "th": "เสริมกำลัง"},
			Descriptions: datatypes.JSONMap{"en": "Slightly reduces incoming damage for a few turns.", "th": "ลดความเสียหายที่ได้รับเล็กน้อยชั่วขณะ"},
			Effects:      []*domain.SpellEffect{{EffectID: 2204, BaseValue: 15, DurationInTurns: 3}}},
		{ID: 4, Name: "Tremor", TargetType: domain.TargetTypeEnemy, ElementID: 1, MasteryID: 4, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Tremor", "th": "สะเทือน"},
			Descriptions: datatypes.JSONMap{"en": "Slightly slows the target's initiative.", "th": "ลดค่าความคิดริเริ่มของเป้าหมายเล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 4101, BaseValue: -15, DurationInTurns: 2}}},

		// --- L (Liquidity) - เน้นฟื้นฟูพื้นฐาน ---
		{ID: 5, Name: "AquaShot", TargetType: domain.TargetTypeEnemy, ElementID: 2, MasteryID: 1, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Aqua Shot", "th": "กระสุนวารี"},
			Descriptions: datatypes.JSONMap{"en": "Deals Liquidity damage.", "th": "สร้างความเสียหายวารี"},
			Effects:      []*domain.SpellEffect{{EffectID: 1101, BaseValue: 50}}},
		{ID: 6, Name: "SoothingMist", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 2, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Soothing Mist", "th": "หมอกบรรเทา"},
			Descriptions: datatypes.JSONMap{"en": "Applies minor HP Regeneration and grants L Stance.", "th": "มอบผลฟื้นฟู HP เล็กน้อยต่อเนื่องและมอบสถานะ L"},
			Effects:      []*domain.SpellEffect{{EffectID: 2101, BaseValue: 25, DurationInTurns: 3}, {EffectID: 3102, DurationInTurns: 3}}},
		{ID: 7, Name: "Meditate", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 3, APCost: 1, MPCost: 0,
			DisplayNames: datatypes.JSONMap{"en": "Meditate", "th": "ทำสมาธิ"},
			Descriptions: datatypes.JSONMap{"en": "Restores a small amount of MP over time.", "th": "ฟื้นฟู MP เล็กน้อยต่อเนื่อง"},
			Effects:      []*domain.SpellEffect{{EffectID: 2102, BaseValue: 10, DurationInTurns: 3}}},
		{ID: 8, Name: "MinorHeal", TargetType: domain.TargetTypeSelf, ElementID: 2, MasteryID: 2, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Minor Heal", "th": "ฟื้นฟูเล็กน้อย"},
			Descriptions: datatypes.JSONMap{"en": "Restores a small amount of HP.", "th": "ฟื้นฟู HP เล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 1103, BaseValue: 75}}},

		// --- G (Gale) - เน้นความเร็ว, ก่อกวนเบาๆ ---
		{ID: 9, Name: "WindSlash", TargetType: domain.TargetTypeEnemy, ElementID: 3, MasteryID: 1, APCost: 1, MPCost: 10,
			DisplayNames: datatypes.JSONMap{"en": "Wind Slash", "th": "ดาบลม"},
			Descriptions: datatypes.JSONMap{"en": "Deals Gale damage quickly.", "th": "สร้างความเสียหายวายุอย่างรวดเร็ว"},
			Effects:      []*domain.SpellEffect{{EffectID: 1101, BaseValue: 35}}},
		{ID: 10, Name: "Blur", TargetType: domain.TargetTypeSelf, ElementID: 3, MasteryID: 2, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Blur", "th": "พร่ามัว"},
			Descriptions: datatypes.JSONMap{"en": "Slightly increases Evasion for one turn.", "th": "เพิ่มอัตราหลบหลีกเล็กน้อยชั่วขณะ"},
			Effects:      []*domain.SpellEffect{{EffectID: 2201, BaseValue: 40, DurationInTurns: 1}}},
		{ID: 11, Name: "SwiftStep", TargetType: domain.TargetTypeSelf, ElementID: 3, MasteryID: 3, APCost: 1, MPCost: 10,
			DisplayNames: datatypes.JSONMap{"en": "Swift Step", "th": "ก้าววายุ"},
			Descriptions: datatypes.JSONMap{"en": "Grants G Stance.", "th": "มอบสถานะ G"},
			Effects:      []*domain.SpellEffect{{EffectID: 3103, DurationInTurns: 2}}},
		{ID: 12, Name: "Gust", TargetType: domain.TargetTypeEnemy, ElementID: 3, MasteryID: 4, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Gust", "th": "ลมกระโชก"},
			Descriptions: datatypes.JSONMap{"en": "Moderately slows the target's initiative.", "th": "ลดค่าความคิดริเริ่มของเป้าหมายปานกลาง"},
			Effects:      []*domain.SpellEffect{{EffectID: 4101, BaseValue: -25, DurationInTurns: 2}}},

		// --- P (Plasma) - เน้นดาเมจพื้นฐาน, บัฟ/ดีบัฟเบาๆ ---
		{ID: 13, Name: "PlasmaBolt", TargetType: domain.TargetTypeEnemy, ElementID: 4, MasteryID: 1, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Plasma Bolt", "th": "กระสุนพลาสมา"},
			Descriptions: datatypes.JSONMap{"en": "Deals Plasma damage.", "th": "สร้างความเสียหายพลาสมา"},
			Effects:      []*domain.SpellEffect{{EffectID: 1101, BaseValue: 65}}},
		{ID: 14, Name: "StaticField", TargetType: domain.TargetTypeSelf, ElementID: 4, MasteryID: 2, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Static Field", "th": "สนามไฟฟ้าสถิต"},
			Descriptions: datatypes.JSONMap{"en": "Creates a weak shield that slightly damages attackers.", "th": "สร้างโล่เบาบางที่สะท้อนความเสียหายเล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 1102, BaseValue: 150, DurationInTurns: 2}, {EffectID: 2203, BaseValue: 10, DurationInTurns: 2}}},
		{ID: 15, Name: "Empower", TargetType: domain.TargetTypeSelf, ElementID: 4, MasteryID: 3, APCost: 1, MPCost: 10,
			DisplayNames: datatypes.JSONMap{"en": "Empower", "th": "เสริมพลัง"},
			Descriptions: datatypes.JSONMap{"en": "Slightly increases damage next turn and grants P Stance.", "th": "เพิ่มความเสียหายเล็กน้อยในเทิร์นถัดไปและมอบสถานะ P"},
			Effects:      []*domain.SpellEffect{{EffectID: 2202, BaseValue: 15, DurationInTurns: 1}, {EffectID: 3104, DurationInTurns: 1}}},
		{ID: 16, Name: "Analyze", TargetType: domain.TargetTypeEnemy, ElementID: 4, MasteryID: 4, APCost: 2, MPCost: 15,
			DisplayNames: datatypes.JSONMap{"en": "Analyze", "th": "วิเคราะห์"},
			Descriptions: datatypes.JSONMap{"en": "Makes the target take slightly increased damage.", "th": "ทำให้เป้าหมายได้รับความเสียหายเพิ่มขึ้นเล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 4102, BaseValue: 10, DurationInTurns: 2}}},

		// --- Tier 1 Spells - ทำให้เบาลง ให้พอเห็นความต่าง แต่ไม่โกง ---
		{ID: 17, Name: "EntanglingRoots", TargetType: domain.TargetTypeEnemy, ElementID: 5, MasteryID: 4, APCost: 2, MPCost: 20,
			DisplayNames: datatypes.JSONMap{"en": "Entangling Roots", "th": "รากไม้พันธนาการ"},
			Descriptions: datatypes.JSONMap{"en": "Greatly slows the target for a short duration.", "th": "ลดค่าความคิดริเริ่มเป้าหมายอย่างมากชั่วขณะ"},
			Effects:      []*domain.SpellEffect{{EffectID: 4101, BaseValue: -40, DurationInTurns: 1}}},
		{ID: 18, Name: "ManaBurn", TargetType: domain.TargetTypeEnemy, ElementID: 5, MasteryID: 4, APCost: 2, MPCost: 25,
			DisplayNames: datatypes.JSONMap{"en": "Mana Burn", "th": "เผาผลาญมานา"},
			Descriptions: datatypes.JSONMap{"en": "Damages the target's MP.", "th": "สร้างความเสียหายแก่ MP ของเป้าหมาย"},
			Effects:      []*domain.SpellEffect{{EffectID: 1104, BaseValue: 30}}},
		{ID: 21, Name: "Fireball", TargetType: domain.TargetTypeEnemy, ElementID: 7, MasteryID: 1, APCost: 2, MPCost: 25,
			DisplayNames: datatypes.JSONMap{"en": "Fireball", "th": "ลูกไฟ"},
			Descriptions: datatypes.JSONMap{"en": "Deals significant damage and applies a minor Burn.", "th": "สร้างความเสียหายรุนแรงและติดสถานะเผาไหม้เล็กน้อย"},
			Effects:      []*domain.SpellEffect{{EffectID: 1101, BaseValue: 70}, {EffectID: 4201, BaseValue: 10, DurationInTurns: 2}}},
	}

	// ⚠️ ลบ spell_effects ก่อน spells เพื่อหลีกเลี่ยง foreign key constraint
	if err := tx.Exec("DELETE FROM spell_effects").Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM spells").Error; err != nil {
		return err
	}

	// ✅ Insert spells พร้อม spell_effects
	return tx.Create(&spells).Error
}

func seedGameConfig(tx *gorm.DB) error {
	log.Println("Seeding/Updating game_configs...")
	configs := []domain.GameConfig{
		// Talent (ค่าพลังดิบ)
		{Key: "TALENT_BASE_ALLOCATION", Value: "3"},
		{Key: "TALENT_PRIMARY_ALLOCATION", Value: "90"},

		// Core Stats
		{Key: "STAT_HP_BASE", Value: "100"},
		{Key: "STAT_HP_PER_TALENT_S", Value: "2"},
		{Key: "STAT_MP_BASE", Value: "50"},
		{Key: "STAT_MP_PER_TALENT_L", Value: "1"},
		{Key: "STAT_MP_REGEN_PER_TURN", Value: "5"},
		{Key: "STAT_INITIATIVE_BASE_MIN", Value: "40"},
		{Key: "STAT_INITIATIVE_BASE_MAX", Value: "60"},
		{Key: "STAT_INITIATIVE_PER_TALENT_G", Value: "0.5"},

		// Combat System
		{Key: "MASTERY_ATTACK_MODIFIER", Value: "1.15"},
		{Key: "MASTERY_HEAL_MODIFIER", Value: "1.10"},
		{Key: "MASTERY_DEFENSE_MODIFIER", Value: "1.05"},
		{Key: "MASTERY_SUPPORT_MODIFIER", Value: "1.12"},
		{Key: "ELEMENT_ADVANTAGE_MULTIPLIER", Value: "1.30"},
		{Key: "ELEMENT_DISADVANTAGE_MULTIPLIER", Value: "0.80"},
		{Key: "COMBAT_TURN_TIMEOUT", Value: "60"},
		{Key: "COMBAT_MATCH_TIMEOUT", Value: "1800"},

		// Regeneration
		{Key: "PASSIVE_HP_REGEN_PER_MINUTE", Value: "0"},
		{Key: "PASSIVE_MP_REGEN_PER_MINUTE", Value: "0"},

		// Tutorial
		{Key: "TUTORIAL_TOTAL_STEPS", Value: "4"},

		// Fusion Tutorial
		{Key: "TUTORIAL_FUSION_OUTPUT", Value: "5"},
		{Key: "TUTORIAL_FUSION_AMOUNT", Value: "10"},

		// Experience & Progression
		{Key: "EXP_TRAINING_MATCH", Value: "50"},
		{Key: "EXP_STORY_MATCH", Value: "100"},
		{Key: "EXP_PVP_MATCH", Value: "150"},
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&configs).Error
}

func seedEnemies(tx *gorm.DB) error {
	log.Println("Seeding/Updating enemies and their AI (Updated with new 1000-based Effect IDs)...")

	// ========================================================================
	// ENEMY 1: TRAINING GOLEM (POTENCY)
	// ========================================================================
	golemP := domain.Enemy{ID: 1, Name: "TRAINING_GOLEM_POTENCY", DisplayNames: datatypes.JSON(`{"en": "Potency Golem", "th": "โกเลมพลังงาน"}`), ElementID: 4, Level: 1, MaxHP: 250, Initiative: 40}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemP)
	abilitiesP := []domain.EnemyAbility{
		// ⭐️ ท่าพื้นฐาน (Punch) - ไม่ใช้ MP
		{ID: 1, EnemyID: 1, Name: "P_PUNCH", DisplayNames: datatypes.JSON(`{"en": "Punch", "th": "หมัดตรง"}`), APCost: 1, MPCost: 0, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 20}]`)},
		// ⭐️ ท่าไม้ตาย (Tremor) - ใช้ MP 10
		{ID: 2, EnemyID: 1, Name: "P_TREMOR", DisplayNames: datatypes.JSON(`{"en": "Tremor", "th": "คลื่นพลัง"}`), APCost: 3, MPCost: 10, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 40}, {"effect_id": 4101, "value": -10, "duration": 2}]`)},
		// ⭐️ ท่าบัฟ (Overcharge) - ใช้ MP 5
		{ID: 9, EnemyID: 1, Name: "P_OVERCHARGE", DisplayNames: datatypes.JSON(`{"en": "Overcharge", "th": "ปลุกพลัง"}`), APCost: 2, MPCost: 5, EffectsJSON: datatypes.JSON(`[{"effect_id": 2202, "target": "SELF", "value": 30, "duration": 2}]`)},
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
	// ENEMY 2: TRAINING GOLEM (SOLIDITY)
	// ========================================================================
	golemS := domain.Enemy{ID: 2, Name: "TRAINING_GOLEM_SOLIDITY", DisplayNames: datatypes.JSON(`{"en": "Solidity Golem", "th": "โกเลมศิลา"}`), ElementID: 1, Level: 1, MaxHP: 300, Initiative: 35}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemS)
	abilitiesS := []domain.EnemyAbility{
		// ⭐️ ท่าพื้นฐาน (Slap) - ไม่ใช้ MP
		{ID: 3, EnemyID: 2, Name: "S_SLAP", DisplayNames: datatypes.JSON(`{"en": "Slap", "th": "ตบ"}`), APCost: 1, MPCost: 0, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 15}]`)},
		// ⭐️ ท่าบัฟ (Harden) - ใช้ MP 5
		{ID: 4, EnemyID: 2, Name: "S_HARDEN", DisplayNames: datatypes.JSON(`{"en": "Harden", "th": "กายาหิน"}`), APCost: 2, MPCost: 5, EffectsJSON: datatypes.JSON(`[{"effect_id": 2204, "target": "SELF", "duration": 2}]`)},
		// ⭐️ ท่าไม้ตาย (Quake) - ใช้ MP 10
		{ID: 10, EnemyID: 2, Name: "S_QUAKE", DisplayNames: datatypes.JSON(`{"en": "Quake", "th": "ดินไหว"}`), APCost: 2, MPCost: 10, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 30}]`)},
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
	// ENEMY 3: TRAINING GOLEM (LIQUIDITY)
	// ========================================================================
	golemL := domain.Enemy{ID: 3, Name: "TRAINING_GOLEM_LIQUIDITY", DisplayNames: datatypes.JSON(`{"en": "Liquidity Golem", "th": "โกเลมวารี"}`), ElementID: 2, Level: 1, MaxHP: 220, Initiative: 45}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemL)
	abilitiesL := []domain.EnemyAbility{
		// ⭐️ ท่าพื้นฐาน (Splash) - ใช้ MP 5 (เพราะมัน 2 AP)
		{ID: 5, EnemyID: 3, Name: "L_SPLASH", DisplayNames: datatypes.JSON(`{"en": "Splash", "th": "สาดน้ำ"}`), APCost: 2, MPCost: 5, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 25}]`)},
		// ⭐️ ท่าบัฟ (Regen) - ใช้ MP 10
		{ID: 6, EnemyID: 3, Name: "L_REGEN", DisplayNames: datatypes.JSON(`{"en": "Regenerate", "th": "ฟื้นฟู"}`), APCost: 2, MPCost: 10, EffectsJSON: datatypes.JSON(`[{"effect_id": 2101, "target": "SELF", "value": 20, "duration": 3}]`)},
		// ⭐️ ท่าไม้ตาย (Drown) - ใช้ MP 15
		{ID: 11, EnemyID: 3, Name: "L_DROWN", DisplayNames: datatypes.JSON(`{"en": "Drown", "th": "กระแสน้ำ"}`), APCost: 3, MPCost: 15, EffectsJSON: datatypes.JSON(`[{"effect_id": 4102, "target": "PLAYER", "value": 20, "duration": 2}]`)},
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
	// ENEMY 4: TRAINING GOLEM (TEMPO)
	// ========================================================================
	golemG := domain.Enemy{ID: 4, Name: "TRAINING_GOLEM_TEMPO", DisplayNames: datatypes.JSON(`{"en": "Tempo Golem", "th": "โกเลมวายุ"}`), ElementID: 3, Level: 1, MaxHP: 200, Initiative: 55}
	tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&golemG)
	abilitiesG := []domain.EnemyAbility{
		// ⭐️ ท่าพื้นฐาน (Wind Slash) - ไม่ใช้ MP (เพราะ 1 AP)
		{ID: 7, EnemyID: 4, Name: "G_WIND_SLASH", DisplayNames: datatypes.JSON(`{"en": "Wind Slash", "th": "คมลม"}`), APCost: 1, MPCost: 0, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 25}]`)},
		// ⭐️ ท่าบัฟ (Evade) - ใช้ MP 5
		{ID: 8, EnemyID: 4, Name: "G_EVADE", DisplayNames: datatypes.JSON(`{"en": "Evade", "th": "หลบหลีก"}`), APCost: 2, MPCost: 5, EffectsJSON: datatypes.JSON(`[{"effect_id": 2201, "target": "SELF", "value": 50, "duration": 1}]`)},
		// ⭐️ ท่าไม้ตาย (Tornado) - ใช้ MP 15
		{ID: 12, EnemyID: 4, Name: "G_TORNADO", DisplayNames: datatypes.JSON(`{"en": "Tornado", "th": "พายุหมุน"}`), APCost: 3, MPCost: 15, EffectsJSON: datatypes.JSON(`[{"effect_id": 1101, "value": 50}]`)},
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
