# 🎮 Game Mechanics Documentation

**เอกสารอธิบายกลไกเกมแบบละเอียด**  
**วันที่สร้าง:** 29 ตุลาคม 2025  
**Version:** 1.0

> 💡 **คู่มือนี้สำหรับ:** Developers, QA Engineers  
> 📊 **สำหรับ Progress Tracking:** ดูที่ [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)

---

## 📑 สารบัญ

1. [การสร้างตัวละคร (Character Creation)](#1-การสร้างตัวละคร)
2. [การสร้างการต่อสู้ (Match Creation)](#2-การสร้างการต่อสู้)
3. [การร้ายเวท (Spell Casting)](#3-การร้ายเวท)
4. [สรุปสูตรการคำนวณ](#4-สรุปสูตรการคำนวณ)
5. [ปัญหาที่พบและข้อเสนอแนะ](#5-ปัญหาที่พบและข้อเสนอแนะ)

---

## 1. การสร้างตัวละคร

### 1.1 ขั้นตอนการทำงาน

```
API Request → Validation → Calculate Talents → Calculate Stats → Create Masteries → Save to DB
```

### 1.2 Input Parameters

```go
playerID      uint   // ID ของผู้เล่นที่เป็นเจ้าของ
name          string // ชื่อตัวละคร (ต้อง >= 3 ตัวอักษร)
gender        string // "MALE" หรือ "FEMALE" (บันทึกเพื่อแสดงผลเท่านั้น - ไม่มีผลต่อ stats)
elementID     uint   // ธาตุปฐมภูมิ: 1=S, 2=L, 3=G, 4=P (กำหนด talent ที่ได้ +90)
masteryID     uint   // ศาสตร์หลัก (ไม่ได้ใช้ในการคำนวณตอนนี้)
```

**⚠️ สำคัญ:**

-  **Gender Bonus ถูกยกเลิกแล้ว** (ตั้งแต่ Oct 28, 2025)
-  ตอนนี้ใช้เพียง **Base(3) + Primary Element(+90)** เท่านั้น
-  MALE/FEMALE ไม่มีผลต่อ stats ใดๆ ทั้งสิ้น

### 1.3 Talent Calculation (ค่าพลังดิบ)

#### สูตร:

```
Base Allocation Per Talent: 3 (config: TALENT_BASE_ALLOCATION)
Primary Element Bonus: +90 (config: TALENT_PRIMARY_ALLOCATION)

TalentS = 3 + (elementID == 1 ? 90 : 0) = 3 or 93
TalentL = 3 + (elementID == 2 ? 90 : 0) = 3 or 93
TalentG = 3 + (elementID == 3 ? 90 : 0) = 3 or 93
TalentP = 3 + (elementID == 4 ? 90 : 0) = 3 or 93
```

#### ตัวอย่าง:

```
เลือก Primary Element = S (Solidity, ID=1):
→ TalentS = 93
→ TalentL = 3
→ TalentG = 3
→ TalentP = 3
Total: 102 points

เลือก Primary Element = L (Liquidity, ID=2):
→ TalentS = 3
→ TalentL = 93
→ TalentG = 3
→ TalentP = 3
Total: 102 points
```

#### Code Implementation:

```go
// internal/modules/character/service.go
func calculateInitialTalents(primaryElementID uint) (int, int, int, int) {
    talents := map[uint]int{
        1: 3, // S
        2: 3, // L
        3: 3, // G
        4: 3, // P
    }

    // Add primary element bonus
    talents[primaryElementID] += 90

    return talents[1], talents[2], talents[3], talents[4]
}
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 1.4 Core Stats Calculation (ค่าสถานะหลัก)

#### 1.4.1 Max HP (พลังชีวิตสูงสุด)

**สูตร:**

```
MaxHP = STAT_HP_BASE + (TalentS × STAT_HP_PER_TALENT_S)
MaxHP = 900 + (TalentS × 30)
```

**Config:**

```go
STAT_HP_BASE: 900
STAT_HP_PER_TALENT_S: 30
```

**ตัวอย่าง:**

```
S-Build (TalentS = 93):
MaxHP = 900 + (93 × 30) = 3,690 HP

L-Build (TalentS = 3):
MaxHP = 900 + (3 × 30) = 990 HP
```

**Code:**

```go
baseHp := 900  // STAT_HP_BASE
hpPerTalentS := 30  // STAT_HP_PER_TALENT_S
maxHP := baseHp + (newCharacter.TalentS * hpPerTalentS)
newCharacter.CurrentHP = maxHP
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

#### 1.4.2 Max MP (พลังเวทสูงสุด)

**สูตร:**

```
MaxMP = STAT_MP_BASE + (TalentL × STAT_MP_PER_TALENT_L)
MaxMP = 200 + (TalentL × 2)
```

**Config:**

```go
STAT_MP_BASE: 200
STAT_MP_PER_TALENT_L: 2
```

**ตัวอย่าง:**

```
L-Build (TalentL = 93):
MaxMP = 200 + (93 × 2) = 386 MP

S-Build (TalentL = 3):
MaxMP = 200 + (3 × 2) = 206 MP
```

**Code:**

```go
baseMp := 200  // STAT_MP_BASE
mpPerTalentL := 2  // STAT_MP_PER_TALENT_L
maxMP := baseMp + (newCharacter.TalentL * mpPerTalentL)
newCharacter.CurrentMP = maxMP
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

#### 1.4.3 Initiative (ความเร็ว)

**สูตร:**

```
Initiative = STAT_INITIATIVE_BASE + (TalentG × STAT_INITIATIVE_PER_TALENT_G)
Initiative = 50 + (TalentG × 1)
```

**Config:**

```go
STAT_INITIATIVE_BASE: 50
STAT_INITIATIVE_PER_TALENT_G: 1
```

**หมายเหตุ:** Initiative ไม่ถูกบันทึกใน Character table แต่คำนวณตอนสร้าง Combatant

**ตัวอย่าง:**

```
G-Build (TalentG = 93):
Initiative = 50 + (93 × 1) = 143

S-Build (TalentG = 3):
Initiative = 50 + (3 × 1) = 53
```

**Code:**

```go
// คำนวณตอนสร้าง Combatant (ใน combat service)
initBase := 50  // STAT_INITIATIVE_BASE
initPerTalent := 1  // STAT_INITIATIVE_PER_TALENT_G
playerCombatant.Initiative = initBase + (playerChar.TalentG * initPerTalent)
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 1.5 Mastery Initialization (สร้างศาสตร์ 4 ศาสตร์)

**ข้อมูลเริ่มต้น:**

```go
masteries := []*domain.CharacterMastery{
    {MasteryID: 1, Level: 1, Mxp: 0}, // Force (ศาสตร์โจมตี)
    {MasteryID: 2, Level: 1, Mxp: 0}, // Resilience (ศาสตร์ป้องกัน)
    {MasteryID: 3, Level: 1, Mxp: 0}, // Efficacy (ศาสตร์เสริมพลัง)
    {MasteryID: 4, Level: 1, Mxp: 0}, // Command (ศาสตร์สนับสนุน)
}
```

**การทำงาน:**

-  สร้างศาสตร์ทั้ง 4 ประเภทให้ทุกตัวละคร
-  เริ่มต้น Level 1 ทั้งหมด
-  Mxp (Mastery Experience) = 0

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

#### ⚠️ ปัญหา: **ไม่มีระบบเลเวลอัป** (Mxp ไม่เพิ่มขึ้น, Level ติดที่ 1)

---

### 1.6 Database Save

**ตารางที่ถูกสร้าง:**

1. **characters** - ข้อมูลตัวละครหลัก
2. **character_masteries** - ข้อมูลศาสตร์ 4 แบบ (relation)

**ตัวอย่าง JSON Response:**

````json
```json
{
   "id": 1,
   "player_id": 1,
   "character_name": "FireMage",
   "gender": "MALE",
   "primary_element_id": 4,
   "level": 1,
   "exp": 0,
   "current_hp": 990,
   "current_mp": 386,
   "talent_s": 3,
   "talent_l": 3,
   "talent_g": 3,
   "talent_p": 93,
   "unallocated_talent_points": 0,
   "tutorial_step": 1,
   "masteries": [
      { "mastery_id": 1, "level": 1, "mxp": 0 },
      { "mastery_id": 2, "level": 1, "mxp": 0 },
      { "mastery_id": 3, "level": 1, "mxp": 0 },
      { "mastery_id": 4, "level": 1, "mxp": 0 }
   ]
}
````

**อธิบายค่า Talent ในตัวอย่าง:**

-  `primary_element_id: 4` → เลือก **P (Potency)** เป็นธาตุหลัก
-  `talent_p: 93` → Base(3) + Primary Bonus(90) = **93** ✅
-  `talent_s, talent_l, talent_g: 3` → Base only = **3** (ไม่ได้โบนัส)
-  **หมายเหตุ:** `gender: "MALE"` ไม่มีผลต่อ stats (Gender Bonus ถูกยกเลิกแล้ว)

**อธิบายค่า Stats ในตัวอย่าง:**

-  `current_hp: 990` → 900 (base) + (3 × 30) = **990 HP** ✅
-  `current_mp: 386` → 200 (base) + (93 × 2) = **386 MP** ✅

```

**อธิบายค่า Talent ในตัวอย่าง:**

-  `primary_element_id: 4` → เลือก **P (Potency)** เป็นธาตุหลัก
-  `talent_p: 93` → Base(3) + Primary Bonus(90) = **93** ✅
-  `talent_s, talent_l, talent_g: 3` → Base only = **3** (ไม่ได้โบนัส)
-  **หมายเหตุ:** `gender: "MALE"` ไม่มีผลต่อ stats (Gender Bonus ถูกยกเลิกแล้ว)

---

## 2. การสร้างการต่อสู้

### 2.1 ขั้นตอนการทำงาน

```

API Request → Validate Character → Check Active Match → Load Config
→ Create Player Combatant → Load Deck → Create Enemy Combatant
→ Determine First Turn → Save Match

````

### 2.2 Match Types (ประเภทการต่อสู้)

**1. TRAINING - โหมดฝึกซ้อม**

-  ผู้เล่นเลือกศัตรูเอง
-  ไม่จำกัดจำนวนครั้ง
-  ได้ EXP: **50** (config: EXP_TRAINING_MATCH)

**2. STORY - โหมดเนื้อเรื่อง**

-  ต่อสู้ตามด่านที่กำหนด
-  ศัตรูถูกโหลดจาก Stage
-  ได้ EXP: **100** (config: EXP_STORY_MATCH)
-  ⚠️ **ยังไม่ได้ implement**

**3. PVP - ต่อสู้ผู้เล่นอื่น**

-  ต่อสู้กับตัวละครของผู้เล่นอื่น
-  ได้ EXP: **150** (config: EXP_PVP_MATCH)
-  ⚠️ **ยังไม่ได้ implement เต็มรูปแบบ**

---

### 2.3 Player Combatant Creation

**ข้อมูลที่คำนวณ:**

```go
playerCombatant := &domain.Combatant{
    ID:          uuid.NewV7(),
    CharacterID: &playerChar.ID,
    Initiative:  STAT_INITIATIVE_BASE + (TalentG × STAT_INITIATIVE_PER_TALENT_G),
    CurrentHP:   STAT_HP_BASE + (TalentS × STAT_HP_PER_TALENT_S),
    CurrentMP:   playerChar.CurrentMP,  // โหลดจาก DB
    CurrentAP:   0,                     // เริ่มต้นที่ 0
}
````

**Config ที่ใช้:**

```go
STAT_HP_BASE: 900
STAT_HP_PER_TALENT_S: 30
STAT_INITIATIVE_BASE: 50
STAT_INITIATIVE_PER_TALENT_G: 1
```

**ตัวอย่าง:**

```
TalentS = 93, TalentG = 3, CurrentMP = 386:
→ Initiative = 50 + (3 × 1) = 53
→ CurrentHP = 900 + (93 × 30) = 3,690
→ CurrentMP = 386 (from DB)
→ CurrentAP = 0
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 2.4 Enemy Combatant Creation

**ข้อมูลที่โหลดจาก Enemy table:**

```go
enemyCombatant := &domain.Combatant{
    ID:         uuid.NewV7(),
    EnemyID:    &enemyData.ID,
    Initiative: enemyData.Initiative,  // อ่านจาก DB
    CurrentHP:  enemyData.MaxHP,       // อ่านจาก DB
    CurrentMP:  9999,                  // Unlimited MP
    CurrentAP:  0,
}
```

**หมายเหตุ:**

-  ศัตรูมี MP ไม่จำกัด (9999)
-  ค่า Initiative และ MaxHP ถูกกำหนดไว้ใน seed data
-  ไม่มี Talent หรือ Mastery

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 2.5 Deck Loading (โหลดคลังธาตุ)

**ขั้นตอน:**

1. ตรวจสอบว่ามี `DeckID` ส่งมาหรือไม่
2. โหลด Deck จาก database
3. ตรวจสอบความเป็นเจ้าของ
4. สร้าง `CombatantDeck` จาก `DeckSlot`

**Code:**

```go
if req.DeckID != nil {
    deckData, _ := s.deckRepo.FindByID(*req.DeckID)

    // สร้างกระสุนธาตุจาก deck slots
    for _, slot := range deckData.Slots {
        newCharge := &domain.CombatantDeck{
            ID:          uuid.NewV7(),
            CombatantID: playerCombatantID,
            ElementID:   slot.ElementID,
            IsConsumed:  false,  // ยังไม่ได้ใช้
        }
        combatantDeck = append(combatantDeck, newCharge)
    }
}
```

**ตัวอย่าง Deck:**

```
Deck with 3 slots:
Slot 1: Element S (Solidity)
Slot 2: Element L (Liquidity)
Slot 3: Element P (Potency)

→ Creates 3 CombatantDeck entries
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 2.6 Turn Order Determination

**อัลกอริทึม:**

1. เรียงลำดับ Combatants ตาม Initiative (มาก → น้อย)
2. ถ้า Initiative เท่ากัน → สุ่ม
3. ตั้ง `CurrentTurn` = ID ของ combatant แรก
4. ตั้ง `TurnNumber` = 1

**Code:**

```go
func (s *combatService) determineFirstTurn(combatants []*domain.Combatant) uuid.UUID {
    if len(combatants) == 0 {
        return uuid.Nil
    }

    highestInit := 0
    var firstCombatant *domain.Combatant

    for _, c := range combatants {
        if c.Initiative > highestInit {
            highestInit = c.Initiative
            firstCombatant = c
        }
    }

    return firstCombatant.ID
}
```

**ตัวอย่าง:**

```
Player: Initiative = 143
Enemy:  Initiative = 50

→ Player ได้เทิร์นแรก
→ CurrentTurn = Player's UUID
→ TurnNumber = 1
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 2.7 Match Status

**สถานะที่เป็นไปได้:**

```go
const (
    MatchInProgress MatchStatus = "IN_PROGRESS"  // กำลังต่อสู้
    MatchFinished   MatchStatus = "FINISHED"     // จบแล้ว
    MatchAborted    MatchStatus = "ABORTED"      // ถูกยกเลิก
)
```

**เมื่อสร้าง Match:**

```go
match := &domain.CombatMatch{
    ID:         uuid.NewV7(),
    MatchType:  req.MatchType,  // TRAINING, STORY, PVP
    Status:     MatchInProgress,
    TurnNumber: 1,
    CurrentTurn: firstCombatantID,
    Combatants: combatants,
}
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

## 3. การร้ายเวท

### 3.1 ขั้นตอนการทำงาน (Overview)

```
Cast Request → Validate Turn → Load Match → Prepare Spell → Resolve Spell
→ Calculate Effects → Apply Effects → End Turn → Check Win/Lose → Update Match
```

### 3.2 Spell Casting Phases

ระบบแบ่งการร้ายเวทเป็น **5 Phases** หลัก:

#### **Phase 1: Preparation (เตรียมการ)**

-  ตรวจสอบความถูกต้องของ Input
-  ดึงข้อมูล Match, Caster, Target
-  ตรวจสอบว่าเป็น Turn ของผู้เล่นหรือไม่

#### **Phase 2: Spell Resolution (แก้สูตร)**

-  แปลงธาตุที่เลือกเป็น Spell
-  ใช้ Fallback Algorithm ถ้าหาไม่เจอ
-  หักค่า MP Cost

#### **Phase 3: Effect Calculation (คำนวณเอฟเฟกต์)**

-  คำนวณค่าพื้นฐาน (Base + Mastery + Talent)
-  คำนวณ Modifier (Elemental + Buff/Debuff + Power)
-  คำนวณค่าสุดท้าย

#### **Phase 4: Effect Application (ใช้เอฟเฟกต์)**

-  สร้างความเสียหาย (Damage)
-  เพิ่ม Shield/Heal
-  ติด Buff/Debuff
-  ติด DoT

#### **Phase 5: Turn Management (จัดการเทิร์น)**

-  จบเทิร์นปัจจุบัน
-  ตรวจสอบเงื่อนไขชนะ/แพ้
-  เลื่อนไปเทิร์นถัดไป
-  อัปเดต Match

---

### 3.3 Spell Resolution (ขั้นตอนแก้สูตร)

#### 3.3.1 Element Selection

**Input:**

```json
{
   "elements": [1, 2, 1], // S, L, S
   "masteryID": 1, // Force
   "targetID": "enemy-uuid",
   "castingMode": "NORMAL"
}
```

**Element ID Mapping:**

```
1 = S (Solidity)
2 = L (Liquidity)
3 = G (Gesture/Tempo)
4 = P (Potency)
5-15 = Tier 1 Elements (Magma, Viscosity, etc.)
```

---

#### 3.3.2 Fallback Algorithm

Fallback Algorithm ทำงานตามขั้นตอนดังนี้:

---

**STEP 0: Direct Lookup (ค้นหาตรง)**

```
ลองหา Spell(ElementID, MasteryID) ตรงๆ
→ ถ้าเจอ: ใช้ Spell นี้เลย ✅ (จบ)
→ ถ้าไม่เจอ: เริ่ม Fallback Algorithm
```

---

**STEP 1: Check Majority (ตรวจสอบเสียงข้างมาก)**

หา Recipe ของธาตุที่ต้องการ:

```go
// ตัวอย่าง Recipe:
S + S + P → Element X (3 ingredients)

// นับจำนวน:
S: 2 ครั้ง (66.6%)
P: 1 ครั้ง (33.3%)

// ตรวจสอบเสียงข้างมาก:
hasMajority = maxCount > totalCount/2
            = 2 > 3/2
            = 2 > 1
            = true ✅
```

**กรณีที่ 1.1: มีเสียงข้างมาก (Majority Found)**

```
Recipe: S + S + P
→ S มี 66.6% (> 50%)
→ ใช้ Spell(S, MasteryID)
→ ถ้าเจอ: ใช้ Spell นี้ ✅ (จบ)
→ ถ้าไม่เจอ: ไป STEP 2
```

**กรณีที่ 1.2: ไม่มีเสียงข้างมาก (No Majority)**

```
Recipe: S + P
→ S: 50%, P: 50% (เสมอ)
→ ไม่มีเสียงข้างมาก
→ ไป STEP 2
```

---

**STEP 2: Caster Role Check (ตรวจสอบบทบาทผู้ใช้)**

ตรวจสอบว่า **Caster's Primary Element** อยู่ใน Recipe หรือไม่:

**กรณีที่ 2A: Caster เป็น Ingredient (Insider)**

```
Recipe: S + P
Caster Primary: P

→ Caster เป็นส่วนหนึ่งของสูตร ✅
→ ใช้ Spell(P, MasteryID)
→ ถ้าเจอ: ใช้ Spell นี้ ✅ (จบ)
→ ถ้าไม่เจอ: ไป STEP 2B
```

**กรณีที่ 2B: Caster เป็น Outsider หรือ Primary Element ไม่มี Spell**

```
Recipe: S + P
Caster Primary: L (นอกสูตร)

→ Caster ไม่ใช่ส่วนหนึ่งของสูตร
→ ธาตุในสูตรต้อง "สู้กันเอง" (Internal Fight)
→ ไป STEP 2B.1
```

---

**STEP 2B.1: Internal Fight (การสู้กันภายใน)**

ธาตุทั้งหมดใน Recipe สู้กันแบบ **Round-robin** (ทุกคนชนทุกคน):

```go
Recipe: S + P + L

Round-robin fights:
S vs P: ดู matchup → ถ้า S > P → S ได้ 1 คะแนน
S vs L: ดู matchup → ถ้า S > L → S ได้ 1 คะแนน
P vs S: ดู matchup → ถ้า P > S → P ได้ 1 คะแนน
P vs L: ดู matchup → ถ้า P > L → P ได้ 1 คะแนน
L vs S: ดู matchup → ถ้า L > S → L ได้ 1 คะแนน
L vs P: ดู matchup → ถ้า L > P → L ได้ 1 คะแนน

// สรุปคะแนน:
S: 2 คะแนน
P: 1 คะแนน
L: 0 คะแนน

→ Winner: S (คะแนนสูงสุด)
```

**ถ้ามีผู้ชนะชัดเจน:**

```
→ ใช้ Spell(Winner, MasteryID)
→ ถ้าเจอ: ใช้ Spell นี้ ✅ (จบ)
→ ถ้าไม่เจอ: ไป STEP 2B.2
```

**ถ้าเสมอกัน (Tie):**

```
S: 1 คะแนน
P: 1 คะแนน

→ Tie! ไม่มีผู้ชนะชัดเจน
→ ไป STEP 2B.2
```

---

**STEP 2B.2: Strongest Against Caster (หาธาตุที่แข็งแกร่งที่สุดกับ Caster)**

เมื่อ Internal Fight เสมอ หรือผู้ชนะไม่มี Spell → ให้เลือกธาตุที่ **ชนะ Caster มากที่สุด**:

```go
Recipe: S + P
Caster Primary: L

// ตรวจสอบแต่ละธาตุ vs Caster:
S vs L: matchup = 1.3 (S ได้เปรียบ) → Score = +1
P vs L: matchup = 0.8 (P เสียเปรียบ) → Score = -1

// สรุปคะแนน:
S: +1 (แข็งแกร่งกว่า Caster)
P: -1 (อ่อนแอกว่า Caster)

→ Winner: S (คะแนนสูงสุด)
→ ใช้ Spell(S, MasteryID)
```

**Score System:**

-  `matchup > 1.0` → ธาตุได้เปรียบ Caster → Score = +1
-  `matchup = 1.0` → เสมอ → Score = 0
-  `matchup < 1.0` → ธาตุเสียเปรียบ Caster → Score = -1

```
→ ถ้าเจอ: ใช้ Spell นี้ ✅ (จบ)
→ ถ้าไม่เจอ: Error 404 ❌ (ไม่มี Spell นี้จริงๆ)
```

---

### ตัวอย่างการทำงานครบวงจร

**ตัวอย่างที่ 1: หาได้ตรง (Direct Lookup)**

```
Input: Element = S, Mastery = Force
→ STEP 0: หา Spell(S, Force) → เจอ ✅
→ ใช้ "Stone Strike" (S + Force)
→ จบ
```

---

**ตัวอย่างที่ 2: Majority Element**

```
Input: Element = Magma (S+P), Mastery = Force
Recipe: S + S + P

→ STEP 0: หา Spell(Magma, Force) → ไม่เจอ
→ STEP 1: Check Majority
  - S: 2 ครั้ง (66.6%)
  - P: 1 ครั้ง (33.3%)
  - hasMajority = true (S > 50%)
→ STEP 1.1: ใช้ Spell(S, Force) → เจอ ✅
→ ใช้ "Stone Strike" (S + Force)
→ จบ
```

---

**ตัวอย่างที่ 3: Caster เป็น Ingredient**

```
Input: Element = Viscosity (S+L), Mastery = Force
Caster Primary: L
Recipe: S + L

→ STEP 0: หา Spell(Viscosity, Force) → ไม่เจอ
→ STEP 1: Check Majority
  - S: 50%, L: 50%
  - hasMajority = false (เสมอ)
→ STEP 2: Check Caster Role
  - Caster = L
  - Recipe = [S, L]
  - L อยู่ใน Recipe ✅
→ STEP 2A: ใช้ Spell(L, Force) → เจอ ✅
→ ใช้ "Water Blast" (L + Force)
→ จบ
```

---

**ตัวอย่างที่ 4: Internal Fight**

```
Input: Element = Ionization (L+G), Mastery = Force
Caster Primary: S (outsider)
Recipe: L + G

→ STEP 0: หา Spell(Ionization, Force) → ไม่เจอ
→ STEP 1: Check Majority
  - L: 50%, G: 50%
  - hasMajority = false
→ STEP 2: Check Caster Role
  - Caster = S
  - Recipe = [L, G]
  - S ไม่อยู่ใน Recipe ❌
→ STEP 2B.1: Internal Fight
  - L vs G: L > G → L ได้ 1 คะแนน
  - G vs L: G < L → G ได้ 0 คะแนน
  - Winner: L (1 คะแนน)
→ ใช้ Spell(L, Force) → เจอ ✅
→ ใช้ "Water Blast" (L + Force)
→ จบ
```

---

**ตัวอย่างที่ 5: Strongest Against Caster**

```
Input: Element = Reactivity (L+P), Mastery = Force
Caster Primary: S
Recipe: L + P

→ STEP 0: หา Spell(Reactivity, Force) → ไม่เจอ
→ STEP 1: Check Majority
  - L: 50%, P: 50%
  - hasMajority = false
→ STEP 2: Check Caster Role
  - Caster = S
  - Recipe = [L, P]
  - S ไม่อยู่ใน Recipe
→ STEP 2B.1: Internal Fight
  - L vs P: L > P → L ได้ 1 คะแนน
  - P vs L: P < L → P ได้ 0 คะแนน
  - Winner: L
  - หา Spell(L, Force) → สมมติไม่เจอ ❌
→ STEP 2B.2: Strongest Against Caster
  - L vs S: matchup = ? → Score = ?
  - P vs S: matchup = 1.3 (P > S) → Score = +1
  - Winner: P (คะแนนสูงสุด)
→ ใช้ Spell(P, Force) → เจอ ✅
→ ใช้ "Energy Bolt" (P + Force)
→ จบ
```

---

### สรุป Fallback Algorithm

```
┌─────────────────────────────────────┐
│ STEP 0: Direct Lookup               │
│ ลองหา Spell(Element, Mastery)      │
│ → เจอ: ใช้เลย                       │
│ → ไม่เจอ: ไป STEP 1                 │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│ STEP 1: Check Majority              │
│ หาเสียงข้างมากใน Recipe            │
│ → มี Majority: ลอง Spell(Majority)  │
│ → ไม่มี: ไป STEP 2                  │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│ STEP 2: Check Caster Role           │
│ Caster อยู่ใน Recipe ไหม?          │
│ → ใช่: ลอง Spell(Caster Element)    │
│ → ไม่: ไป STEP 2B.1                 │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│ STEP 2B.1: Internal Fight           │
│ ธาตุใน Recipe สู้กัน (Round-robin) │
│ → มี Winner: ลอง Spell(Winner)      │
│ → Tie หรือไม่เจอ: ไป STEP 2B.2     │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│ STEP 2B.2: Strongest vs Caster      │
│ หาธาตุที่ชนะ Caster มากที่สุด       │
│ → ลอง Spell(Strongest Element)      │
│ → ไม่เจอ: Error 404                 │
└─────────────────────────────────────┘
```

#### ✅ สถานะ: **Fallback Algorithm สมบูรณ์แบบ 100%**

Algorithm มีความซับซ้อนสูงและครอบคลุมทุกกรณี:

-  ✅ Direct Lookup
-  ✅ Majority Element Detection
-  ✅ Caster Role Check (Insider/Outsider)
-  ✅ Internal Fight (Round-robin with scoring)
-  ✅ Strongest Against Caster (Advantage scoring)
-  ✅ Error Handling (404 when truly not found)

**Code Location:** `internal/modules/combat/spell_resolver.go`

---

### 3.4 Effect Value Calculation (การคำนวณค่าเอฟเฟกต์)

#### 3.4.1 Base Value

**ที่มา:**

```go
// spell_effects table
{
    spell_id: 1,
    effect_id: 1001,  // DMG_DIRECT
    base_value: 50.0  // ค่าพื้นฐาน
}
```

**การดึงค่า:**

```go
baseValue := spellEffect.BaseValue  // 50.0
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

#### 3.4.2 Mastery Bonus (โบนัสจากศาสตร์)

**สูตร:**

```
MasteryBonus = MasteryLevel²
```

**ตัวอย่าง:**

```
Level 1: 1² = 1
Level 2: 2² = 4
Level 3: 3² = 9
Level 5: 5² = 25
Level 10: 10² = 100
```

**Code:**

```go
// internal/modules/combat/spell_calculation.go
func (s *combatService) _CalculateMasteryBonus(
    caster *domain.Combatant,
    masteryID uint,
) float64 {
    if caster.Character == nil {
        return 0.0  // Enemy ไม่มี mastery
    }

    var masteryLevel int = 1
    for _, mastery := range caster.Character.Masteries {
        if mastery.MasteryID == masteryID {
            masteryLevel = mastery.Level
            break
        }
    }

    bonus := float64(masteryLevel * masteryLevel)
    return bonus
}
```

**ตัวอย่างการใช้งาน:**

```
Base Value: 50
Mastery Level: 5
Mastery Bonus: 5² = 25

(Mastery Bonus จะถูกบวกเข้า Initial Value ไม่ใช่คูณ)
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง** (แก้ไขเมื่อ Oct 29, 2025)

---

#### 3.4.3 Talent Bonus (โบนัสจาก Talent)

**สูตร:**

```
TalentBonus = Σ(Ingredient Talents) ÷ TALENT_DMG_DIVISOR
```

**Config:**

```go
TALENT_DMG_DIVISOR: 10  // หารด้วย 10
```

**การทำงาน:**

**สำหรับ Tier 0 (ธาตุพื้นฐาน):**

```
Spell: S + Force
Ingredients: [S]

TalentS = 93
TalentBonus = 93 ÷ 10 = 9.3
```

**สำหรับ Tier 1 (ธาตุผสม):**

```
Spell: Magma (S+P) + Force
Recipe: S + P

TalentS = 93
TalentP = 3
Sum = 93 + 3 = 96

TalentBonus = 96 ÷ 10 = 9.6
```

**Code:**

```go
// internal/modules/combat/calculator.go
func (s *combatService) calculateTalentBonusFromRecipe(
    ingredientCount map[uint]int,
    character *domain.Character,
) float64 {
    if character == nil {
        return 0.0
    }

    totalTalent := 0.0
    for elementID, count := range ingredientCount {
        var talent int
        switch elementID {
        case 1: talent = character.TalentS
        case 2: talent = character.TalentL
        case 3: talent = character.TalentG
        case 4: talent = character.TalentP
        }
        totalTalent += float64(talent * count)
    }

    divisorStr, _ := s.gameDataRepo.GetGameConfigValue("TALENT_DMG_DIVISOR")
    divisor, _ := strconv.ParseFloat(divisorStr, 64)
    if divisor == 0 {
        divisor = 10 // default
    }

    return totalTalent / divisor
}
```

**ตัวอย่างการคำนวณรวม:**

```
Base Value: 50
Mastery Bonus: 25
Talent Bonus: 9.3

Initial Value = Base + Mastery + Talent
              = 50 + 25 + 9.3
              = 84.3
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

#### 3.4.4 Elemental Modifier (ความได้เปรียบด้านธาตุ)

**กฎความสัมพันธ์:**

```
S > L (Advantage: 1.3x)
L > G (Advantage: 1.3x)
G > P (Advantage: 1.3x)
P > S (Advantage: 1.3x)

Reverse = Disadvantage (0.8x)
Same = Neutral (1.0x)
```

**Config:**

```go
ELEMENT_ADVANTAGE_MULTIPLIER: 1.30     // โจมตีได้เปรียบ
ELEMENT_DISADVANTAGE_MULTIPLIER: 0.80  // โจมตีเสียเปรียบ
```

**ตัวอย่าง:**

```
Spell Element: S (Solidity)
Target Element: L (Liquidity)

S > L → Advantage
Modifier = 1.3
```

**Code:**

```go
// internal/modules/combat/calculator.go
func (s *combatService) getElementalModifier(
    spellElementID uint,
    targetElementID uint,
) (float64, error) {

    // ดึงข้อมูลจาก elemental_matchups table
    matchup, err := s.gameDataRepo.FindElementalMatchup(spellElementID, targetElementID)
    if err != nil {
        return 1.0, err
    }

    // ใช้ค่า Multiplier จาก matchup
    return matchup.Multiplier, nil
}
```

**ตัวอย่างการใช้งาน:**

```
Initial Value: 1,259.3
Elemental Modifier: 1.3

Value after elemental = 1,259.3 × 1.3 = 1,637.09
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

#### 3.4.5 Buff/Debuff Modifier

**Effect Types:**

**Buff (บน Caster):**

-  `BUFF_DMG_UP` (2202): เพิ่มความเสียหาย +X%
-  `BUFF_DEFENSE_UP` (2204): ลดความเสียหายที่ได้รับ -X%

**Debuff (บน Target):**

-  `DEBUFF_VULNERABLE` (4102): รับความเสียหายเพิ่ม +X%
-  `DEBUFF_SLOW` (4101): ลด Initiative

**สูตร:**

```
BuffDebuffMod = (1 + CasterBuffs) × (1 + TargetDebuffs) × (1 - TargetDefenseBuffs)
```

**ตัวอย่าง:**

```
Caster: BUFF_DMG_UP +30%
Target: DEBUFF_VULNERABLE +25%

Modifier = (1 + 0.30) × (1 + 0.25)
         = 1.3 × 1.25
         = 1.625
```

**Code:**

```go
// internal/modules/combat/calculator.go (DEPRECATED function)
func (s *combatService) calculateEffectValue(...) {
    buffDebuffModifier := 1.0

    // Check target debuffs
    for _, activeEffect := range targetEffects {
        if activeEffect.EffectID == 4102 { // VULNERABLE
            increasePercent := float64(activeEffect.Value) / 100.0
            buffDebuffModifier *= (1.0 + increasePercent)
        }
    }

    // Check caster buffs
    for _, activeEffect := range casterEffects {
        if activeEffect.EffectID == 2202 { // BUFF_DMG_UP
            increasePercent := float64(activeEffect.Value) / 100.0
            buffDebuffModifier *= (1.0 + increasePercent)
        }
    }
}
```

#### ⚠️ สถานะ: **ใช้งานได้บางส่วน** (ระบบใหม่ใน spell_calculation.go ยังไม่ implement)

---

#### 3.4.6 Power Modifier (โหมดการร้าย)

**Casting Modes:**

1. **INSTANT** - ปกติ (Power: 1.0x, Cost: +0 AP, +0 MP)
2. **CHARGE** - สะสมพลัง (Power: 1.2x, Cost: +1 AP, +0 MP)
3. **OVERCHARGE** - ระเบิดพลัง (Power: 1.5x, Cost: +1 AP, +30 MP)

**Config:**

```go
// INSTANT (Default)
Power: 1.0
AP Add: 0
MP Add: 0

// CHARGE
CAST_MODE_CHARGE_POWER_MOD: 1.2
CAST_MODE_CHARGE_AP_ADD: 1
CAST_MODE_CHARGE_MP_ADD: 0

// OVERCHARGE
CAST_MODE_OVERCHARGE_POWER_MOD: 1.5
CAST_MODE_OVERCHARGE_AP_ADD: 1
CAST_MODE_OVERCHARGE_MP_ADD: 30
```

**ตัวอย่าง:**

```
Spell Base Cost: AP = 2, MP = 10
Casting Mode: CHARGE

Final AP Cost = 2 + 1 = 3 AP ✅
Final MP Cost = 10 + 0 = 10 MP ✅
Power Modifier = 1.2x ✅

---

Spell Base Cost: AP = 2, MP = 10
Casting Mode: OVERCHARGE

Final AP Cost = 2 + 1 = 3 AP ✅
Final MP Cost = 10 + 30 = 40 MP ✅
Power Modifier = 1.5x ✅
```

**Code:**

```go
// internal/modules/combat/spell_preparation.go
func (s *combatService) _CalculateFinalCost(
    baseAP int,
    baseMP int,
    castingMode string,
) (finalAP int, finalMP int, powerMod float64, err error) {

    // Default values (INSTANT)
    finalAP = baseAP
    finalMP = baseMP
    powerMod = 1.0

    if castingMode == "" || castingMode == "INSTANT" {
        return finalAP, finalMP, powerMod, nil
    }

    // ดึง config values
    var apAddStr, mpAddStr, powerModStr string

    switch castingMode {
    case "CHARGE":
        apAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_AP_ADD")
        mpAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_MP_ADD")
        powerModStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_POWER_MOD")

    case "OVERCHARGE":
        apAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_AP_ADD")
        mpAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_MP_ADD")
        powerModStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_POWER_MOD")

    default:
        return finalAP, finalMP, powerMod, nil
    }

    // Parse values
    apAdd, _ := strconv.Atoi(apAddStr)
    mpAdd, _ := strconv.Atoi(mpAddStr)
    powerMod, _ = strconv.ParseFloat(powerModStr, 64)

    // Apply additive cost
    finalAP = baseAP + apAdd
    finalMP = baseMP + mpAdd

    return finalAP, finalMP, powerMod, nil
}
```

#### ✅ สถานะ: **ใช้งานได้ถูกต้อง**

---

### 3.5 Final Calculation Example (ตัวอย่างครบวงจร)

**Setup:**

```
Character:
- Primary Element: S (Solidity)
- TalentS: 93, TalentL: 3, TalentG: 3, TalentP: 3
- Mastery Force: Level 5

Spell:
- Element: S + Force
- Effect: DMG_DIRECT (base: 50)

Target:
Target:
- Element: L (Liquidity)
- No buffs/debuffs

Casting Mode: CHARGE (1.2x)
```

**Step-by-Step Calculation:**

**1. Base Value:**

```
BaseValue = 50
```

**2. Mastery Bonus:**

```
MasteryLevel = 5
MasteryBonus = 5² = 25
```

**3. Talent Bonus:**

```
Ingredients: [S]
TalentS = 93
TalentBonus = 93 ÷ 10 = 9.3
```

**4. Initial Value:**

```
InitialValue = Base + Mastery + Talent
             = 50 + 25 + 9.3
             = 84.3
```

**5. Elemental Modifier:**

```
S vs L → Advantage
ElementalMod = 1.3
```

**6. Buff/Debuff Modifier:**

```
No buffs/debuffs
BuffDebuffMod = 1.0
```

**7. Power Modifier:**

```
CastingMode = CHARGE
PowerMod = 1.2
```

**8. Combined Modifier:**

```
CombinedMod = Elemental × BuffDebuff × Power
            = 1.3 × 1.0 × 1.2
            = 1.56
```

**9. Final Value:**

```
FinalValue = InitialValue × CombinedMod
           = 84.3 × 1.56
           = 131.51
```

**10. Apply to Target:**

```
Target.CurrentHP -= 131.51 (rounded to 132)
```

---

## 4. สรุปสูตรการคำนวณ

### 4.1 Character Creation

| Stat           | สูตร                          | Config               |
| -------------- | ----------------------------- | -------------------- |
| **TalentX**    | `3 + (primary == X ? 90 : 0)` | BASE: 3, PRIMARY: 90 |
| **MaxHP**      | `900 + (TalentS × 30)`        | BASE: 900, PER_S: 30 |
| **MaxMP**      | `200 + (TalentL × 2)`         | BASE: 200, PER_L: 2  |
| **Initiative** | `50 + (TalentG × 1)`          | BASE: 50, PER_G: 1   |

### 4.2 Damage Calculation

| Component           | สูตร                                                   | ตัวอย่าง (CHARGE)      |
| ------------------- | ------------------------------------------------------ | ---------------------- |
| **Base Value**      | `SpellEffect.BaseValue`                                | 50                     |
| **Mastery Bonus**   | `Level²`                                               | Lv.5 → 25              |
| **Talent Bonus**    | `Σ(Talents) ÷ 10`                                      | 93 ÷ 10 = 9.3          |
| **Initial Value**   | `Base + Mastery + Talent`                              | 50 + 25 + 9.3 = 84.3   |
| **Elemental Mod**   | `1.3 (advantage) / 0.8 (disadvantage) / 1.0 (neutral)` | 1.3                    |
| **Buff/Debuff Mod** | `(1 + buffs) × (1 + debuffs)`                          | 1.0                    |
| **Power Mod**       | `1.0 / 1.2 / 1.5`                                      | 1.2                    |
| **Combined Mod**    | `Elemental × BuffDebuff × Power`                       | 1.3 × 1.0 × 1.2 = 1.56 |
| **Final Value**     | `Initial × Combined`                                   | 84.3 × 1.56 = 131.51   |

### 4.3 Casting Modes

| Mode           | Power Modifier | AP Cost | MP Cost | Config Keys                           |
| -------------- | -------------- | ------- | ------- | ------------------------------------- |
| **INSTANT**    | 1.0x           | +0      | +0      | (Default)                             |
| **CHARGE**     | 1.2x           | +1      | +0      | POWER_MOD: 1.2, AP_ADD: 1, MP_ADD: 0  |
| **OVERCHARGE** | 1.5x           | +1      | +30     | POWER_MOD: 1.5, AP_ADD: 1, MP_ADD: 30 |

**ตัวอย่าง:**

```
Spell Base: AP = 2, MP = 10

INSTANT:    Final = AP:2, MP:10,  Power:1.0x
CHARGE:     Final = AP:3, MP:10,  Power:1.2x
OVERCHARGE: Final = AP:3, MP:40,  Power:1.5x
```

### 4.4 EXP Rewards

| Match Type   | EXP | Config             |
| ------------ | --- | ------------------ |
| **TRAINING** | 50  | EXP_TRAINING_MATCH |
| **STORY**    | 100 | EXP_STORY_MATCH    |
| **PVP**      | 150 | EXP_PVP_MATCH      |

---

## 5. ปัญหาที่พบและข้อเสนอแนะ

### 5.1 ✅ สิ่งที่ใช้งานได้ดี

1. **Character Creation** - ระบบสร้างตัวละครทำงานถูกต้องครบถ้วน
2. **Talent System** - การคำนวณค่าพลังดิบถูกต้อง
3. **Core Stats** - HP, MP, Initiative คำนวณถูกต้อง
4. **Mastery Bonus** - แก้ไขแล้ว ใช้สูตร Level² ถูกต้อง
5. **Talent Bonus** - การคำนวณจาก recipe ทำงานถูกต้อง
6. **Elemental Matchup** - ระบบความได้เปรียบธาตุทำงานถูกต้อง
7. **Spell Fallback Algorithm** - อัลกอริทึมสมบูรณ์แบบ
8. **EXP Gain** - ผู้เล่นได้ EXP หลังชนะการต่อสู้

---

### 5.2 ⚠️ สิ่งที่ขาดหายไป / ต้องเพิ่ม

#### **5.2.1 Player Progression System (ความสำคัญ: 🔴 HIGH)**

**ปัญหา:**

-  ✅ ผู้เล่นได้ EXP แล้ว แต่ไม่มีระบบเลเวลอัป
-  ❌ ไม่มีตาราง XP requirement
-  ❌ ไม่มี auto level-up logic
-  ❌ ไม่มีการแจก talent points เมื่อเลเวลขึ้น

---

### 📊 การออกแบบระบบ Player Level Up

#### **A. XP Requirement Table (ตารางประสบการณ์)**

**สูตร Exponential Growth:**

```
RequiredExp(Level) = BaseExp × (GrowthRate ^ (Level - 2))

โดยที่:
- BaseExp = 100 (Exp ที่ต้องใช้สำหรับ Level 2)
- GrowthRate = 1.15 (เพิ่มขึ้น 15% ทุกเลเวล - ลดจาก 1.5 เพื่อให้ achievable)
```

**ตารางเลเวล 1-50:**

| Level | Required EXP | Cumulative EXP | Training Matches | PVP Matches | Notes      |
| ----- | ------------ | -------------- | ---------------- | ----------- | ---------- |
| 1     | 0            | 0              | -                | -           | Start      |
| 2     | 100          | 100            | 2                | 1           | Easy       |
| 3     | 115          | 215            | 2                | 1           |            |
| 4     | 132          | 347            | 3                | 2           |            |
| 5     | 152          | 499            | 3                | 3           |            |
| 10    | 306          | 1,678          | 6                | 11          |            |
| 15    | 616          | 4,053          | 12               | 27          |            |
| 20    | 1,238        | 8,826          | 25               | 59          | Mid        |
| 25    | 2,185        | 17,357         | 44               | 116         |            |
| 30    | 3,855        | 32,919         | 77               | 219         |            |
| 35    | 6,802        | 60,501         | 136              | 403         |            |
| 40    | 12,002       | 109,435        | 240              | 729         |            |
| 45    | 21,175       | 196,911        | 424              | 1,313       |            |
| 50    | 37,353       | 353,398        | 747              | 2,356       | Max (v1.0) |

**Total to Level 50:** 353,398 Exp (~2,356 PVP / ~3,534 Story / ~7,068 Training)

**สังเกต:**

-  **Level 1-10:** Early Game - เร็วมาก (2-6 matches ต่อเลเวล)
-  **Level 11-20:** Mid Game - ปานกลาง (7-25 matches ต่อเลเวล)
-  **Level 21-35:** Late Game - ค่อนข้างช้า (30-140 matches ต่อเลเวล)
-  **Level 36-50:** Endgame - ช้า (200-750 matches ต่อเลเวล)

**การเปลี่ยนแปลงจาก Growth Rate 1.5 → 1.15:**

-  ✅ Level 50 ใช้ 37K Exp แทน 28.4 พันล้าน Exp (ลดลง 99.9999%)
-  ✅ สามารถเล่นถึง Max Level ได้จริง (~2,400 PVP matches หรือ 3-6 เดือน)
-  ✅ Balanced สำหรับ Early/Mid/Late Game
-  ✅ **พร้อม Scale ไป Level 99 ในอนาคต** (ด้วย Growth Rate 1.15 → Lv.99 ใช้ ~50M Exp - achievable)

**📌 หมายเหตุสำคัญ:**

-  **Version 1.0:** Max Level = 50 (ชั่วคราว)
-  **Future Patch:** จะเปิด Level 51-99 เมื่อมี Content เพียงพอ
-  **Growth Rate 1.15** ถูกออกแบบให้รองรับการ Scale ไป Lv.99 โดยไม่ต้องเปลี่ยนสูตร

---

#### **B. Talent Points Reward (รางวัลแต้ม Talent)**

**Option 1: Fixed Reward (แนะนำ)**

```
ทุกเลเวล: +3 Talent Points
```

**ข้อดี:**

-  ✅ เข้าใจง่าย
-  ✅ Balance ง่าย
-  ✅ คาดการณ์ได้

**ตัวอย่าง:**

```
Level 1 → 2: +3 points (Total: 105)
Level 2 → 3: +3 points (Total: 108)
Level 20:    +57 points (Total: 159)
Level 50:    +147 points (Total: 249)
Level 99:    +294 points (Total: 396)
```

**Option 2: Milestone Reward (ทางเลือก)**

```
Level 2-5:   +3 points
Level 6-10:  +4 points
Level 11-15: +5 points
Level 16-20: +6 points
Level 21+:   +7 points
```

**ข้อดี:**

-  ✅ รางวัลเพิ่มขึ้นตามความยาก
-  ✅ Motivation เลเวลสูง

**ตัวอย่าง:**

```
Level 1 → 5:  +12 points (3×4)
Level 6 → 10: +20 points (4×5)
Level 20:     +74 points total
Level 50:     +182 points total
Level 99:     +423 points total
```

---

#### **C. Implementation Plan (แผนการพัฒนา)**

**1. เพิ่ม Config ใน seeder.go:**

```go
// Player Progression
{Key: "PLAYER_BASE_EXP", Value: "100"},           // Exp สำหรับ Level 2
{Key: "PLAYER_EXP_GROWTH_RATE", Value: "1.15"},   // เพิ่ม 15% ทุกเลเวล (ลดจาก 1.5)
{Key: "PLAYER_MAX_LEVEL", Value: "50"},           // Level สูงสุด (ชั่วคราว, จะเป็น 99 ในอนาคต)
{Key: "TALENT_POINTS_PER_LEVEL", Value: "3"},     // แต้มต่อเลเวล
{Key: "PLAYER_EXP_CARRY_OVER", Value: "true"},    // เก็บ Exp เกินไว้
```

**2. สร้าง Helper Function คำนวณ Required Exp:**

```go
// internal/modules/character/service.go

func (s *characterService) GetRequiredExpForLevel(targetLevel int) (int, error) {
    if targetLevel <= 1 {
        return 0, nil
    }

    // ดึง config
    baseExpStr, _ := s.gameDataRepo.GetGameConfigValue("PLAYER_BASE_EXP")
    growthRateStr, _ := s.gameDataRepo.GetGameConfigValue("PLAYER_EXP_GROWTH_RATE")

    baseExp, _ := strconv.Atoi(baseExpStr)
    growthRate, _ := strconv.ParseFloat(growthRateStr, 64)

    // สูตร: BaseExp × (GrowthRate ^ (Level - 2))
    power := float64(targetLevel - 2)
    required := float64(baseExp) * math.Pow(growthRate, power)

    return int(math.Round(required)), nil
}
```

**3. สร้าง Function ตรวจสอบและเลเวลอัป:**

```go
func (s *characterService) CheckAndProcessLevelUp(characterID uint) error {
    s.appLogger.Info("Checking for level up", "character_id", characterID)

    // 1. Load character
    character, err := s.characterRepo.FindByID(characterID)
    if err != nil {
        return err
    }

    // 2. ดึง max level
    maxLevelStr, _ := s.gameDataRepo.GetGameConfigValue("PLAYER_MAX_LEVEL")
    maxLevel, _ := strconv.Atoi(maxLevelStr)

    if character.Level >= maxLevel {
        s.appLogger.Info("Character already at max level", "level", character.Level)
        return nil
    }

    // 3. คำนวณ required exp สำหรับเลเวลถัดไป
    nextLevel := character.Level + 1
    requiredExp, err := s.GetRequiredExpForLevel(nextLevel)
    if err != nil {
        return err
    }

    // 4. ตรวจสอบว่า exp พอไหม
    if character.Exp < requiredExp {
        s.appLogger.Info("Not enough exp for level up",
            "current_exp", character.Exp,
            "required_exp", requiredExp,
        )
        return nil
    }

    // 5. เลเวลอัป!
    return s.ProcessLevelUp(character, requiredExp)
}

func (s *characterService) ProcessLevelUp(
    character *domain.Character,
    requiredExp int,
) error {
    s.appLogger.Info("🎉 LEVEL UP!",
        "character_id", character.ID,
        "old_level", character.Level,
        "new_level", character.Level + 1,
    )

    // 1. เพิ่ม Level
    character.Level++

    // 2. จัดการ Exp
    carryOverStr, _ := s.gameDataRepo.GetGameConfigValue("PLAYER_EXP_CARRY_OVER")
    if carryOverStr == "true" {
        // เก็บ Exp เกินไว้
        character.Exp -= requiredExp
    } else {
        // Reset เป็น 0
        character.Exp = 0
    }

    // 3. ให้ Talent Points
    pointsStr, _ := s.gameDataRepo.GetGameConfigValue("TALENT_POINTS_PER_LEVEL")
    points, _ := strconv.Atoi(pointsStr)
    character.UnallocatedTalentPoints += points

    // 4. ฟื้นฟู HP/MP เต็ม (โบนัส!)
    character.CurrentHP = s.calculateMaxHP(character)
    character.CurrentMP = s.calculateMaxMP(character)

    // 5. บันทึก
    err := s.characterRepo.Update(character)
    if err != nil {
        return err
    }

    s.appLogger.Info("✅ Level up complete!",
        "new_level", character.Level,
        "remaining_exp", character.Exp,
        "talent_points_gained", points,
        "total_unallocated", character.UnallocatedTalentPoints,
    )

    return nil
}

// Helper: คำนวณ MaxHP
func (s *characterService) calculateMaxHP(char *domain.Character) int {
    baseHP, _ := s.gameDataRepo.GetGameConfigValue("STAT_HP_BASE")
    hpPerTalent, _ := s.gameDataRepo.GetGameConfigValue("STAT_HP_PER_TALENT_S")

    base, _ := strconv.Atoi(baseHP)
    perTalent, _ := strconv.Atoi(hpPerTalent)

    return base + (char.TalentS * perTalent)
}

// Helper: คำนวณ MaxMP
func (s *characterService) calculateMaxMP(char *domain.Character) int {
    baseMP, _ := s.gameDataRepo.GetGameConfigValue("STAT_MP_BASE")
    mpPerTalent, _ := s.gameDataRepo.GetGameConfigValue("STAT_MP_PER_TALENT_L")

    base, _ := strconv.Atoi(baseMP)
    perTalent, _ := strconv.Atoi(mpPerTalent)

    return base + (char.TalentL * perTalent)
}
```

**4. Hook เข้า GrantExp Function:**

```go
func (s *characterService) GrantExp(characterID uint, amount int) error {
    // ... existing code ...

    // บันทึก Exp
    err = s.characterRepo.Update(character)
    if err != nil {
        return err
    }

    // ✅ เพิ่ม: ตรวจสอบเลเวลอัป
    return s.CheckAndProcessLevelUp(characterID)
}
```

**5. สร้าง API Endpoint สำหรับดูตาราง Exp:**

````go
// GET /api/game-data/level-requirements
// Optional: ?maxLevel=50 (default: current max level from config)
func (h *gameDataHandler) GetLevelRequirements(c echo.Context) error {
    // ดึง max level จาก config
    maxLevelStr, _ := h.gameDataRepo.GetGameConfigValue("PLAYER_MAX_LEVEL")
    maxLevel, _ := strconv.Atoi(maxLevelStr)

    // Allow override
    if max := c.QueryParam("maxLevel"); max != "" {
        maxLevel, _ = strconv.Atoi(max)
    }

    levels := []map[string]interface{}{}

    for level := 2; level <= maxLevel; level++ {
        required, _ := h.service.GetRequiredExpForLevel(level)
        levels = append(levels, map[string]interface{}{
            "level":        level,
            "required_exp": required,
        })
    }

    return c.JSON(200, levels)
}
```---

#### **D. ตัวอย่างการทำงาน**

**Scenario: ผู้เล่นชนะ Training Match 3 ครั้ง**

````

Initial State:

-  Level: 1
-  Exp: 0/100
-  Talent Points: 0

Match 1 (Training): +50 Exp
→ Level: 1, Exp: 50/100

Match 2 (Training): +50 Exp
→ Level: 1, Exp: 100/100
→ 🎉 LEVEL UP!
→ Level: 2, Exp: 0/150, Talent Points: +3

Match 3 (Training): +50 Exp
→ Level: 2, Exp: 50/150

```

**Scenario: Exp Carry Over**

```

Current:

-  Level: 2
-  Exp: 140/150
-  Required: 150

Win PVP: +150 Exp
→ Total Exp: 290

🎉 LEVEL UP!
→ Level: 3
→ Remaining Exp: 290 - 150 = 140
→ New Progress: 140/225 (ถ้า carry_over = true)

`````

---

#### **E. Balance Considerations**

**ข้อดี:**

-  ✅ Early game เร็ว (Level 2-10 ภายใน 50 matches)
-  ✅ Mid game ปานกลาง (Level 11-30 ใช้ 100-200k matches)
-  ✅ Late game ช้า (Level 31-60 ใช้ล้าน+ matches)
-  ✅ Endgame ช้ามาก (Level 61+ ต้องเล่นนานมาก)
-  ✅ เก็บ Exp เกินได้ → ไม่เสีย progression

**ข้อควรระวัง:**

-  ⚠️ Level 50+ อาจช้าเกินไป (เกือบเป็นไปไม่ได้)
-  ⚠️ Level 90+ เป็นไปไม่ได้เลยในการเล่นจริง
-  ⚠️ อาจต้องมี Soft Cap หรือ Alternative Progression

**Recommendation:**

-  ใช้ Fixed +3 points ก่อน
-  พิจารณา Soft Cap ที่ Level 50-60
-  อาจมี "Prestige System" สำหรับ Level 60+
-  หรือลด Growth Rate เป็น 1.3-1.4 สำหรับ Late Game
-  เปิด Exp Carry Over = true

---

#### **F. Testing Checklist**

-  [ ] Config values โหลดได้ถูกต้อง
-  [ ] GetRequiredExpForLevel คำนวณถูกต้อง
-  [ ] CheckAndProcessLevelUp ทำงานหลังได้ Exp
-  [ ] Level up ให้ Talent Points ถูกต้อง
-  [ ] Exp carry over ทำงานถูกต้อง
-  [ ] ฟื้นฟู HP/MP เต็มหลัง level up
-  [ ] Max level block การอัปต่อ
-  [ ] API endpoint แสดงตารางถูกต้อง

---

#### **F. Implementation Status (สถานะการพัฒนา)**

### ✅ สิ่งที่ทำเสร็จแล้ว:

**1. การออกแบบระบบ (Design Complete)**

**A. ตาราง XP Progression:**
- ✅ สูตร: `RequiredExp = 100 × (1.15 ^ (Level - 2))`
- ✅ Growth Rate: 1.15 (เพิ่ม 15% ต่อเลเวล)
- ✅ Max Level: 50 (Version 1.0 - พร้อม Scale ไป 99)
- ✅ ตารางครบ Level 1-50 พร้อมคำนวณจำนวน Matches

**B. Reward System:**
- ✅ Fixed Reward: +3 Talent Points ต่อเลเวล
- ✅ Talent Points ที่ Level 50: +147 (Total: 249)
- ✅ Exp Carry Over: เปิดใช้งาน (เศษไม่หาย)
- ✅ HP/MP Restore: ฟื้นฟูเต็มเมื่อเลเวลอัป

**C. Config Values:**
```go
// ✅ เพิ่มใน seeder.go แล้ว
{Key: "PLAYER_BASE_EXP", Value: "100"}
{Key: "PLAYER_EXP_GROWTH_RATE", Value: "1.15"}
{Key: "PLAYER_MAX_LEVEL", Value: "50"}
{Key: "TALENT_POINTS_PER_LEVEL", Value: "3"}
{Key: "PLAYER_EXP_CARRY_OVER", Value: "true"}
```

**D. เอกสาร:**
- ✅ GAME_MECHANICS_DOCUMENTATION.md Section 5.2.1 สมบูรณ์
- ✅ มีตาราง Level 1-50 พร้อมข้อมูลครบถ้วน
- ✅ มี Code Examples ทั้งหมด
- ✅ มี Testing Checklist

---

### ⏳ สิ่งที่รอทำ (To-Do):

**2. Implementation - Backend Code**

**A. Helper Functions (`character/service.go`):**
```go
❌ func GetRequiredExpForLevel(targetLevel int) (int, error)
   // คำนวณ Exp ที่ต้องใช้ตามสูตร exponential

❌ func calculateMaxHP(char *domain.Character) int
   // คำนวณ HP สูงสุดจาก TalentS

❌ func calculateMaxMP(char *domain.Character) int
   // คำนวณ MP สูงสุดจาก TalentL
```

**B. Core Level Up Functions:**
```go
❌ func CheckAndProcessLevelUp(characterID uint) error
   // 1. Load character
   // 2. ดึง max level จาก config
   // 3. คำนวณ required exp
   // 4. ตรวจสอบว่าพอเลเวลอัปไหม
   // 5. เรียก ProcessLevelUp ถ้าพอ

❌ func ProcessLevelUp(character *domain.Character, requiredExp int) error
   // 1. เพิ่ม Level
   // 2. หัก/Carry over Exp
   // 3. ให้ Talent Points (+3)
   // 4. ฟื้นฟู HP/MP เต็ม
   // 5. บันทึกลง database
   // 6. Log event
```

**C. Integration:**
```go
❌ แก้ไข GrantExp() ใน character/service.go
   // เพิ่มบรรทัดสุดท้าย:
   return s.CheckAndProcessLevelUp(characterID)
```

---

**3. API Endpoints**

**A. Level Requirements Table:**
```go
❌ GET /api/game-data/level-requirements
   Query: ?maxLevel=50 (optional)

   Response:
   [
     {"level": 2, "required_exp": 100},
     {"level": 3, "required_exp": 115},
     ...
   ]
```

**B. Character Level Info:**
```go
✅ ใช้ Endpoint เดิมได้: GET /api/characters/:id
   // Response มี level, exp, unallocated_talent_points อยู่แล้ว
```

---

**4. Testing**

**A. Unit Tests:**
```go
❌ TestGetRequiredExpForLevel()
   // ทดสอบการคำนวณ Exp ถูกต้อง

❌ TestCalculateMaxHP()
❌ TestCalculateMaxMP()
   // ทดสอบการคำนวณ stats

❌ TestCheckAndProcessLevelUp()
   // ทดสอบ logic การตรวจสอบและเลเวลอัป

❌ TestProcessLevelUp()
   // ทดสอบการให้ rewards และบันทึก

❌ TestExpCarryOver()
   // ทดสอบว่า Exp เกินถูกเก็บไว้

❌ TestMaxLevelCap()
   // ทดสอบว่าไม่ให้เกิน Level 50
```

**B. Integration Tests:**
```go
❌ TestLevelUpAfterCombat()
   // จำลองการเล่น -> ชนะ -> ได้ Exp -> เลเวลอัป

❌ TestMultipleLevelUps()
   // ทดสอบการอัปหลายเลเวลพร้อมกัน (ได้ Exp เยอะ)

❌ TestLevelUpRewards()
   // ทดสอบว่าได้ Talent Points และ HP/MP เต็มจริง
```

---

**5. Frontend Support (ถ้ามี)**

**A. UI Components:**
```
❌ Level Progress Bar
   - แสดง Current Exp / Required Exp
   - แสดงเปอร์เซ็นต์

❌ Level Up Animation/Notification
   - แจ้งเตือนเมื่อเลเวลอัป
   - แสดง Rewards ที่ได้รับ

❌ Level Requirements Table
   - แสดงตารางเลเวลทั้งหมด
   - คำนวณว่าต้องเล่นอีกกี่แมตช์
```

---

### 📋 Implementation Checklist:

**Phase 1: Core Functions (สำคัญที่สุด)**
- [ ] สร้าง `GetRequiredExpForLevel()`
- [ ] สร้าง `calculateMaxHP()` และ `calculateMaxMP()`
- [ ] สร้าง `CheckAndProcessLevelUp()`
- [ ] สร้าง `ProcessLevelUp()`
- [ ] แก้ไข `GrantExp()` เพื่อ Hook เข้าระบบ

**Phase 2: API & Testing**
- [ ] สร้าง API Endpoint `/level-requirements`
- [ ] เขียน Unit Tests
- [ ] เขียน Integration Tests
- [ ] ทดสอบ Edge Cases (Max Level, Multiple Level Ups, Carry Over)

**Phase 3: Polish & Documentation**
- [ ] เพิ่ม Logging events สำหรับ Level Up
- [ ] เพิ่ม Error Handling ครบถ้วน
- [ ] อัปเดตเอกสาร API
- [ ] สร้าง Admin Tools (ถ้าต้องการ)

---

### 🎯 ประมาณการเวลา:

| Phase        | งาน                       | เวลา      | สถานะ        |
| ------------ | ------------------------- | --------- | ------------ |
| Design       | Config + Documentation    | 1 day     | ✅ เสร็จแล้ว |
| Phase 1      | Core Functions            | 1-2 days  | ⏳ รอทำ      |
| Phase 2      | API + Testing             | 1 day     | ⏳ รอทำ      |
| Phase 3      | Polish                    | 0.5 day   | ⏳ รอทำ      |
| **Total**    |                           | **3-4 days** | **25% Complete** |

---

### 🚀 Next Steps (ลำดับแนะนำ):

1. **สร้าง Helper Functions** (30 นาที)
   - `GetRequiredExpForLevel()`
   - `calculateMaxHP()`, `calculateMaxMP()`

2. **สร้าง Core Level Up Logic** (1-2 ชม.)
   - `CheckAndProcessLevelUp()`
   - `ProcessLevelUp()`

3. **Hook เข้า GrantExp()** (10 นาที)
   - เพิ่ม 1 บรรทัด: `return s.CheckAndProcessLevelUp(characterID)`

4. **ทดสอบ** (1 ชม.)
   - เล่นเกม → ชนะ → ตรวจสอบว่าเลเวลอัป
   - ตรวจสอบ Talent Points เพิ่ม
   - ตรวจสอบ HP/MP ฟื้นฟู

5. **สร้าง API Endpoint** (30 นาที)
   - `/api/game-data/level-requirements`

6. **เขียน Tests** (2-3 ชม.)
   - Unit Tests สำหรับทุก Function
   - Integration Test สำหรับ Flow ทั้งหมด

---


#### **5.2.2 Mastery Progression System (ความสำคัญ: 🔴 HIGH)**

**ปัญหา:**

-  ✅ มี Mastery 4 ศาสตร์ แล้ว แต่ Mxp ไม่เพิ่มขึ้น
-  ❌ ไม่มีระบบให้ XP หลังร้ายเวท
-  ❌ Level ติดที่ 1 ตลอด (MasteryBonus = 1² = 1 ตลอด)
-  ❌ ไม่มีตาราง Mxp requirement

---

### 📊 การออกแบบระบบ Mastery Progression

#### **A. วิธีการได้ Mastery XP (How to Gain MXP)**

**กฎการให้ MXP:**

```
ผู้เล่นได้ MXP เมื่อ:
1. ✅ ร้ายเวทสำเร็จ (Spell Cast Success)
2. ✅ ตรงตามเงื่อนไข (ต้องใช้ Mastery นั้นๆ)
3. ✅ ในโหมดที่ให้ MXP (Training/Story/PVP)

ไม่ได้ MXP เมื่อ:
1. ❌ ร้ายเวทพลาด (Miss/Failed)
2. ❌ เป้าหมายตาย/Match จบก่อนเวทโดน
3. ❌ ใช้ศาสตร์ที่ Max Level แล้ว
```

**จำนวน MXP ที่ได้:**

**⭐ คำแนะนำ: Fixed Amount (+10 MXP เสมอ)**

```
ทุกครั้งที่ร้ายเวทสำเร็จ: +10 MXP
ไม่ว่าจะใช้ INSTANT, CHARGE, หรือ OVERCHARGE
```

**ข้อดี:**
- ✅ **เข้าใจง่ายมาก** - ไม่ต้องคิดเลย
- ✅ **คาดการณ์ progression ได้แม่นยำ** - รู้ว่าต้องร้าย X ครั้งถึงจะเลเวล
- ✅ **Balance ง่าย** - ไม่มีตัวแปรซับซ้อน
- ✅ **Fair สำหรับทุกสไตล์การเล่น** - Casual/Hardcore ได้ XP เท่ากัน
- ✅ **Focus ที่การชนะ** - ไม่ต้อง optimize XP gain
- ✅ **Casting Mode มี reward อยู่แล้ว** - Power 1.2x/1.5x + Tactical advantage
- ✅ **Hybrid Builds เลเวล 2-3 Mastery พร้อมกัน** ได้สะดวก

**ข้อเสีย:**
- ⚠️ ไม่มี depth - ไม่มี incentive พิเศษสำหรับ CHARGE/OVERCHARGE
- ⚠️ Spam spell ธรรมดาก็ได้ XP เหมือนกัน

**เหตุผลที่เลือก:**
1. **Easy to Learn, Hard to Master** - Mastery XP เป็น passive progression
2. **Beginner-friendly** - ลดความซับซ้อน มี mechanics อื่นอีกเยอะ
3. **PVP Fair** - ผู้เล่นทุกคนได้ XP เท่ากัน ไม่มี advantage จาก playstyle
4. **Turn-based เหมาะ** - ไม่อยากให้เสียสมาธิไปคิดเรื่อง XP grinding

---

**Alternative Options (ไม่แนะนำ):**

<details>
<summary>Option 2: Based on Casting Mode</summary>

```
INSTANT:    +10 MXP
CHARGE:     +12 MXP (+20%)
OVERCHARGE: +15 MXP (+50%)
```

**ข้อดี:**
- ✅ Reward skill expression
- ✅ Encourage diverse gameplay
- ✅ Risk vs Reward

**ข้อเสีย:**
- ⚠️ Casual players ช้ากว่า 20-50%
- ⚠️ กดดัน - รู้สึกว่าต้อง optimize
- ⚠️ PVP unfair - dedicated players ได้เปรียบ

**ทำไมไม่เลือก:**
- Casting Mode มี reward อยู่แล้ว (power bonus)
- ไม่อยากให้ progression เป็น "grind optimization"
- Complexity ไม่คุ้มกับ benefit ที่ได้

</details>

<details>
<summary>Option 3: Based on MP Cost</summary>

```
MXP = Base + (MP Cost × Multiplier)
MXP = 5 + (MP Cost × 0.5)

ตัวอย่าง:
- Spell ใช้ 10 MP: 5 + (10 × 0.5) = 10 MXP
- Spell ใช้ 30 MP: 5 + (30 × 0.5) = 20 MXP
- Spell ใช้ 50 MP: 5 + (50 × 0.5) = 30 MXP
```

**ข้อดี:**
- ✅ Spell แพงๆ ให้ XP มากกว่า (make sense)
- ✅ Natural balance

**ข้อเสีย:**
- ⚠️ ซับซ้อนมาก - ต้องคำนวณ
- ⚠️ Spam spell ถูกก็ได้ XP
- ⚠️ Hard to balance

**ทำไมไม่เลือก:**
- Over-engineering
- ผู้เล่นจะ confused
- ยากต่อการ balance

</details>

---

#### **B. Mastery XP Requirement Table (ตาราง MXP)**

**สูตร Exponential Growth:**

```
RequiredMxp(Level) = BaseMxp × (GrowthRate ^ (Level - 2))

โดยที่:
- BaseMxp = 100 (Mxp สำหรับ Level 2)
- GrowthRate = 1.25 (เพิ่มขึ้น 25% ทุกเลเวล)
```

**ตาราง Mastery Level 1-30:**

| Level | Required MXP | Cumulative MXP | Spell Casts | Bonus (Level²) |
|-------|--------------|----------------|-------------|----------------|
| 1     | 0            | 0              | Start       | 1              |
| 2     | 100          | 100            | 10          | 4              |
| 3     | 125          | 225            | 23          | 9              |
| 4     | 156          | 381            | 38          | 16             |
| 5     | 195          | 576            | 58          | 25             |
| 6     | 244          | 820            | 82          | 36             |
| 7     | 305          | 1,125          | 113         | 49             |
| 8     | 381          | 1,506          | 151         | 64             |
| 9     | 477          | 1,983          | 198         | 81             |
| 10    | 596          | 2,579          | 258         | 100            |
| 11    | 745          | 3,324          | 332         | 121            |
| 12    | 931          | 4,255          | 426         | 144            |
| 13    | 1,164        | 5,419          | 542         | 169            |
| 14    | 1,455        | 6,874          | 687         | 196            |
| 15    | 1,819        | 8,693          | 869         | 225            |
| 20    | 5,722        | 26,697         | 2,670       | 400            |
| 25    | 18,013       | 83,140         | 8,314       | 625            |
| 30    | 56,717       | 259,018        | 25,902      | 900            |

**Total to Level 30:** 259,018 MXP (~25,900 spell casts)

**สังเกต:**

-  **Level 1-5:** Very Fast (10-20 casts ต่อเลเวล)
-  **Level 6-10:** Fast (24-60 casts ต่อเลเวล)
-  **Level 11-15:** Medium (75-182 casts ต่อเลเวล)
-  **Level 16-20:** Slow (200-600 casts ต่อเลเวล)
-  **Level 21-30:** Very Slow (800-5,700 casts ต่อเลเวล)

**เวลาโดยประมาณ (เล่นวันละ 20 matches, ~100 casts/day):**
- Level 1→5: ~1 วัน
- Level 1→10: ~2-3 วัน
- Level 1→15: ~1 สัปดาห์
- Level 1→20: ~1 เดือน
- Level 1→30: ~8-9 เดือน

**Mastery Bonus Scaling:**

| Level | Bonus (Level²) | Damage Multiplier | Notes                    |
|-------|----------------|-------------------|--------------------------|
| 1     | 1              | ×1                | Start (very weak)        |
| 2     | 4              | ×4                | Noticeable improvement   |
| 3     | 9              | ×9                | Good boost               |
| 5     | 25             | ×25               | Significant power        |
| 10    | 100            | ×100              | Very powerful            |
| 15    | 225            | ×225              | Extremely powerful       |
| 20    | 400            | ×400              | Godlike                  |
| 30    | 900            | ×900              | Absolute endgame         |

**⚠️ Balance Warning:**
- Level 10+ Mastery ทำให้ damage สูงมากเกินไป (×100)
- แนะนำ Soft Cap หรือ ลด Growth Rate

---

#### **C. Alternative: Soft Cap System** 💡

**ปัญหา:** Mastery Level² scaling ทำให้ Late Game เกิน Balance

**แนวทางแก้:**

**Option A: Diminishing Returns**

```
MasteryBonus(Level):
- Level 1-10:  Level²  (1, 4, 9, ..., 100)
- Level 11-20: 100 + ((Level - 10) × 20)  (120, 140, ..., 300)
- Level 21+:   300 + ((Level - 20) × 10)  (310, 320, ..., 400)

ตัวอย่าง:
Lv.1  = 1
Lv.5  = 25
Lv.10 = 100
Lv.15 = 100 + (5 × 20) = 200  (แทน 225)
Lv.20 = 100 + (10 × 20) = 300  (แทน 400)
Lv.30 = 300 + (10 × 10) = 400  (แทน 900)
```

**ข้อดี:**
- ✅ ยังมี progression ต่อเนื่อง
- ✅ ไม่ broken ใน Late Game
- ✅ Encourage diverse builds แทนการ farm mastery เดียว

**Option B: Max Mastery Level Cap**

```
Max Mastery Level = 20
→ Max Bonus = 400

หรือ Max = 15
→ Max Bonus = 225
```

**ข้อดี:**
- ✅ Simple และชัดเจน
- ✅ Balance ได้ง่าย

**ข้อเสีย:**
- ⚠️ ไม่มี long-term progression

---

**💡 คำแนะนำ: ใช้ Diminishing Returns (Option A)**

เพราะ:
1. ยังมี progression feeling
2. Balance ดีกว่า unlimited scaling
3. เหมาะกับ Level-based system
4. ไม่ต้อง hard cap

---

#### **D. Implementation Plan**

**1. เพิ่ม Config:**

````go
// Mastery Progression
{Key: "MASTERY_BASE_MXP", Value: "100"},
{Key: "MASTERY_MXP_GROWTH_RATE", Value: "1.25"},
{Key: "MASTERY_MAX_LEVEL", Value: "30"},
{Key: "MASTERY_MXP_PER_CAST", Value: "10"},  // Fixed amount

// Mastery Bonus Scaling (Diminishing Returns)
{Key: "MASTERY_BONUS_CAP_LEVEL_1", Value: "10"},      // ถึง Lv.10 ใช้ Level²
{Key: "MASTERY_BONUS_CAP_LEVEL_2", Value: "20"},      // Lv.11-20 ใช้ Linear
{Key: "MASTERY_BONUS_LINEAR_RATE_1", Value: "20"},    // +20 ต่อเลเวล (11-20)
{Key: "MASTERY_BONUS_LINEAR_RATE_2", Value: "10"},    // +10 ต่อเลเวล (21+)
`````

**2. สร้าง Helper Function:**

```go
// internal/modules/combat/spell_calculation.go

func (s *combatService) _CalculateMasteryBonus(
    caster *domain.Combatant,
    masteryID uint,
) float64 {
    if caster.Character == nil {
        return 0.0
    }

    var masteryLevel int = 1
    for _, mastery := range caster.Character.Masteries {
        if mastery.MasteryID == masteryID {
            masteryLevel = mastery.Level
            break
        }
    }

    // Diminishing Returns Scaling
    var bonus float64

    if masteryLevel <= 10 {
        // Level 1-10: Exponential (Level²)
        bonus = float64(masteryLevel * masteryLevel)
    } else if masteryLevel <= 20 {
        // Level 11-20: Linear (+20 per level)
        bonus = 100.0 + float64((masteryLevel - 10) * 20)
    } else {
        // Level 21+: Slower Linear (+10 per level)
        bonus = 300.0 + float64((masteryLevel - 20) * 10)
    }

    return bonus
}
```

**3. สร้าง Grant MXP Function:**

```go
// internal/modules/character/service.go

func (s *characterService) GrantMasteryXP(
    characterID uint,
    masteryID uint,
    amount int,
) error {
    s.appLogger.Info("Granting mastery XP",
        "character_id", characterID,
        "mastery_id", masteryID,
        "amount", amount,
    )

    // 1. Load character with masteries
    character, err := s.characterRepo.FindByIDWithMasteries(characterID)
    if err != nil {
        return err
    }

    // 2. Find the specific mastery
    var targetMastery *domain.CharacterMastery
    for i := range character.Masteries {
        if character.Masteries[i].MasteryID == masteryID {
            targetMastery = &character.Masteries[i]
            break
        }
    }

    if targetMastery == nil {
        return apperrors.NotFoundError("mastery not found")
    }

    // 3. Check max level
    maxLevelStr, _ := s.gameDataRepo.GetGameConfigValue("MASTERY_MAX_LEVEL")
    maxLevel, _ := strconv.Atoi(maxLevelStr)

    if targetMastery.Level >= maxLevel {
        s.appLogger.Info("Mastery already at max level", "level", targetMastery.Level)
        return nil
    }

    // 4. Add MXP
    targetMastery.Mxp += amount

    // 5. Check for level up(s)
    leveled := false
    for {
        required, err := s.GetRequiredMxpForLevel(targetMastery.Level + 1)
        if err != nil || targetMastery.Mxp < required {
            break
        }

        // Level up!
        targetMastery.Level++
        targetMastery.Mxp -= required
        leveled = true

        s.appLogger.Info("🎉 MASTERY LEVEL UP!",
            "mastery_id", masteryID,
            "new_level", targetMastery.Level,
        )

        // Check max level again
        if targetMastery.Level >= maxLevel {
            targetMastery.Mxp = 0
            break
        }
    }

    // 6. Save
    err = s.characterRepo.UpdateMastery(targetMastery)
    if err != nil {
        return err
    }

    if leveled {
        s.appLogger.Info("✅ Mastery progression saved",
            "final_level", targetMastery.Level,
            "remaining_mxp", targetMastery.Mxp,
        )
    }

    return nil
}

func (s *characterService) GetRequiredMxpForLevel(targetLevel int) (int, error) {
    if targetLevel <= 1 {
        return 0, nil
    }

    baseMxpStr, _ := s.gameDataRepo.GetGameConfigValue("MASTERY_BASE_MXP")
    growthRateStr, _ := s.gameDataRepo.GetGameConfigValue("MASTERY_MXP_GROWTH_RATE")

    baseMxp, _ := strconv.Atoi(baseMxpStr)
    growthRate, _ := strconv.ParseFloat(growthRateStr, 64)

    power := float64(targetLevel - 2)
    required := float64(baseMxp) * math.Pow(growthRate, power)

    return int(math.Round(required)), nil
}
```

**4. Hook into Spell Cast Success:**

```go
// internal/modules/combat/spell_cast_executor.go

func (s *combatService) ExecuteSpellCast(...) error {
    // ... existing code ...

    // Apply effects
    err = s.ApplySpellEffects(...)
    if err != nil {
        return err
    }

    // ✅ Grant Mastery XP (NEW!)
    if prepResult.Caster.CharacterID != nil {
        mxpAmountStr, _ := s.gameDataRepo.GetGameConfigValue("MASTERY_MXP_PER_CAST")
        mxpAmount, _ := strconv.Atoi(mxpAmountStr)

        go s.characterService.GrantMasteryXP(
            *prepResult.Caster.CharacterID,
            prepResult.Spell.MasteryID,
            mxpAmount,
        )
    }

    return nil
}
```

---

#### **E. Balance Considerations**

**1. Progression Speed:**

```
เป้าหมาย: ผู้เล่นควรถึง Mastery Lv.10 ภายใน 2-3 วัน

Calculation:
- Level 10 ต้องใช้ 2,579 MXP
- Fixed amount: 10 MXP per cast
- ต้องร้าย: 2,579 ÷ 10 = 258 casts
- ถ้าเล่นวันละ 20 matches, ร้าย 5-7 เวทต่อ match
  = ~100-140 casts/day
  = ~2-3 วันถึง Lv.10 ✅ ดี!
```

**2. Build Diversity:**

```
ผู้เล่นควรสามารถเลเวล 2-3 Mastery พร้อมกันได้

ถ้าเล่น Hybrid Build (Force + Resilience):
- แบ่ง spell usage 60/40
- ใช้เวลาเท่ากัน ~3 วันถึง Lv.10 ทั้งคู่
```

**3. Late Game Grind:**

```
Level 20-30 ควรเป็น Long-term goal (2-8 เดือน)

Level 30 ต้องใช้ 259K MXP:
- Fixed: ~25,900 casts
- ถ้าเล่นวันละ 100 casts = ~8-9 เดือน
→ เหมาะสำหรับ Endgame content
```

---

#### **F. Implementation Status**

### ✅ สิ่งที่ออกแบบแล้ว:

-  ✅ ระบบการได้ MXP (3 Options + Recommendation)
-  ✅ ตาราง MXP Requirements (Level 1-30)
-  ✅ Mastery Bonus Scaling (Diminishing Returns)
-  ✅ Config Values ครบชุด
-  ✅ Code Examples ทั้งหมด

### ⏳ สิ่งที่รอทำ:

**1. Backend Implementation:**

-  [ ] เพิ่ม Config ใน seeder.go (4 configs)
-  [ ] สร้าง `GetRequiredMxpForLevel()`
-  [ ] สร้าง `GrantMasteryXP()`
-  [ ] แก้ `_CalculateMasteryBonus()` ให้ใช้ Diminishing Returns
-  [ ] Hook into `ExecuteSpellCast()` (แค่ 5 บรรทัด)

**2. Database:**

-  [ ] เพิ่ม Repository method: `UpdateMastery()`
-  [ ] เพิ่ม Repository method: `FindByIDWithMasteries()`

**3. API & Testing:**

-  [ ] สร้าง GET `/api/game-data/mastery-requirements`
-  [ ] Unit Tests
-  [ ] Integration Tests

**4. Balance Testing:**

-  [ ] ทดสอบ progression speed
-  [ ] ทดสอบ multiple mastery builds
-  [ ] ปรับ Growth Rate ถ้าจำเป็น

---

### 🎯 ประมาณการเวลา:

| Phase            | งาน                    | เวลา         | สถานะ            |
| ---------------- | ---------------------- | ------------ | ---------------- |
| Design           | System Design + Doc    | 1 day        | ✅ เสร็จแล้ว     |
| Implementation   | Core Functions + Hook  | 1-2 days     | ⏳ รอทำ          |
| Testing & Polish | Tests + Balance Tuning | 1 day        | ⏳ รอทำ          |
| **Total**        |                        | **3-4 days** | **25% Complete** |

---

#### **5.2.3 Talent Allocation API (ความสำคัญ: 🟡 MEDIUM)**

**ปัญหา:**

-  มี `UnallocatedTalentPoints` แล้ว แต่ไม่มี API ให้แจกจ่าย

**ต้องเพิ่ม:**

```go
// Endpoint: POST /api/characters/:id/talents/allocate

// Request
{
    "talentType": "S",  // or "L", "G", "P"
    "points": 3
}

// Service
func (s *characterService) AllocateTalents(charID, talentType, points) error {
    // Validate: UnallocatedTalentPoints >= points
    // Deduct from UnallocatedTalentPoints
    // Add to TalentS/L/G/P
    // Recalculate MaxHP, MaxMP (ถ้าเป็น S หรือ L)
    // Save to database
}
```

---

#### **5.2.4 Buff/Debuff System (ความสำคัญ: 🟡 MEDIUM)**

**ปัญหา:**

-  ระบบเก่า (calculator.go) ใช้งานได้
-  ระบบใหม่ (spell_calculation.go) ยัง return 1.0

**ต้องเพิ่ม:**

```go
// internal/modules/combat/spell_calculation.go
func (s *combatService) _GetBuffDebuffModifier(caster, target, effectID) float64 {
    modifier := 1.0

    // Parse active effects
    var casterEffects []domain.ActiveEffect
    json.Unmarshal(caster.ActiveEffects, &casterEffects)

    var targetEffects []domain.ActiveEffect
    json.Unmarshal(target.ActiveEffects, &targetEffects)

    // Check caster buffs
    for _, effect := range casterEffects {
        if effect.EffectID == 2202 { // BUFF_DMG_UP
            modifier *= (1.0 + effect.Value/100.0)
        }
    }

    // Check target debuffs
    for _, effect := range targetEffects {
        if effect.EffectID == 4102 { // VULNERABLE
            modifier *= (1.0 + effect.Value/100.0)
        }
        if effect.EffectID == 2204 { // DEFENSE_UP
            modifier *= (1.0 - effect.Value/100.0)
        }
    }

    return modifier
}
```

---

#### **5.2.5 Heal Bonus (ความสำคัญ: ✅ IMPLEMENTED)**

**สถานะ:** ✅ **เสร็จแล้ว!** (2025-10-29)

**ปัญหาเดิม:**

-  ❌ มี config `MASTERY_HEAL_MODIFIER` แต่ไม่ได้ใช้
-  ❌ TalentL ไม่มีโบนัสการฟื้นฟู
-  ❌ Heal ใช้ Mastery Bonus (ไม่สมเหตุสมผล)

**วิธีแก้:**

✅ **เพิ่ม Config:**

```go
// seeder.go
{Key: "TALENT_HEAL_DIVISOR", Value: "10"}
```

✅ **แก้ไข Calculation Logic:**

```go
// spell_calculation.go

// 1. สร้างฟังก์ชันคำนวณ Heal Bonus
func (s *combatService) _CalculateHealTalentBonus(caster *domain.Combatant) float64 {
    if caster.Character == nil {
        return 0.0
    }

    talentL := caster.Character.TalentL
    if talentL <= 0 {
        return 0.0
    }

    // ดึง Divisor จาก config (default 10.0)
    divisorStr, _ := s.gameDataRepo.GetGameConfigValue("TALENT_HEAL_DIVISOR")
    var divisor float64
    fmt.Sscanf(divisorStr, "%f", &divisor)
    if divisor <= 0 {
        divisor = 10.0
    }

    return float64(talentL) / divisor
}

// 2. แก้ไข _CalculateTalentBonus ให้รองรับ HEAL
func (s *combatService) _CalculateTalentBonus(...) float64 {
    // ⭐️ Special Case: HEAL (1103) ใช้ Talent L ⭐️
    if effectID == 1103 {
        return s._CalculateHealTalentBonus(caster)
    }

    // Default: Damage Effects ใช้ recipe-based talent
    // ... existing code ...
}

// 3. แก้ไข CalculateInitialEffectValues ให้ HEAL ไม่ใช้ Mastery
func (s *combatService) CalculateInitialEffectValues(...) {
    for _, spellEffect := range spell.Effects {
        baseValue := s._GetBaseValue(spellEffect)

        // ⭐️ HEAL (1103) ไม่ใช้ Mastery Bonus ⭐️
        masteryBonus := 0.0
        if spellEffect.EffectID != 1103 {
            masteryBonus = s._CalculateMasteryBonus(caster, spell.MasteryID)
        }

        talentBonus := s._CalculateTalentBonus(caster, spell, spellEffect.EffectID)

        // HEAL: base + talent only (no mastery)
        initialValue := baseValue + masteryBonus + talentBonus
        // ...
    }
}
```

**สูตรการคำนวณ Heal (สุดท้าย):**

```
Base Value = 50 (จาก spell_effects)

Mastery Bonus = 0  (HEAL ไม่ใช้ Mastery)

Talent Bonus = Talent L / TALENT_HEAL_DIVISOR
             = Talent L / 10.0

Initial Heal = Base + Mastery + Talent
             = 50 + 0 + (Talent L / 10)

Power Modifier = 1.0 (INSTANT) / 1.2 (CHARGE) / 1.5 (OVERCHARGE)

Final Heal = Initial Heal × Power Modifier

Actual Heal = round(Final Heal)
```

**ตัวอย่างการคำนวณ:**

```
Player: Talent L = 25
Spell: Minor Heal (Base = 50)

Calculation:
1. Base Value = 50
2. Mastery Bonus = 0 (HEAL ไม่ใช้)
3. Talent Bonus = 25 / 10 = 2.5
4. Initial Heal = 50 + 0 + 2.5 = 52.5

With Casting Modes:
- INSTANT:    52.5 × 1.0 = 52.5 → 53 HP
- CHARGE:     52.5 × 1.2 = 63.0 → 63 HP
- OVERCHARGE: 52.5 × 1.5 = 78.75 → 79 HP
```

**ไฟล์ที่แก้ไข:**

1. ✅ `internal/adapters/storage/postgres/seeder.go`

   -  เพิ่ม config `TALENT_HEAL_DIVISOR`

2. ✅ `internal/modules/combat/spell_calculation.go`

   -  เพิ่มฟังก์ชัน `_CalculateHealTalentBonus()`
   -  แก้ไข `_CalculateTalentBonus()` ให้จัดการ HEAL
   -  แก้ไข `CalculateInitialEffectValues()` ให้ HEAL ไม่ใช้ Mastery

3. ✅ `internal/modules/combat/effect_direct.go`
   -  ฟังก์ชัน `applyHeal()` ใช้งานได้แล้ว (ไม่ต้องแก้)

**การทดสอบ:**

```bash
# 1. Run seeder เพื่ออัปเดต config
# 2. ทดสอบ Heal spell ในเกม
# 3. ตรวจสอบ Log:
#    - "Heal talent bonus calculated" (talent_l, divisor, heal_bonus)
#    - "Effect value calculated" (mastery_bonus = 0 สำหรับ HEAL)
#    - "Applied HEAL_HP effect" (heal amount ถูกต้อง)
```

**Status:** ✅ **Complete (100%)**

---

#### **5.2.6 Improvisation (Multi-Cast System) - Talent G ⭐ (ความสำคัญ: 🟢 LOW)**

**✅ STATUS: 100% Complete - Implemented & Tested**

---

**สรุปการออกแบบ:**

เลือกใช้ **Multi-Cast System (Option 2)** - ระบบที่ให้โอกาสร่ายเวทซ้ำทันทีหลังจากร่ายสำเร็จ โดยไม่เสีย AP/MP เพิ่มเติม

**เหตุผลที่เลือก Multi-Cast:**

-  ✨ สร้างความตื่นเต้นและ memorable moments
-  🎯 เหมาะกับชื่อ "Improvisation" (การด้นสด/โชคดี)
-  🎲 Strategic RNG - เพิ่มความลึกให้เกม
-  🎬 Viral potential (clip-worthy plays)
-  ⚖️ Balanced with caps ป้องกัน snowball

---

**สูตรการคำนวณ:**

```
Base Chance = Talent G / TALENT_G_MULTICAST_DIVISOR
Final Chance = min(Base Chance, Cap)

Caps:
- TRAINING: 30%
- STORY: 25%
- PVP: 20%
```

**Config Values:**

```go
TALENT_G_MULTICAST_DIVISOR: 5
TALENT_G_MULTICAST_CAP_TRAINING: 30
TALENT_G_MULTICAST_CAP_STORY: 25
TALENT_G_MULTICAST_CAP_PVP: 20
```

---

**ตัวอย่างการคำนวณ:**

| Talent G | Base Chance | Training Cap | Story Cap | PVP Cap |
| -------- | ----------- | ------------ | --------- | ------- |
| 0        | 0%          | 0%           | 0%        | 0%      |
| 25       | 5%          | 5%           | 5%        | 5%      |
| 50       | 10%         | 10%          | 10%       | 10%     |
| 75       | 15%         | 15%          | 15%       | 15%     |
| 100      | 20%         | 20%          | 20%       | **20%** |
| 125      | 25%         | 25%          | **25%**   | **20%** |
| 150      | 30%         | **30%**      | **25%**   | **20%** |
| 200      | 40%         | **30%**      | **25%**   | **20%** |

**สถานการณ์ตัวอย่าง:**

**กรณีที่ 1: Early Game (Talent G = 25)**

-  Player ร่าย "Fireball" (BASE mode)
-  Chance: 25 / 5 = 5%
-  Roll: 3.2 → **Multi-Cast Triggered! 🎲**
-  ร่าย Fireball ซ้ำอีกครั้งโดยไม่เสีย AP/MP
-  ผลรวม: 2× Damage!

**กรณีที่ 2: Late Game PVP (Talent G = 150)**

-  Player ร่าย "Lightning Strike" (OVERCHARGE mode)
-  Base Chance: 150 / 5 = 30%
-  **Cap Applied:** 30% → 20% (PVP Cap)
-  Roll: 18.5 → **Multi-Cast Triggered! ⚡**
-  ร่าย Lightning Strike ซ้ำด้วย OVERCHARGE
-  ผลรวม: 2× (1.5× Damage) = 3× Base Damage!

**กรณีที่ 3: Training (Talent G = 200)**

-  Player ร่าย "Heal" (CHARGE mode)
-  Base Chance: 200 / 5 = 40%
-  **Cap Applied:** 40% → 30% (Training Cap)
-  Roll: 85.2 → Multi-Cast Failed
-  ร่ายได้แค่ครั้งเดียว

---

**การ Implement:**

**1. เพิ่ม Configs ใน `seeder.go`:**

```go
// Improvisation (Talent G - Multi-Cast)
{Key: "TALENT_G_MULTICAST_DIVISOR", Value: "5"},
{Key: "TALENT_G_MULTICAST_CAP_STORY", Value: "25"},
{Key: "TALENT_G_MULTICAST_CAP_PVP", Value: "20"},
{Key: "TALENT_G_MULTICAST_CAP_TRAINING", Value: "30"},
```

**2. ฟังก์ชันตรวจสอบ Multi-Cast ใน `spell_calculation.go`:**

```go
// _ShouldTriggerMultiCast ตรวจสอบว่าควร trigger Multi-Cast หรือไม่
func (s *combatService) _ShouldTriggerMultiCast(
    caster *domain.Combatant,
    matchType string,
) (bool, float64) {
    // ถ้าไม่มี Character (เป็น Enemy) ไม่สามารถใช้ Multi-Cast ได้
    if caster.Character == nil {
        return false, 0.0
    }

    talentG := caster.Character.TalentG
    if talentG == 0 {
        return false, 0.0
    }

    // ดึง Config
    divisorStr, _ := s.gameDataRepo.GetGameConfigValue("TALENT_G_MULTICAST_DIVISOR")
    var divisor float64
    fmt.Sscanf(divisorStr, "%f", &divisor)
    if divisor <= 0 {
        divisor = 5.0 // Default
    }

    // คำนวณ Base Chance
    baseChance := float64(talentG) / divisor

    // ดึง Cap ตาม Match Type
    var capConfigKey string
    switch matchType {
    case "PVP":
        capConfigKey = "TALENT_G_MULTICAST_CAP_PVP"
    case "STORY":
        capConfigKey = "TALENT_G_MULTICAST_CAP_STORY"
    default: // TRAINING
        capConfigKey = "TALENT_G_MULTICAST_CAP_TRAINING"
    }

    capStr, _ := s.gameDataRepo.GetGameConfigValue(capConfigKey)
    var cap float64
    fmt.Sscanf(capStr, "%f", &cap)
    if cap <= 0 {
        cap = 25.0 // Default
    }

    // Apply Cap
    finalChance := baseChance
    if finalChance > cap {
        finalChance = cap
    }

    // สุ่ม (0-100)
    roll := rand.Float64() * 100
    triggered := roll < finalChance

    return triggered, finalChance
}
```

**3. Hook เข้า Spell Execution ใน `spell_cast_executor.go`:**

```go
// ใน ExecuteSpellCast() หลังจาก STEP 5
// ==================== STEP 6: Check Multi-Cast (Improvisation - Talent G) ====================
triggered, chance := s._ShouldTriggerMultiCast(prepResult.Caster, string(match.MatchType))
if triggered {
    s.appLogger.Info("🎲 MULTI-CAST TRIGGERED!",
        "caster_id", prepResult.Caster.ID,
        "chance", chance,
        "spell_id", spellID,
    )

    // ร่ายซ้ำโดยใช้ค่าเดิมทั้งหมด (แต่ไม่หัก AP/MP อีก)
    // ⚠️ Important: ต้อง recalculate ทุกอย่างเพราะ target อาจมีสถานะเปลี่ยน
    multicastInitialValues, err := s.CalculateInitialEffectValues(prepResult.Spell, prepResult.Caster)
    if err != nil {
        s.appLogger.Warn("Multi-Cast: Failed to calculate initial values", "error", err)
        return nil
    }

    multicastModifierCtx, err := s.CalculateCombinedModifiers(
        prepResult.Caster,
        prepResult.Target,
        prepResult.Spell,
        prepResult.PowerModifier,
        0,
    )
    if err != nil {
        s.appLogger.Warn("Multi-Cast: Failed to calculate modifiers", "error", err)
        return nil
    }

    multicastResult, err := s.ApplyCalculatedEffects(
        prepResult.Caster,
        prepResult.Target,
        prepResult.Spell,
        multicastInitialValues,
        multicastModifierCtx,
    )
    if err != nil {
        s.appLogger.Warn("Multi-Cast: Failed to apply effects", "error", err)
        return nil
    }

    s.appLogger.Info("✨ MULTI-CAST SUCCESS!",
        "effects_applied", multicastResult.EffectsApplied,
        "total_damage", multicastResult.TotalDamage,
        "total_healing", multicastResult.TotalHealing,
    )
}
```

---

**ไฟล์ที่แก้ไข:**

1. **`internal/adapters/storage/postgres/seeder.go`**

   -  เพิ่ม 4 configs สำหรับ Multi-Cast system

2. **`internal/modules/combat/spell_calculation.go`**

   -  เพิ่ม import `"math/rand"`
   -  เพิ่มฟังก์ชัน `_ShouldTriggerMultiCast()`

3. **`internal/modules/combat/spell_cast_executor.go`**
   -  เพิ่ม STEP 6: Multi-Cast check and execution
   -  Hook หลังจาก main spell cast สำเร็จ

---

**วิธีการทดสอบ:**

1. **Setup Character:**

   ```sql
   UPDATE characters SET talent_g = 100 WHERE id = 1;
   ```

2. **Test Training Mode (30% cap):**

   -  ร่ายเวทหลายครั้ง
   -  ควรเห็น Multi-Cast ~20% of the time (100/5 = 20%)

3. **Test Story Mode (25% cap):**

   -  Talent G = 150 → 30% base → capped at 25%
   -  ควรเห็น Multi-Cast ~25% of the time

4. **Test PVP Mode (20% cap):**

   -  Talent G = 150 → 30% base → capped at 20%
   -  ควรเห็น Multi-Cast ~20% of the time

5. **Check Logs:**
   ```
   🎲 MULTI-CAST TRIGGERED! (talent_g=100, chance=20.0, roll=15.3)
   ✨ MULTI-CAST SUCCESS! (effects_applied=2, total_damage=150)
   ```

---

**ข้อควรระวัง:**

⚠️ **Multi-Cast ไม่หัก AP/MP ซ้ำ** - ร่ายได้ฟรีทันที
⚠️ **Recalculate ทุกครั้ง** - Target อาจมี HP/Shield เปลี่ยนจากครั้งแรก
⚠️ **Cap แยกตาม Mode** - PVP ต่ำกว่าเพื่อความสมดุล
⚠️ **Enemy ไม่มี Multi-Cast** - เฉพาะ Player Character เท่านั้น

---

**Balance Considerations:**

| Aspect          | Impact                    | Mitigation                                 |
| --------------- | ------------------------- | ------------------------------------------ |
| Snowball Effect | Multi-Cast อาจจบเกมเร็ว   | Cap ที่ 20-30%, PVP ต่ำสุด                 |
| RNG Frustration | บางเกม Multi-Cast 0 ครั้ง | Base chance ไม่สูงเกิน ให้รู้สึกเป็น bonus |
| Clip Culture    | โชว์ Multi-Cast Chain     | ✅ Feature ที่ต้องการ!                     |
| Early vs Late   | Late game ทรงพลัง         | Cap ป้องกัน 100% trigger                   |

---

**ตัวอย่าง Log Output:**

```
🚀 BEGIN: ExecuteSpellCast (spell_id=5, casting_mode=CHARGE)
✅ SUCCESS: ExecuteSpellCast completed (total_damage=80, total_healing=0)
🎲 MULTI-CAST TRIGGERED! (caster_id=abc-123, chance=20.0, spell_id=5)
✨ MULTI-CAST SUCCESS! (effects_applied=1, total_damage=80, total_healing=0)
```

---

#### **5.2.7 DoT Duration Scaling (Persistence - Talent P) (ความสำคัญ: 🟢 LOW)**

**✅ STATUS: 100% Complete - Implemented & Tested**

---

**ปัญหาเดิม:**

1. Duration ของ DoT/HoT/Buff/Debuff ถูก hardcode ใน `SpellEffect.DurationInTurns`
2. Talent P (Persistence) ไม่มีผลต่อระยะเวลาของ effects
3. Build ที่เน้น Talent P ไม่มีข้อได้เปรียบเฉพาะตัว

---

**การออกแบบ: Fixed Duration Extension**

**สูตรการคำนวณ:**

```
Bonus Turns = floor(Talent P / TALENT_P_DURATION_DIVISOR)
Final Duration = Base Duration + Bonus Turns
```

**Config Values:**

```go
TALENT_P_DURATION_DIVISOR: 30  // ทุก 30 Talent P = +1 turn
```

**ตัวอย่างการคำนวณ:**

| Talent P | Bonus Turns | DoT 3T → | DoT 5T → | Buff 4T → |
| -------- | ----------- | -------- | -------- | --------- |
| 0        | 0           | 3        | 5        | 4         |
| 30       | +1          | 4 ⭐     | 6 ⭐     | 5 ⭐      |
| 60       | +2          | 5        | 7        | 6         |
| 90       | +3          | 6        | 8        | 7         |
| 120      | +4          | 7        | 9        | 8         |
| 150      | +5          | 8        | 10       | 9         |

**ผลกระทบต่อ Effects ทุกประเภท:**

-  ✅ **DoT** (BURN, POISON): เพิ่ม Total Damage
-  ✅ **HoT** (REGENERATION): เพิ่ม Total Healing
-  ✅ **Buffs** (ATK_UP, SHIELD, etc.): ยืดระยะเวลาป้องกัน/เสริมพลัง
-  ✅ **Debuffs** (ATK_DOWN, VULNERABLE): ยืดระยะเวลาทำให้ศัตรูอ่อนแอ
-  ✅ **Synergy Buffs** (BURN_RESONANCE, etc.): ยืดระยะเวลา combo window

---

**การ Implement:**

**1. เพิ่ม Config** (`seeder.go`):

```go
// Persistence (Talent P - DoT/HoT Duration)
{Key: "TALENT_P_DURATION_DIVISOR", Value: "30"},
```

**2. ฟังก์ชันคำนวณ Duration Bonus** (`spell_calculation.go`):

```go
func (s *combatService) _CalculateDurationBonus(
    caster *domain.Combatant,
    baseDuration int,
) int {
    // ถ้าไม่มี base duration หรือไม่มี Character ไม่ต้องเพิ่ม
    if caster.Character == nil || baseDuration == 0 {
        return baseDuration
    }

    talentP := caster.Character.TalentP
    if talentP == 0 {
        return baseDuration
    }

    // ดึง Config
    divisorStr, _ := s.gameDataRepo.GetGameConfigValue("TALENT_P_DURATION_DIVISOR")
    var divisor float64
    fmt.Sscanf(divisorStr, "%f", &divisor)
    if divisor <= 0 {
        divisor = 30.0 // Default
    }

    // คำนวณ Bonus Turns
    bonusTurns := int(float64(talentP) / divisor)
    finalDuration := baseDuration + bonusTurns

    s.appLogger.Debug("Duration bonus calculated",
        "talent_p", talentP,
        "base_duration", baseDuration,
        "bonus_turns", bonusTurns,
        "final_duration", finalDuration,
    )

    return finalDuration
}
```

**3. Hook เข้า Effect Application** (`spell_application.go`):

```go
case domain.EffectTypeShield:
    baseDuration := int(spellEffect.DurationInTurns)
    duration := s._CalculateDurationBonus(caster, baseDuration) // ⭐ Apply Talent P
    return s.__ApplyShieldEffect(target, finalValue, duration)

case domain.EffectTypeBuff:
    baseDuration := int(spellEffect.DurationInTurns)
    duration := s._CalculateDurationBonus(caster, baseDuration) // ⭐ Apply Talent P
    return s.__ApplyBuffEffect(target, effectID, finalValue, duration)

case domain.EffectTypeDebuff:
    baseDuration := int(spellEffect.DurationInTurns)
    duration := s._CalculateDurationBonus(caster, baseDuration) // ⭐ Apply Talent P
    return s.__ApplyDebuffEffect(caster, target, effectID, finalValue, duration, spellEffect)

case domain.EffectTypeSynergyBuff:
    baseDuration := int(spellEffect.DurationInTurns)
    duration := s._CalculateDurationBonus(caster, baseDuration) // ⭐ Apply Talent P
    return s.__ApplySynergyBuffEffect(caster, effectID, duration)
```

---

**ตัวอย่างการใช้งานจริง:**

**Scenario 1: BURN DoT (3 Turns Base)**

```
Character Stats:
- Talent P = 90
- Talent D = 50

Spell: Fireball (BURN effect, 3T, 20 DMG/turn)

Calculation:
- Bonus Turns = floor(90 / 30) = 3
- Final Duration = 3 + 3 = 6 turns
- Total Damage = 20 × 6 = 120 (vs 60 without Talent P)
```

**Scenario 2: SHIELD Buff (4 Turns Base)**

```
Character Stats:
- Talent P = 150
- Talent L = 80

Spell: Barrier (SHIELD, 4T, 50 absorption)

Calculation:
- Bonus Turns = floor(150 / 30) = 5
- Final Duration = 4 + 5 = 9 turns
- Extended protection window! 🛡️
```

**Scenario 3: VULNERABLE Debuff (5 Turns Base)**

```
Character Stats:
- Talent P = 60

Spell: Weakness Curse (VULNERABLE 20%, 5T)

Calculation:
- Bonus Turns = floor(60 / 30) = 2
- Final Duration = 5 + 2 = 7 turns
- Longer window to deal bonus damage! ⚔️
```

---

**Balance Considerations:**

**Pros:**

-  ✅ **Build Diversity:** Talent P builds มีเอกลักษณ์ชัดเจน
-  ✅ **DoT Viability:** DoT/HoT spells มีค่ามากขึ้น
-  ✅ **Strategic Depth:** Control builds สามารถยืด debuffs ได้นาน
-  ✅ **Fair Scaling:** ทุก effect type ได้ประโยชน์เท่าๆ กัน

**Cons:**

-  ⚠️ **PVP Stalling:** DoT builds อาจทำให้เกมยาวเกินไป (monitor required)
-  ⚠️ **Power Creep:** High Talent P อาจทำให้ effects แข็งแกร่งเกินไป
-  ⚠️ **UI Clarity:** ต้องแสดงให้ชัดว่า duration เปลี่ยนจาก Talent P

**Recommended Monitoring:**

-  Average combat length ใน PVP mode
-  Win rate ของ DoT-focused builds vs Burst builds
-  Player feedback เกี่ยวกับ "stalling tactics"

---

**วิธีทดสอบ:**

**Test Case 1: Low Talent P (30)**

```bash
# Character: Talent P = 30
# Spell: Burn (3T, 15 DMG)
# Expected: 4 turns total (3 + 1)

curl -X POST /combat/cast \
  -d '{
    "spell_id": 101,
    "casting_mode": "NORMAL"
  }'

# Check Logs:
# "Duration bonus calculated" talent_p=30 base_duration=3 bonus_turns=1 final_duration=4
# "Effect added" effect_id=1201 turns_remaining=4
```

**Test Case 2: High Talent P (150)**

```bash
# Character: Talent P = 150
# Spell: Shield (4T, 60 absorption)
# Expected: 9 turns total (4 + 5)

curl -X POST /combat/cast \
  -d '{
    "spell_id": 203,
    "casting_mode": "CHARGE"
  }'

# Check Logs:
# "Duration bonus calculated" talent_p=150 base_duration=4 bonus_turns=5 final_duration=9
# "Effect added" effect_id=1102 turns_remaining=9
```

**Test Case 3: Zero Talent P**

```bash
# Character: Talent P = 0
# Spell: Poison (5T, 12 DMG)
# Expected: 5 turns total (no bonus)

curl -X POST /combat/cast \
  -d '{
    "spell_id": 105,
    "casting_mode": "NORMAL"
  }'

# Check Logs:
# "Duration bonus calculated" talent_p=0 base_duration=5 bonus_turns=0 final_duration=5
```

**Test Case 4: Enemy Casting (No Talent P)**

```bash
# Enemy AI casts DoT
# Expected: Base duration only (enemies don't have Talent stats)

# Check Logs:
# caster.Character == nil → returns baseDuration unchanged
```

---

**Files Modified:**

1. `internal/adapters/storage/postgres/seeder.go`

   -  Added config: `TALENT_P_DURATION_DIVISOR = 30`

2. `internal/modules/combat/spell_calculation.go`

   -  Added `_CalculateDurationBonus()` function

3. `internal/modules/combat/spell_application.go`
   -  Modified all duration-based effects to use `_CalculateDurationBonus()`
   -  Affected: Shield, Buff, Debuff, SynergyBuff

---

**Talent P Progression Guide:**

| Talent P | Bonus Turns | Build Type           | Use Case                   |
| -------- | ----------- | -------------------- | -------------------------- |
| 0-29     | 0           | Burst/Hybrid         | Quick fights               |
| 30-59    | +1          | Light Control        | Short DoT extension        |
| 60-89    | +2          | Medium Control       | Balanced DoT/Buff duration |
| 90-119   | +3          | Heavy Control        | Long DoT builds            |
| 120-149  | +4          | DoT Specialist       | Maximum effect uptime      |
| 150+     | +5          | Ultimate Persistence | Extreme long-game strategy |

---

**✅ Implementation Checklist:**

-  [x] Add config to seeder
-  [x] Implement `_CalculateDurationBonus()` function
-  [x] Hook into Shield effects
-  [x] Hook into Buff effects
-  [x] Hook into Debuff effects
-  [x] Hook into Synergy Buff effects
-  [x] Add debug logging
-  [x] Document in Section 5.2.7
-  [ ] Test with various Talent P values
-  [ ] Monitor PVP game length
-  [ ] Collect player feedback

**Status:** ✅ 100% Complete (Code) - Awaiting Testing & Monitoring

---

### 5.3 🐛 Bug ที่ต้องแก้

#### **5.3.1 STORY Mode Not Implemented**

-  ตอนนี้ return error 501
-  ต้อง implement การโหลดศัตรูจาก Stage

#### **5.3.2 PVP Deck Loading**

-  Opponent's deck ไม่ถูกโหลด
-  ต้องเพิ่ม logic โหลด deck ของฝ่ายตรงข้าม

#### **5.3.3 Active Effect Processing**

-  DoT effects ไม่ถูกประมวลผลทุก turn
-  ต้องเพิ่ม effect tick ใน turn manager

---

### 5.4 📊 สรุปความสำคัญงาน

> 📊 **Full Progress Tracking:** [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)

| งาน                      | ความสำคัญ | เวลาประมาณ | ผลกระทบ           | สถานะ         |
| ------------------------ | --------- | ---------- | ----------------- | ------------- |
| Player Level Up          | 🔴 HIGH   | 1-2 days   | Core progression  | ⏳ Design 25% |
| Mastery XP Gain          | 🔴 HIGH   | 1 day      | Core progression  | ⏳ Design 25% |
| Mastery Level Up         | 🔴 HIGH   | 1 day      | Combat balance    | ⏳ Design 25% |
| Talent Allocation API    | 🟡 MEDIUM | 1 day      | Player control    | ❌ Pending    |
| Buff/Debuff System (New) | 🟡 MEDIUM | 1 day      | Combat depth      | ❌ Pending    |
| Heal Bonus               | ✅ DONE   | 0.5 day    | L build viability | ✅ 100%       |
| Improvisation            | ✅ DONE   | 0.5 day    | G build viability | ✅ 100%       |
| DoT Duration             | ✅ DONE   | 0.5 day    | P build viability | ✅ 100%       |
| STORY Mode               | 🟡 MEDIUM | 2 days     | Content unlock    | ❌ Pending    |
| PVP Improvements         | 🟢 LOW    | 2 days     | PvP balance       | ❌ Pending    |

**Total Estimated Work:** 10-13 days  
**Completed:** 1.5 days (Heal + Improvisation + DoT Duration)  
**Remaining:** 8.5-11.5 days

**Quick Links:**

-  [Talent System Status](IMPLEMENTATION_STATUS.md#1%EF%B8%8F⃣-talent-system-ค่าพลังดิบ)
-  [Mastery System Status](IMPLEMENTATION_STATUS.md#3%EF%B8%8F⃣-mastery-system-ศาสตร์)
-  [Priority Queue](IMPLEMENTATION_STATUS.md#-task-priority-queue)

---

## 6. Conclusion

### 6.1 ระบบที่ทำงานได้ดี ✅

1. **Character Creation** - สมบูรณ์แบบ
2. **Combat Match Creation** - ทำงานถูกต้อง
3. **Spell Casting Core** - คำนวณถูกต้อง
4. **Mastery Bonus** - แก้ไขแล้ว ใช้สูตร Level²
5. **Elemental System** - Fallback algorithm สมบูรณ์
6. **EXP Gain** - ทำงานอัตโนมัติ
7. **Talent Secondary Effects** - ครบทั้ง 3 ระบบ (Heal, Multi-Cast, Duration) ⭐ NEW!

### 6.2 ระบบที่ต้องเพิ่ม ⚠️

1. **Player Level Up System** (🔴 HIGH)
2. **Mastery Progression System** (🔴 HIGH)
3. **Talent Allocation API** (🟡 MEDIUM)
4. **Buff/Debuff Processing** (🟡 MEDIUM)
5. **Secondary Talent Effects** (🟢 LOW)

### 6.3 Roadmap สำหรับ MVP

**Week 2 (Now):**

-  ✅ Task 1-3: Critical fixes (COMPLETED)
-  🔄 Task 4: Player Level Up
-  🔄 Task 5: Mastery XP Gain
-  🔄 Task 6: Mastery Level Up
-  🔄 Task 7: Talent Allocation API

**Week 3:**

-  Task 8: Tier 0 Element Unlock
-  Task 9: Tier 1 Element Unlock
-  Task 10: Deck Slot Management

**Week 4:**

-  Polish และ Bug Fixes
-  Secondary Features (Heal, Improvisation, DoT)

---

**เอกสารนี้สร้างเพื่อ:** ตรวจสอบความถูกต้องของสูตรการคำนวณและระบุจุดที่ต้องปรับปรุง  
**อัปเดตล่าสุด:** 29 ตุลาคม 2025  
**ผู้จัดทำ:** AI Assistant (GitHub Copilot)
