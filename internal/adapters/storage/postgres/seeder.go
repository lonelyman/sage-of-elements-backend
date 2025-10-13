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

	// ใช้ Transaction เพื่อให้แน่ใจว่าทุกอย่างสำเร็จหรือล้มเหลวไปพร้อมกัน
	return db.Transaction(func(tx *gorm.DB) error {

		// --- 1. Masteries ---
		if err := seedMasteries(tx); err != nil {
			return err
		}

		// --- 2. Elements ---
		if err := seedElements(tx); err != nil {
			return err
		}

		// --- 3. Effects ---
		if err := seedEffects(tx); err != nil {
			return err
		}

		// --- 4. Recipes ---
		if err := seedRecipes(tx); err != nil {
			return err
		}

		// --- 5. Spells (The final boss!) ---
		if err := seedSpells(tx); err != nil {
			return err
		}

		// --- 6. GameConfigs ---
		if err := seedGameConfigs(tx); err != nil {
			return err
		}
		// --- 6. GameConfigs ---
		if err := seedEnemies(tx); err != nil {
			log.Printf("Failed to seed enemies: %v", err)
			return err
		}

		log.Println("Database seeding process finished successfully.")
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
	// ในฟังก์ชัน Seed(), ส่วนที่ 3
	log.Println("Seeding/Updating effects...")
	effects := []domain.Effect{
		// ID Range 1-99: Basic Actions & Utilities
		{ID: 1, Name: "DAMAGE", Type: domain.EffectTypeDamage,
			DisplayNames: datatypes.JSONMap{"en": "Damage", "th": "ความเสียหาย"},
			Descriptions: datatypes.JSONMap{"en": "Deals direct damage to the target.", "th": "สร้างความเสียหายโดยตรงแก่เป้าหมาย"},
		},
		{ID: 2, Name: "SHIELD", Type: domain.EffectTypeShield,
			DisplayNames: datatypes.JSONMap{"en": "Shield", "th": "โล่ป้องกัน"},
			Descriptions: datatypes.JSONMap{"en": "Creates a temporary shield that absorbs damage.", "th": "สร้างเกราะชั่วคราวที่ดูดซับความเสียหาย"},
		},
		{ID: 3, Name: "HEAL", Type: domain.EffectTypeHeal,
			DisplayNames: datatypes.JSONMap{"en": "Heal", "th": "ฟื้นฟูพลังชีวิต"},
			Descriptions: datatypes.JSONMap{"en": "Restores the target's HP.", "th": "ฟื้นฟู HP ให้กับเป้าหมาย"},
		},
		{ID: 4, Name: "TRUE_DAMAGE", Type: domain.EffectTypeTrueDamage,
			DisplayNames: datatypes.JSONMap{"en": "True Damage", "th": "ความเสียหายจริง"},
			Descriptions: datatypes.JSONMap{"en": "Deals damage that ignores all defenses.", "th": "สร้างความเสียหายที่เมินเกราะป้องกันทุกชนิด"},
		},
		{ID: 5, Name: "DRAIN_MP", Type: domain.EffectTypeResource,
			DisplayNames: datatypes.JSONMap{"en": "MP Drain", "th": "ดูดพลังเวท"},
			Descriptions: datatypes.JSONMap{"en": "Reduces the target's MP.", "th": "ลด MP ของเป้าหมาย"},
		},
		{ID: 6, Name: "GAIN_AP", Type: domain.EffectTypeResource,
			DisplayNames: datatypes.JSONMap{"en": "Gain AP", "th": "เร่งจังหวะ"},
			Descriptions: datatypes.JSONMap{"en": "Grants additional AP.", "th": "ได้รับ AP เพิ่มเติม"},
		},
		{ID: 7, Name: "CLEANSE", Type: domain.EffectTypeUtility,
			DisplayNames: datatypes.JSONMap{"en": "Cleanse", "th": "ชำระล้าง"},
			Descriptions: datatypes.JSONMap{"en": "Removes debuffs from the target.", "th": "ลบล้างสถานะผิดปกติออกจากเป้าหมาย"},
		},

		// ID Range 100-199: Buffs
		{ID: 100, Name: "BUFF_HP_REGEN", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "HP Regen", "th": "ฟื้นฟูต่อเนื่อง"},
			Descriptions: datatypes.JSONMap{"en": "Restores HP at the start of the turn.", "th": "ฟื้นฟู HP เมื่อเริ่มต้นเทิร์น"},
		},
		{ID: 101, Name: "BUFF_MP_REGEN", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "MP Regen", "th": "ฟื้นฟูมานาต่อเนื่อง"},
			Descriptions: datatypes.JSONMap{"en": "Restores MP at the start of the turn.", "th": "ฟื้นฟู MP เมื่อเริ่มต้นเทิร์น"},
		},
		{ID: 102, Name: "BUFF_EVASION", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "Evasion", "th": "หลบหลีก"},
			Descriptions: datatypes.JSONMap{"en": "Grants a chance to evade physical attacks.", "th": "มอบโอกาสในการหลบหลีกการโจมตีกายภาพ"},
		},
		{ID: 103, Name: "BUFF_DMG_UP", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "Damage Up", "th": "เพิ่มพลังโจมตี"},
			Descriptions: datatypes.JSONMap{"en": "Increases the damage of the next attack.", "th": "เพิ่มความเสียหายของการโจมตีครั้งถัดไป"},
		},
		{ID: 104, Name: "BUFF_RETALIATION", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "Retaliation", "th": "หนามสะท้อน"},
			Descriptions: datatypes.JSONMap{"en": "Reflects a portion of damage taken back to the attacker.", "th": "สะท้อนความเสียหายส่วนหนึ่งกลับไปหาผู้โจมตี"},
		},
		{ID: 105, Name: "BUFF_MAX_HP", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "Max HP Up", "th": "เพิ่มพลังชีวิตสูงสุด"},
			Descriptions: datatypes.JSONMap{"en": "Temporarily increases maximum HP.", "th": "เพิ่มพลังชีวิตสูงสุดชั่วคราว"},
		},
		{ID: 106, Name: "BUFF_CC_RESIST", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "CC Resist", "th": "ต้านทาน CC"},
			Descriptions: datatypes.JSONMap{"en": "Grants resistance to crowd control effects.", "th": "มอบความต้านทานต่อเอฟเฟกต์ประเภทขัดขวาง"},
		},
		{ID: 108, Name: "BUFF_PENETRATION", Type: domain.EffectTypeBuff,
			DisplayNames: datatypes.JSONMap{"en": "Penetration", "th": "ทะลุเกราะ"},
			Descriptions: datatypes.JSONMap{"en": "Allows the next attack to ignore a percentage of the target's defenses.", "th": "ทำให้การโจมตีครั้งต่อไปเมินพลังป้องกันของเป้าหมายเป็น %"},
		},

		// ID Range 150-199: Synergy Buffs
		{ID: 200, Name: "SYNERGY_GRANT_STANCE_S", Type: domain.EffectTypeSynergyBuff,
			DisplayNames: datatypes.JSONMap{"en": "Stance", "th": "มั่นคงดั่งขุนเขา"},
			Descriptions: datatypes.JSONMap{"en": "A defensive stance that enhances Solidity spells.", "th": "สถานะตั้งรับที่เสริมความสามารถให้เวทสายสถิตยภาพ"},
		},
		{ID: 201, Name: "SYNERGY_GRANT_FLOW_L", Type: domain.EffectTypeSynergyBuff,
			DisplayNames: datatypes.JSONMap{"en": "Flow", "th": "เปี่ยมล้นด้วยชีวา"},
			Descriptions: datatypes.JSONMap{"en": "A regenerative state that enhances Liquidity spells.", "th": "สถานะฟื้นฟูที่เสริมความสามารถให้เวทสายสภาพคล่อง"},
		},
		{ID: 202, Name: "SYNERGY_GRANT_TEMPO_G", Type: domain.EffectTypeSynergyBuff,
			DisplayNames: datatypes.JSONMap{"en": "Tempo", "th": "จังหวะลื่นไหล"},
			Descriptions: datatypes.JSONMap{"en": "A state of heightened reflexes that enhances Tempo spells.", "th": "สถานะตื่นตัวที่เสริมความสามารถให้เวทสายจังหวะ"},
		},
		{ID: 203, Name: "SYNERGY_GRANT_OVERCHARGE_P", Type: domain.EffectTypeSynergyBuff,
			DisplayNames: datatypes.JSONMap{"en": "Overcharge", "th": "พลังงานปะทุ"},
			Descriptions: datatypes.JSONMap{"en": "An overcharged state that enhances Potency spells.", "th": "สถานะพลังงานปะทุที่เสริมความสามารถให้เวทสายพลังงาน"},
		},

		// ID Range 200-299: Debuffs
		{ID: 300, Name: "DEBUFF_REDUCE_ARMOR", Type: domain.EffectTypeDebuff,
			DisplayNames: datatypes.JSONMap{"en": "Armor Break", "th": "ลดเกราะ"},
			Descriptions: datatypes.JSONMap{"en": "Reduces the target's defenses.", "th": "ลดพลังป้องกันของเป้าหมาย"},
		},
		{ID: 301, Name: "DEBUFF_SLOW", Type: domain.EffectTypeDebuffCC,
			DisplayNames: datatypes.JSONMap{"en": "Slow", "th": "เชื่องช้า"},
			Descriptions: datatypes.JSONMap{"en": "Reduces the target's Initiative.", "th": "ลดค่า Initiative ของเป้าหมาย"},
		},
		{ID: 302, Name: "DEBUFF_VULNERABLE", Type: domain.EffectTypeDebuff,
			DisplayNames: datatypes.JSONMap{"en": "Vulnerable", "th": "เปราะบาง"},
			Descriptions: datatypes.JSONMap{"en": "Increases all damage taken by the target.", "th": "ทำให้เป้าหมายได้รับความเสียหายทุกประเภทแรงขึ้น"},
		},
		{ID: 303, Name: "DEBUFF_ROOT", Type: domain.EffectTypeDebuffCC,
			DisplayNames: datatypes.JSONMap{"en": "Root", "th": "ตรึง"},
			Descriptions: datatypes.JSONMap{"en": "Prevents the target from using movement abilities.", "th": "ขัดขวางไม่ให้เป้าหมายใช้ความสามารถประเภทเคลื่อนที่"},
		},
		{ID: 304, Name: "DEBUFF_AP_DRAIN", Type: domain.EffectTypeDebuff,
			DisplayNames: datatypes.JSONMap{"en": "AP Drain", "th": "ดูดจังหวะ"},
			Descriptions: datatypes.JSONMap{"en": "Reduces the target's AP regeneration next turn.", "th": "ลดการฟื้นฟู AP ของเป้าหมายในเทิร์นถัดไป"},
		},
		{ID: 305, Name: "DEBUFF_STUN", Type: domain.EffectTypeDebuffHardCC,
			DisplayNames: datatypes.JSONMap{"en": "Stun", "th": "มึนงง"},
			Descriptions: datatypes.JSONMap{"en": "Prevents the target from taking any action.", "th": "ขัดขวางไม่ให้เป้าหมายสามารถกระทำการใดๆ ได้"},
		},
		{ID: 306, Name: "DEBUFF_IGNITE", Type: domain.EffectTypeDebuffDOT,
			DisplayNames: datatypes.JSONMap{"en": "Ignite", "th": "เผาไหม้"},
			Descriptions: datatypes.JSONMap{"en": "Deals damage over time.", "th": "สร้างความเสียหายต่อเนื่อง"},
		},
		{ID: 308, Name: "DEBUFF_CORROSION", Type: domain.EffectTypeDebuffDOT,
			DisplayNames: datatypes.JSONMap{"en": "Corrosion", "th": "กัดกร่อน"},
			Descriptions: datatypes.JSONMap{"en": "Deals damage over time and reduces defenses.", "th": "สร้างความเสียหายต่อเนื่องและลดพลังป้องกัน"},
		},
	}
	// ของใหม่ที่ฉลาดกว่า!
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}}, // <-- บอก GORM ให้เช็คความซ้ำที่คอลัมน์ "name"
		UpdateAll: true,                            // <-- ถ้าซ้ำ ก็ให้อัปเดตคอลัมน์อื่นๆ ทั้งหมด
	}).Create(&effects).Error; err != nil {
		log.Printf("❌ ERROR during effects seeding: %v\n", err)

		return err
	}
	return nil
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
	// --- 5. เพาะเมล็ดพันธุ์ "Spells" ---
	log.Println("Seeding/Updating spells...")
	spells := []domain.Spell{
		// --- หมวดที่ 1: เวทมนตร์ธาตุปฐมภูมิ (16 Spells) ---
		// S
		{
			ID:           1,
			Name:         "EarthSlam",
			DisplayNames: datatypes.JSONMap{"en": "Earth Slam", "th": "พสุธากัมปนาท"},
			Descriptions: datatypes.JSONMap{"en": "Slams the target with the power of the earth, dealing damage and potentially stunning them.", "th": "ทุบเป้าหมายด้วยพลังแห่งพสุธา สร้างความเสียหายและมีโอกาสทำให้มึนงง"},
			ElementID:    1,
			MasteryID:    1,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 60},
				{EffectID: 200, BaseValue: 10, DurationInTurns: 2},
				{EffectID: 201, ConditionType: domain.ConditionTypeSelfHasBuff, ConditionDetails: "SYNERGY_GRANT_STANCE_S"},
			},
		},
		{
			ID:           2,
			Name:         "StoneFortress",
			DisplayNames: datatypes.JSONMap{"en": "Stone Fortress", "th": "ป้อมปราการศิลา"},
			Descriptions: datatypes.JSONMap{"en": "Forms a protective barrier of solid rock, absorbing incoming damage.", "th": "สร้างเกราะป้องกันจากหินแข็งแกร่ง เพื่อดูดซับความเสียหายที่เข้ามา"},
			ElementID:    1,
			MasteryID:    2,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 2, BaseValue: 450},
				{EffectID: 200, DurationInTurns: 2},
			},
		},
		{
			ID:           3,
			Name:         "IronBody",
			DisplayNames: datatypes.JSONMap{"en": "Iron Body", "th": "กายาเหล็กไหล"},
			Descriptions: datatypes.JSONMap{"en": "Hardens the caster's body to be as tough as iron, increasing their physical defense.", "th": "ทำให้ร่างกายของผู้ร่ายแข็งแกร่งดุจเหล็กไหล เพิ่มพลังป้องกันกายภาพ"},
			ElementID:    1,
			MasteryID:    3,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 105, BaseValue: 20, DurationInTurns: 3},
			},
		},
		{
			ID:           4,
			Name:         "GravityWave",
			DisplayNames: datatypes.JSONMap{"en": "Gravity Wave", "th": "คลื่นแรงโน้มถ่วง"},
			Descriptions: datatypes.JSONMap{"en": "Unleashes a wave of gravitational force, slowing down enemies.", "th": "ปล่อยคลื่นแรงโน้มถ่วงออกไปเพื่อลดความเร็วของศัตรู"},
			ElementID:    1,
			MasteryID:    4,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 301, BaseValue: -20, DurationInTurns: 2},
			},
		},
		// L

		{
			ID:           5,
			Name:         "WaterWhip",
			DisplayNames: datatypes.JSONMap{"en": "Water Whip", "th": "แส้ชลธี"},
			Descriptions: datatypes.JSONMap{"en": "Lashes out at the enemy with a whip made of pure water.", "th": "ฟาดฟันศัตรูด้วยแส้ที่สร้างจากสายน้ำบริสุทธิ์"},
			ElementID:    2,
			MasteryID:    1,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 50},
				{EffectID: 5, BaseValue: 15},
			},
		},
		{
			ID:           6,
			Name:         "HealingTide",
			DisplayNames: datatypes.JSONMap{"en": "Healing Tide", "th": "วารีบำบัด"},
			Descriptions: datatypes.JSONMap{"en": "Summons a gentle tide of healing water that restores health over time.", "th": "อัญเชิญกระแสธาราแห่งการเยียวยา เพื่อฟื้นฟูพลังชีวิตอย่างต่อเนื่อง"},
			ElementID:    2,
			MasteryID:    2,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 100, BaseValue: 50, DurationInTurns: 3},
				{EffectID: 201, DurationInTurns: 3}, // มอบบัฟ Synergy (Condition: NONE)
			},
		},
		{
			ID:           7,
			Name:         "FocusWill",
			DisplayNames: datatypes.JSONMap{"en": "Focus Will", "th": "หลอมรวมสมาธิ"},
			Descriptions: datatypes.JSONMap{"en": "The caster focuses their will, channeling inner power to increase their magical attack.", "th": "ผู้ร่ายรวบรวมสมาธิเพื่อปลุกพลังที่อยู่ภายใน เพิ่มพลังโจมตีเวทมนตร์"},
			ElementID:    2,
			MasteryID:    3,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 101, BaseValue: 10, DurationInTurns: 3},
			},
		},
		{
			ID:           8,
			Name:         "Cleanse",
			DisplayNames: datatypes.JSONMap{"en": "Cleanse", "th": "ธาราชำระล้าง"},
			Descriptions: datatypes.JSONMap{"en": "Washes away negative effects with pure water, cleansing debuffs from the target.", "th": "ชำระล้างผลกระทบด้านลบด้วยสายน้ำอันบริสุทธิ์ ลบล้างดีบัฟทั้งหมดออกจากเป้าหมาย"},
			ElementID:    2,
			MasteryID:    4,
			APCost:       1,
			MPCost:       10,
			Effects: []*domain.SpellEffect{
				{EffectID: 7, BaseValue: 1},
				{EffectID: 3, BaseValue: 100},
			},
		},
		// G
		{
			ID:           9,
			Name:         "WindBlade",
			DisplayNames: datatypes.JSONMap{"en": "Wind Blade", "th": "ดาบสายลม"},
			Descriptions: datatypes.JSONMap{"en": "Forms a sharp blade of wind to slice through the enemy.", "th": "สร้างใบมีดสายลมอันคมกริบเพื่อฟาดฟันศัตรู"},
			ElementID:    3,
			MasteryID:    1,
			APCost:       1,
			MPCost:       10,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 40},
			},
		},
		{
			ID:           10,
			Name:         "Mirage",
			DisplayNames: datatypes.JSONMap{"en": "Mirage", "th": "ภาพลวงตา"},
			Descriptions: datatypes.JSONMap{"en": "Creates an illusionary double, making the caster harder to hit.", "th": "สร้างภาพลวงตาขึ้นมา ทำให้ผู้ร่ายถูกโจมตีได้ยากขึ้น"},
			ElementID:    3,
			MasteryID:    2,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 102, BaseValue: 50, DurationInTurns: 1},
			},
		},
		{
			ID:           11,
			Name:         "TempoSurge",
			DisplayNames: datatypes.JSONMap{"en": "Tempo Surge", "th": "คลื่นเร่งจังหวะ"},
			Descriptions: datatypes.JSONMap{"en": "Rides a surge of wind to act faster, gaining extra Action Points.", "th": "โต้คลื่นสายลมเพื่อทำให้行動ได้รวดเร็วยิ่งขึ้น ได้รับ Action Points เพิ่มเติม"},
			ElementID:    3,
			MasteryID:    3,
			APCost:       1,
			MPCost:       10,
			Effects: []*domain.SpellEffect{
				{EffectID: 6, BaseValue: 1},
				{EffectID: 202, DurationInTurns: 1},
			},
		},
		{
			ID:           12,
			Name:         "HinderingWind",
			DisplayNames: datatypes.JSONMap{"en": "Hindering Wind", "th": "ลมหน่วง"},
			Descriptions: datatypes.JSONMap{"en": "Summons a strong headwind to slow down the target.", "th": "เรียกพายุลมต้านที่รุนแรงเพื่อทำให้เป้าหมายช้าลง"},
			ElementID:    3,
			MasteryID:    4,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 301, BaseValue: -30, DurationInTurns: 2},
			},
		},
		// P
		{
			ID:           13,
			Name:         "PureBeam",
			DisplayNames: datatypes.JSONMap{"en": "Pure Beam", "th": "ลำแสงบริสุทธิ์"},
			Descriptions: datatypes.JSONMap{"en": "Fires a concentrated beam of pure energy that bypasses all defenses.", "th": "ยิงลำแสงพลังงานบริสุทธิ์เข้มข้นที่สามารถทะลุทะลวงการป้องกันทั้งหมดได้"},
			ElementID:    4,
			MasteryID:    1,
			APCost:       2,
			MPCost:       20,
			Effects: []*domain.SpellEffect{
				{EffectID: 4, BaseValue: 80},
			},
		},
		{
			ID:           14,
			Name:         "EnergyThorns",
			DisplayNames: datatypes.JSONMap{"en": "Energy Thorns", "th": "เกราะหนามพลังงาน"},
			Descriptions: datatypes.JSONMap{"en": "Erects a barrier of energy that damages attackers upon contact.", "th": "สร้างเกราะพลังงานที่จะสะท้อนความเสียหายกลับไปยังผู้ที่โจมตี"},
			ElementID:    4,
			MasteryID:    2,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 2, BaseValue: 250},
				{EffectID: 104, BaseValue: 25},
			},
		},
		{
			ID:           15,
			Name:         "AwakenPower",
			DisplayNames: datatypes.JSONMap{"en": "Awaken Power", "th": "ปลุกพลังแฝง"},
			Descriptions: datatypes.JSONMap{"en": "Unlocks the caster's hidden potential, drastically increasing their critical power.", "th": "ปลดปล่อยศักยภาพที่ซ่อนอยู่ของผู้ร่าย เพิ่มพลังโจมตีคริติคอลอย่างมหาศาล"},
			ElementID:    4,
			MasteryID:    3,
			APCost:       1,
			MPCost:       10,
			Effects: []*domain.SpellEffect{
				{EffectID: 103, BaseValue: 30, DurationInTurns: 1},
				{EffectID: 203, DurationInTurns: 1},
			},
		},
		{
			ID:           16,
			Name:         "ExposeWeakness",
			DisplayNames: datatypes.JSONMap{"en": "Expose Weakness", "th": "เปิดโปงจุดอ่อน"},
			Descriptions: datatypes.JSONMap{"en": "Exposes a critical weakness, causing the target to take increased damage.", "th": "เปิดโปงจุดอ่อนที่สำคัญของเป้าหมาย ทำให้ได้รับความเสียหายเพิ่มขึ้น"},
			ElementID:    4,
			MasteryID:    4,
			APCost:       2,
			MPCost:       15,
			Effects: []*domain.SpellEffect{
				{EffectID: 302, BaseValue: 20, DurationInTurns: 2},
			},
		},

		// --- หมวดที่ 2: เวทมนตร์ธาตุพันธะ (22 Signature Spells) ---
		{
			ID:           17,
			Name:         "MudPrison",
			DisplayNames: datatypes.JSONMap{"en": "Mud Prison", "th": "คุกโคลน"},
			Descriptions: datatypes.JSONMap{"en": "Traps the target in a prison of thick mud, rendering them unable to move.", "th": "กักขังเป้าหมายไว้ในคุกโคลนเหนียวหนืด ทำให้ไม่สามารถเคลื่อนไหวได้"},
			ElementID:    5,
			MasteryID:    2,
			APCost:       2,
			MPCost:       25,
			Effects: []*domain.SpellEffect{
				{EffectID: 303, DurationInTurns: 1},
			},
		},
		{
			ID:           18,
			Name:         "EnergySap",
			DisplayNames: datatypes.JSONMap{"en": "Energy Sap", "th": "ทรายดูดพลัง"},
			Descriptions: datatypes.JSONMap{"en": "Creates a patch of sapping sand that drains the target's energy resources.", "th": "สร้างพื้นทรายดูดพลังที่จะสูบพลังงานของเป้าหมาย"},
			ElementID:    5,
			MasteryID:    4,
			APCost:       3,
			MPCost:       30,
			Effects: []*domain.SpellEffect{
				{EffectID: 5, BaseValue: 25},
			},
		},
		{
			ID:           19,
			Name:         "WindWall",
			DisplayNames: datatypes.JSONMap{"en": "Wind Wall", "th": "กำแพงลม"},
			Descriptions: datatypes.JSONMap{"en": "Conjures a swirling wall of wind to block incoming attacks.", "th": "ร่ายกำแพงลมหมุนวนเพื่อป้องกันการโจมตีที่เข้ามา"},
			ElementID:    6,
			MasteryID:    2,
			APCost:       2,
			MPCost:       25,
			Effects: []*domain.SpellEffect{
				{EffectID: 2, BaseValue: 300},
			},
		},
		{
			ID:           20,
			Name:         "BreakTheChains",
			DisplayNames: datatypes.JSONMap{"en": "Break The Chains", "th": "ปลดปล่อยพันธนาการ"},
			Descriptions: datatypes.JSONMap{"en": "A powerful gust of wind shatters all magical and physical restraints.", "th": "ลมกระโชกแรงทำลายพันธนาการทั้งทางกายภาพและเวทมนตร์ทั้งหมด"},
			ElementID:    6,
			MasteryID:    4,
			APCost:       2,
			MPCost:       20,
			Effects: []*domain.SpellEffect{
				{EffectID: 7, BaseValue: 99},
			},
		},
		{
			ID:           21,
			Name:         "MoltenMeteor",
			DisplayNames: datatypes.JSONMap{"en": "Molten Meteor", "th": "อุกกาบาต"},
			Descriptions: datatypes.JSONMap{"en": "Calls down a meteor of molten rock, dealing impact damage and leaving the ground burning.", "th": "เรียกอุกกาบาตหินหลอมเหลวลงมา สร้างความเสียหายเมื่อกระทบและทิ้งพื้นที่ที่ลุกไหม้ไว้"},
			ElementID:    7,
			MasteryID:    1,
			APCost:       2,
			MPCost:       30,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 75},
				{EffectID: 306, BaseValue: 20, DurationInTurns: 2},
			},
		},
		{
			ID:           22,
			Name:         "LavaShield",
			DisplayNames: datatypes.JSONMap{"en": "Lava Shield", "th": "โล่เพลิงลาวา"},
			Descriptions: datatypes.JSONMap{"en": "Forms a shield of superheated lava that protects the caster and burns attackers.", "th": "สร้างโล่จากลาวาที่ร้อนระอุเพื่อป้องกันผู้ร่ายและเผาไหม้ผู้ที่โจมตีเข้ามา"},
			ElementID:    7,
			MasteryID:    2,
			APCost:       2,
			MPCost:       25,
			Effects: []*domain.SpellEffect{
				{EffectID: 2, BaseValue: 350},
			},
		},
		{
			ID:           23,
			Name:         "ChainLightning",
			DisplayNames: datatypes.JSONMap{"en": "Chain Lightning", "th": "สายฟ้าฟาด"},
			Descriptions: datatypes.JSONMap{"en": "Unleashes a bolt of lightning that arcs between multiple enemies.", "th": "ปล่อยสายฟ้าที่สามารถชิ่งไปมาระหว่างศัตรูหลายตัวได้"},
			ElementID:    8,
			MasteryID:    1,
			APCost:       2,
			MPCost:       30,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 65},
			},
		},
		{
			ID:           24,
			Name:         "ElectricCage",
			DisplayNames: datatypes.JSONMap{"en": "Electric Cage", "th": "พันธนาการไฟฟ้า"},
			Descriptions: datatypes.JSONMap{"en": "Surrounds the target with a cage of crackling electricity, paralyzing them.", "th": "ล้อมรอบเป้าหมายด้วยกรงไฟฟ้าสถิต ทำให้เป็นอัมพาต"},
			ElementID:    8,
			MasteryID:    2,
			APCost:       3,
			MPCost:       35,
			Effects: []*domain.SpellEffect{
				{EffectID: 305, DurationInTurns: 1},
			},
		},
		{
			ID:           25,
			Name:         "AcidSpray",
			DisplayNames: datatypes.JSONMap{"en": "Acid Spray", "th": "สาดกรด"},
			Descriptions: datatypes.JSONMap{"en": "Sprays a corrosive acid that deals damage over time and melts away armor.", "th": "สาดกรดกัดกร่อนที่สร้างความเสียหายต่อเนื่องและสลายเกราะป้องกัน"},
			ElementID:    9,
			MasteryID:    1,
			APCost:       2,
			MPCost:       30,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 50},
				{EffectID: 308, BaseValue: 15, DurationInTurns: 3},
			},
		},
		{
			ID:           26,
			Name:         "VeilbreakerEdge",
			DisplayNames: datatypes.JSONMap{"en": "Veilbreaker Edge", "th": "คมดาบทลายม่าน"},
			Descriptions: datatypes.JSONMap{"en": "Empowers a weapon to slice through magical barriers and defenses.", "th": "เสริมพลังให้กับอาวุธเพื่อให้สามารถฟันฝ่าม่านพลังและเกราะป้องกันเวทมนตร์ได้"},
			ElementID:    9,
			MasteryID:    3,
			APCost:       2,
			MPCost:       20,
			Effects: []*domain.SpellEffect{
				{EffectID: 108, BaseValue: 75, DurationInTurns: 1},
			},
		},
		{
			ID:           27,
			Name:         "UnstableBlast",
			DisplayNames: datatypes.JSONMap{"en": "Unstable Blast", "th": "ระเบิดพลัง"},
			Descriptions: datatypes.JSONMap{"en": "Releases a volatile and unpredictable blast of raw energy, dealing massive damage.", "th": "ปลดปล่อยการระเบิดของพลังงานดิบที่ไม่เสถียรและคาดเดายาก สร้างความเสียหายมหาศาล"},
			ElementID:    10,
			MasteryID:    1,
			APCost:       3,
			MPCost:       35,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 140},
			},
		},
		{
			ID:           28,
			Name:         "ResetTime",
			DisplayNames: datatypes.JSONMap{"en": "Reset Time", "th": "รีเซ็ตเวลา"},
			Descriptions: datatypes.JSONMap{"en": "Manipulates time to give the caster another turn to act immediately.", "th": "ควบคุมเวลาเพื่อให้ผู้ร่ายสามารถลงมือได้อีกครั้งในทันที"},
			ElementID:    10,
			MasteryID:    4,
			APCost:       3,
			MPCost:       30,
			Effects: []*domain.SpellEffect{
				{EffectID: 6, BaseValue: 2},
			},
		},
		{
			ID:           29,
			Name:         "ShieldShatter",
			DisplayNames: datatypes.JSONMap{"en": "Shield Shatter", "th": "ทลายเกราะ"},
			Descriptions: datatypes.JSONMap{"en": "A powerful concussive force that instantly shatters any active shields on the target.", "th": "ปล่อยคลื่นกระแทกอันทรงพลังที่จะทำลายโล่ป้องกันของเป้าหมายในทันที"},
			ElementID:    11,
			MasteryID:    1,
			APCost:       3,
			MPCost:       40,
			Effects: []*domain.SpellEffect{
				{EffectID: 1, BaseValue: 0},
			},
		},
		{
			ID:           30,
			Name:         "Aegis",
			DisplayNames: datatypes.JSONMap{"en": "Aegis", "th": "เกราะเทวะ"},
			Descriptions: datatypes.JSONMap{"en": "Summons a divine shield that provides immense protection and temporary immunity.", "th": "อัญเชิญโล่ศักดิ์สิทธิ์ที่มอบการป้องกันอันมหาศาลและต้านทานสถานะผิดปกติชั่วคราว"},
			ElementID:    11,
			MasteryID:    2,
			APCost:       3,
			MPCost:       40,
			Effects: []*domain.SpellEffect{
				{EffectID: 2, BaseValue: 500},
				{EffectID: 106},
			},
		},
		{
			ID:           31,
			Name:         "Lifesteal",
			DisplayNames: datatypes.JSONMap{"en": "Lifesteal", "th": "สูบชีวิต"},
			Descriptions: datatypes.JSONMap{"en": "Drains the life force of an enemy to heal the caster.", "th": "ดูดพลังชีวิตของศัตรูเพื่อฟื้นฟูพลังชีวิตของผู้ร่าย"},
			ElementID:    12,
			MasteryID:    1,
			APCost:       2,
			MPCost:       40,
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 60}},
		},
		{
			ID:           32,
			Name:         "Revive",
			DisplayNames: datatypes.JSONMap{"en": "Revive", "th": "ชุบชีวิต"},
			Descriptions: datatypes.JSONMap{"en": "Breathes life back into a fallen ally, returning them to the fight.", "th": "เป่าลมหายใจแห่งชีวิตให้กับสหายที่ล้มลง เพื่อให้พวกเขากลับสู่สนามรบอีกครั้ง"},
			ElementID:    12,
			MasteryID:    4,
			APCost:       3,
			MPCost:       50,
			Effects:      []*domain.SpellEffect{{EffectID: 3, BaseValue: 40}},
		},
		{
			ID:           33,
			Name:         "VacuumBlade",
			DisplayNames: datatypes.JSONMap{"en": "Vacuum Blade", "th": "ดาบสูญญากาศ"},
			Descriptions: datatypes.JSONMap{"en": "Creates a blade of vacuum that slices through enemies and pulls them closer.", "th": "สร้างดาบแห่งสุญญากาศที่ฟาดฟันศัตรูและสามารถดึงพวกเขาเข้ามาใกล้ได้"},
			ElementID:    13,
			MasteryID:    1,
			APCost:       2,
			MPCost:       40,
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 60}},
		},
		{
			ID:           34,
			Name:         "PhaseShift",
			DisplayNames: datatypes.JSONMap{"en": "Phase Shift", "th": "ม่านมิติ"},
			Descriptions: datatypes.JSONMap{"en": "The caster briefly shifts into another dimension, becoming immune to all damage.", "th": "ผู้ร่ายจะย้ายตัวเองไปยังอีกมิติหนึ่งชั่วขณะ ทำให้เป็นอมตะต่อความเสียหายทั้งหมด"},
			ElementID:    13,
			MasteryID:    2,
			APCost:       3,
			MPCost:       40,
			Effects:      []*domain.SpellEffect{},
		},
		{
			ID:           35,
			Name:         "Supernova",
			DisplayNames: datatypes.JSONMap{"en": "Supernova", "th": "ซูเปอร์โนวา"},
			Descriptions: datatypes.JSONMap{"en": "Detonates with the force of a dying star, inflicting immense damage to all enemies.", "th": "ระเบิดออกด้วยพลังของดาวฤกษ์ที่กำลังจะดับสูญ สร้างความเสียหายมหาศาลแก่ศัตรูทั้งหมด"},
			ElementID:    14,
			MasteryID:    1,
			APCost:       3,
			MPCost:       60,
			Effects:      []*domain.SpellEffect{{EffectID: 1, BaseValue: 180}},
		},
		{
			ID:           36,
			Name:         "SolarCorona",
			DisplayNames: datatypes.JSONMap{"en": "Solar Corona", "th": "รัศมีสุริยะ"},
			Descriptions: datatypes.JSONMap{"en": "Emits an aura of intense solar energy that strengthens allies and weakens enemies.", "th": "ปล่อยออร่าพลังงานสุริยะอันเข้มข้นที่จะเสริมความแข็งแกร่งให้พันธมิตรและทำให้ศัตรูอ่อนแอลง"},
			ElementID:    14,
			MasteryID:    2,
			APCost:       2,
			MPCost:       40,
			Effects:      []*domain.SpellEffect{},
		},
		{
			ID:           37,
			Name:         "Entropy",
			DisplayNames: datatypes.JSONMap{"en": "Entropy", "th": "เอนโทรปี"},
			Descriptions: datatypes.JSONMap{"en": "Inflicts a chaotic state upon the target, causing them to decay over time.", "th": "ทำให้เป้าหมายตกอยู่ในสภาวะโกลาหล ส่งผลให้เสื่อมสลายไปตามกาลเวลา"},
			ElementID:    15,
			MasteryID:    1,
			APCost:       3,
			MPCost:       50,
			Effects:      []*domain.SpellEffect{},
		},
		{
			ID:           38,
			Name:         "InversionField",
			DisplayNames: datatypes.JSONMap{"en": "Inversion Field", "th": "ม่านมิติผกผัน"},
			Descriptions: datatypes.JSONMap{"en": "Creates a field that inverts certain magical effects, turning healing into damage.", "th": "สร้างพื้นที่ที่กลับผลของเวทมนตร์บางอย่าง เปลี่ยนการรักษาเป็นการสร้างความเสียหาย"},
			ElementID:    15,
			MasteryID:    2,
			APCost:       2,
			MPCost:       40,
			Effects:      []*domain.SpellEffect{},
		},
	}
	// For complex associations, it's safer to delete and recreate
	tx.Exec("DELETE FROM spell_effects")
	tx.Exec("DELETE FROM spells")
	if err := tx.Create(&spells).Error; err != nil {
		return err
	}
	return nil
}

func seedGameConfigs(tx *gorm.DB) error {

	configs := []domain.GameConfig{
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
		// Combat System & Other Rules (from previous agreements)
		// ========================================================================
		{Key: "GAME_VERSION", Value: "1.0.0", Description: "The current version of the game data."},
		{Key: "COMBAT_AP_PER_TURN", Value: "3", Description: "The amount of Action Points a player gains at the start of their turn."},
		{Key: "COMBAT_MAX_AP", Value: "6", Description: "The maximum amount of Action Points a player can accumulate."},
	}

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
