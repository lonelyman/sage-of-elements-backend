# 🎮 User Journey Analysis: Account → Character → Match → Victory

**วิเคราะห์เส้นทางผู้ใช้งานตั้งแต่เริ่มต้นจนจบ Match**  
**Date:** November 1, 2025  
**Analyzed By:** GitHub Copilot + nipon.k

---

## 📊 สรุปภาพรวม

| Journey Step             | APIs Required | Completion Status | Priority  |
| ------------------------ | ------------- | ----------------- | --------- |
| 1️⃣ Account Creation      | 1 API         | ✅ 100% Complete  | -         |
| 2️⃣ Character Creation    | 1 API         | ✅ 100% Complete  | -         |
| 3️⃣ Deck Setup            | 2 APIs        | ✅ 100% Complete  | -         |
| 4️⃣ Match Creation        | 1 API         | ⚠️ 80% Complete   | 🟡 MEDIUM |
| 5️⃣ Combat Actions        | 2 APIs        | ✅ 95% Complete   | 🟢 LOW    |
| 6️⃣ Match Victory         | Auto          | ⚠️ 70% Complete   | 🔴 HIGH   |
| 7️⃣ Rewards & Progression | Auto          | ⏳ 50% Complete   | 🔴 HIGH   |

**Overall Progress:** 🟢 85% Complete (โครงสร้างหลักพร้อมใช้งาน แต่ยังมีฟีเจอร์ที่ต้องเติม)

---

## 🛣️ ขั้นตอนละเอียด

### **1️⃣ Account Creation (สร้างบัญชี)**

#### 📡 API Endpoint

```
POST /api/v1/players/register
```

#### 📥 Request Body

```json
{
   "username": "player123",
   "email": "player@example.com",
   "password": "securepass123"
}
```

#### 📤 Response

```json
{
   "success": true,
   "message": "Registration successful",
   "data": {
      "id": 1,
      "username": "player123",
      "email": "player@example.com",
      "created_at": "2025-11-01T10:00:00Z"
   }
}
```

#### ✅ Status: **COMPLETE (100%)**

-  ✅ Input validation (username min 4, password min 8, email format)
-  ✅ Duplicate username check
-  ✅ Password hashing (bcrypt)
-  ✅ Database storage
-  ✅ Error handling

#### 📂 Files Involved

-  Handler: `internal/modules/player/handler.go` (L73)
-  Service: `internal/modules/player/service.go`
-  Repository: `internal/adapters/storage/postgres/player_repository.go`

---

### **2️⃣ Character Creation (สร้างตัวละคร)**

#### 📡 API Endpoint

```
POST /api/v1/characters/
Authorization: Bearer <access_token>
```

#### 📥 Request Body

```json
{
   "name": "FireMage",
   "gender": "MALE",
   "elementId": 1, // 1=S (Solidity), 2=L (Liquidity), 3=G (Gas), 4=P (Plasma)
   "masteryId": 1 // 1=Creation, 2=Destruction, 3=Restoration, 4=Transmutation
}
```

#### 📤 Response

```json
{
   "success": true,
   "message": "Character created successfully",
   "data": {
      "id": 1,
      "player_id": 1,
      "character_name": "FireMage",
      "gender": "MALE",
      "primary_element_id": 1,
      "level": 1,
      "exp": 0,
      "talent_s": 93, // Base(3) + Primary Element Bonus(+90)
      "talent_l": 3,
      "talent_g": 3,
      "talent_p": 3,
      "current_hp": 1023, // STAT_HP_BASE(100) + (TalentS × STAT_HP_PER_TALENT_S(10))
      "current_mp": 330, // STAT_MP_BASE(100) + (TalentL × STAT_MP_PER_TALENT_L(25))
      "masteries": [
         { "mastery_id": 1, "level": 1, "mxp": 0 },
         { "mastery_id": 2, "level": 1, "mxp": 0 },
         { "mastery_id": 3, "level": 1, "mxp": 0 },
         { "mastery_id": 4, "level": 1, "mxp": 0 }
      ],
      "created_at": "2025-11-01T10:05:00Z"
   }
}
```

#### ✅ Status: **COMPLETE (100%)**

-  ✅ Name validation (min 3 characters)
-  ✅ Duplicate name check
-  ✅ Talent calculation (Base + Primary Element Bonus)
-  ✅ HP/MP calculation based on talents
-  ✅ Mastery initialization (all start at level 1)
-  ✅ Database storage with relationships
-  ✅ Gender is cosmetic only (no stat bonuses since Oct 28, 2025)

#### 🧮 Calculation Details

```
TalentS = 3 + (elementID == 1 ? 90 : 0)
TalentL = 3 + (elementID == 2 ? 90 : 0)
TalentG = 3 + (elementID == 3 ? 90 : 0)
TalentP = 3 + (elementID == 4 ? 90 : 0)

MaxHP = STAT_HP_BASE(100) + (TalentS × STAT_HP_PER_TALENT_S(10))
MaxMP = STAT_MP_BASE(100) + (TalentL × STAT_MP_PER_TALENT_L(25))
CurrentHP = MaxHP
CurrentMP = MaxMP
```

#### 📂 Files Involved

-  Handler: `internal/modules/character/handler.go` (L38)
-  Service: `internal/modules/character/service.go` (L49-124)
-  Repository: `internal/adapters/storage/postgres/character_repository.go`
-  Domain: `internal/domain/character.go`

---

### **3️⃣ Deck Setup (จัดเตรียมสำรับการ์ด)**

#### 📡 API Endpoints

**3.1 Create Deck (สร้างสำรับใหม่)**

```
POST /api/v1/decks/
Authorization: Bearer <access_token>
```

**Request Body:**

```json
{
   "character_id": 1,
   "deck_name": "My First Deck",
   "slots": [
      { "slot_num": 1, "element_id": 5 }, // T1 Basic Element
      { "slot_num": 2, "element_id": 6 },
      { "slot_num": 3, "element_id": 7 },
      { "slot_num": 4, "element_id": 8 },
      { "slot_num": 5, "element_id": 9 },
      { "slot_num": 6, "element_id": 10 },
      { "slot_num": 7, "element_id": 11 },
      { "slot_num": 8, "element_id": 12 }
   ]
}
```

**3.2 Get Decks (ดึงรายการสำรับทั้งหมด)**

```
GET /api/v1/decks/?character_id=1
Authorization: Bearer <access_token>
```

#### ✅ Status: **COMPLETE (100%)**

-  ✅ Deck name validation
-  ✅ Slot validation (1-8 slots, element_id >= 5)
-  ✅ Character ownership check
-  ✅ Create, Read, Update, Delete operations
-  ✅ Multiple decks per character support

#### 📂 Files Involved

-  Handler: `internal/modules/deck/handler.go` (L47-50)
-  Service: `internal/modules/deck/service.go`
-  Repository: `internal/adapters/storage/postgres/deck_repository.go`

---

### **4️⃣ Match Creation (สร้างห้องต่อสู้)**

#### 📡 API Endpoint

```
POST /api/v1/combat/
Authorization: Bearer <access_token>
```

#### 📥 Request Body

**4.1 TRAINING Mode (โหมดฝึกซ้อม)**

```json
{
   "character_id": 1,
   "match_type": "TRAINING",
   "deck_id": 1,
   "training_enemies": [{ "enemy_id": 1 }]
}
```

**4.2 STORY Mode (โหมดเนื้อเรื่อง) - ⚠️ Not Fully Implemented**

```json
{
   "character_id": 1,
   "match_type": "STORY",
   "deck_id": 1,
   "stage_id": 1
}
```

**4.3 PVP Mode (ต่อสู้ผู้เล่น) - ⚠️ Not Fully Implemented**

```json
{
   "character_id": 1,
   "match_type": "PVP",
   "deck_id": 1,
   "opponent_id": 2
}
```

#### 📤 Response

```json
{
   "success": true,
   "message": "Match created successfully",
   "data": {
      "id": "01932f5d-8e9f-7890-abcd-ef1234567890",
      "match_type": "TRAINING",
      "status": "IN_PROGRESS",
      "current_turn": 1,
      "current_phase": "START",
      "active_combatant_id": "01932f5d-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      "combatants": [
         {
            "id": "01932f5d-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
            "character_id": 1,
            "initiative": 330, // STAT_INITIATIVE_BASE(300) + (TalentG × STAT_INITIATIVE_PER_TALENT_G(10))
            "current_hp": 1023,
            "current_mp": 330,
            "current_ap": 0,
            "deck": [
               { "element_id": 5, "is_consumed": false },
               { "element_id": 6, "is_consumed": false }
               // ... 8 slots total
            ]
         },
         {
            "id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
            "enemy_id": 1,
            "initiative": 280,
            "current_hp": 800,
            "current_mp": 9999,
            "current_ap": 0
         }
      ],
      "created_at": "2025-11-01T10:15:00Z"
   }
}
```

#### ⚠️ Status: **PARTIAL (80%)**

**✅ Implemented:**

-  ✅ TRAINING mode (เลือกศัตรูเองได้)
-  ✅ Character ownership validation
-  ✅ Active match check (ป้องกันมีหลาย match พร้อมกัน)
-  ✅ Deck loading and validation
-  ✅ Player combatant initialization
-  ✅ Enemy combatant initialization
-  ✅ Initiative calculation
-  ✅ Turn order sorting
-  ✅ Database storage

**⏳ Not Implemented:**

-  ❌ STORY mode (ต้องโหลดศัตรูจากด่าน)
-  ❌ PVP mode (ต้องสร้าง combatant ของผู้เล่นอีกคน)
-  ❌ Match modifiers system (buffs/debuffs ที่มีผลตลอด match)

#### 📂 Files Involved

-  Handler: `internal/modules/combat/handler.go` (L81, L87-113)
-  Service: `internal/modules/combat/service.go` (L61-269)
-  Repository: `internal/adapters/storage/postgres/combat_repository.go`

---

### **5️⃣ Combat Actions (การกระทำในการต่อสู้)**

#### 📡 API Endpoints

**5.1 Perform Action (ทำการกระทำ)**

```
POST /api/v1/combat/:match_id/actions
Authorization: Bearer <access_token>
```

**Request Body Options:**

**A. End Turn (จบเทิร์น)**

```json
{
   "action_type": "END_TURN"
}
```

**B. Cast Spell (ร้ายเวท)**

```json
{
   "action_type": "CAST_SPELL",
   "cast_mode": "INSTANT", // INSTANT, CHARGE, OVERCHARGE
   "spell_id": 101,
   "target_id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
}
```

**5.2 Resolve Spell (ดูเวทที่จะได้)**

```
GET /api/v1/combat/resolve-spell?element_id=5&mastery_id=1&caster_element_id=1
Authorization: Bearer <access_token>
```

#### 📤 Response (Perform Action)

```json
{
   "success": true,
   "message": "Action performed successfully",
   "data": {
      "updatedMatch": {
         "id": "01932f5d-8e9f-7890-abcd-ef1234567890",
         "current_turn": 2,
         "current_phase": "ACTION",
         "active_combatant_id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
         "combatants": [
            {
               "id": "01932f5d-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
               "current_hp": 1023,
               "current_mp": 310, // Consumed 20 MP for spell
               "current_ap": 0
            },
            {
               "id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
               "current_hp": 650, // Took 150 damage
               "current_mp": 9999,
               "current_ap": 0
            }
         ],
         "combat_logs": [
            {
               "turn": 1,
               "action": "CAST_SPELL",
               "caster_id": "01932f5d-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
               "target_id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
               "spell_id": 101,
               "damage_dealt": 150,
               "mp_consumed": 20,
               "timestamp": "2025-11-01T10:16:00Z"
            }
         ]
      },
      "performedAction": {
         "action_type": "CAST_SPELL",
         "cast_mode": "INSTANT",
         "spell_id": 101,
         "target_id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
      }
   }
}
```

#### ✅ Status: **MOSTLY COMPLETE (95%)**

**✅ Implemented:**

-  ✅ END_TURN action
-  ✅ CAST_SPELL action (INSTANT, CHARGE, OVERCHARGE modes)
-  ✅ Spell resolution (element + mastery → spell lookup)
-  ✅ Damage calculation with Talent bonuses
-  ✅ MP consumption
-  ✅ Target validation
-  ✅ Turn progression
-  ✅ Combat log recording
-  ✅ AI opponent actions (basic)
-  ✅ Effect application (Damage, Heal, Buff, Debuff, DoT, Shield)
-  ✅ Talent Secondary Effects:
   -  ✅ S: Physical Defense (reduces incoming physical damage)
   -  ✅ L: Heal Bonus (increases healing effectiveness)
   -  ✅ G: Improvisation (chance to multicast spells)
   -  ✅ P: Duration Extension (extends DoT/HoT/Buff/Debuff duration)

**⏳ Partially Implemented:**

-  ⚠️ AI decision making (basic rule-based, ไม่ฉลาดมาก)
-  ⚠️ Complex effect interactions (synergies ยังไม่ครบ)

**❌ Not Implemented:**

-  ❌ FORFEIT action (ยอมแพ้)
-  ❌ STATUS_EFFECT actions (dispel, cleanse)
-  ❌ Advanced AI strategies

#### 🧮 Damage Calculation Example

```
Spell: Fireball (Base DMG = 100, Element = S, Mastery = Destruction)
Caster: TalentS = 93, TalentL = 3, MasteryLevel = 5

Step 1: Mastery Bonus
MasteryBonus = MasteryLevel² = 5² = 25

Step 2: Talent Bonus (Primary Element)
TalentBonus = TalentS / 10 = 93 / 10 = 9

Step 3: Total Damage
FinalDMG = (BaseDMG + MasteryBonus + TalentBonus) × CastModifier
         = (100 + 25 + 9) × 1.0 (INSTANT)
         = 134 DMG

If CHARGE mode: × 1.5 = 201 DMG
If OVERCHARGE mode: × 2.0 = 268 DMG
```

#### 📂 Files Involved

-  Handler: `internal/modules/combat/handler.go` (L82, L117-136)
-  Service: `internal/modules/combat/service.go` (L271-449)
-  Spell Calculation: `internal/modules/combat/spell_calculation.go`
-  Effect Manager: `internal/modules/combat/effect_manager.go`
-  AI Manager: `internal/modules/combat/ai_manager.go`

---

### **6️⃣ Match Victory (จบการต่อสู้)**

#### 🤖 Auto-Detection (ไม่ต้องเรียก API)

**Match จะจบอัตโนมัติเมื่อ:**

-  ✅ ศัตรูทุกตัวตาย (HP ≤ 0) → Player VICTORY
-  ✅ ผู้เล่นตาย (HP ≤ 0) → Player DEFEAT
-  ❌ เทิร์นเกินจำนวนที่กำหนด (ยังไม่ implement)
-  ❌ ผู้เล่นยอมแพ้ (ยังไม่ implement)

#### 📤 Response (จาก PerformAction ที่ทำให้จบ)

```json
{
   "success": true,
   "message": "Action performed successfully",
   "data": {
      "updatedMatch": {
         "id": "01932f5d-8e9f-7890-abcd-ef1234567890",
         "status": "COMPLETED", // ⭐ เปลี่ยนจาก IN_PROGRESS
         "winner_id": "01932f5d-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
         "result": "VICTORY",
         "ended_at": "2025-11-01T10:20:00Z",
         "combatants": [
            {
               "id": "01932f5d-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
               "current_hp": 850,
               "current_mp": 200
            },
            {
               "id": "01932f5d-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
               "current_hp": 0, // ⭐ Dead
               "current_mp": 9999
            }
         ]
      }
   }
}
```

#### ⚠️ Status: **PARTIAL (70%)**

**✅ Implemented:**

-  ✅ Victory/Defeat detection
-  ✅ Match status update (IN_PROGRESS → COMPLETED)
-  ✅ Winner recording
-  ✅ Combat log finalization

**❌ Not Implemented:**

-  ❌ Match timeout (max turns)
-  ❌ Forfeit functionality
-  ❌ Draw conditions
-  ❌ Reconnection handling (if disconnected mid-match)

#### 📂 Files Involved

-  Service: `internal/modules/combat/service.go` (Method: `_EndMatch()`)
-  Turn Manager: `internal/modules/combat/turn_manager.go`

---

### **7️⃣ Rewards & Progression (รางวัลและการพัฒนาตัวละคร)**

#### 🤖 Auto-Processing (ทำอัตโนมัติเมื่อ Match จบ)

**7.1 EXP Gain (ได้ค่าประสบการณ์)**

-  ✅ TRAINING Match: +50 EXP
-  ✅ STORY Match: +100 EXP
-  ⏳ PVP Match: +150 EXP (ยังไม่ implement PVP)

**7.2 Level Up (เลเวลขึ้น)**

-  ⏳ EXP threshold check
-  ⏳ Level increment
-  ⏳ Stat recalculation
-  ❌ Talent points allocation (ยังไม่มีระบบแจก talent points)
-  ❌ Spell unlock (ยังไม่มีระบบปลดล็อคเวท)

**7.3 Mastery EXP (MXP)**

-  ❌ Not implemented yet
-  ❌ No MXP gain from spell usage
-  ❌ No mastery level up system

**7.4 Rewards (ไอเทม/การ์ด)**

-  ❌ Not implemented yet
-  ❌ No item drops
-  ❌ No card rewards

#### ⚠️ Status: **BASIC ONLY (50%)**

**✅ Implemented:**

-  ✅ EXP reward calculation based on match type
-  ✅ EXP added to character on victory
-  ✅ Database update for character EXP

**⏳ Partially Implemented:**

-  ⚠️ Level up system (EXP gain works, but level progression not complete)

**❌ Not Implemented:**

-  ❌ Talent points distribution on level up
-  ❌ Mastery EXP (MXP) system
-  ❌ Spell unlock system
-  ❌ Item/Card rewards
-  ❌ Quest completion tracking
-  ❌ Achievement system

#### 🧮 EXP Calculation

```go
// internal/modules/combat/service.go
func (s *combatService) _CalculateExpReward(matchType string) int {
    switch matchType {
    case "TRAINING":
        return 50
    case "STORY":
        return 100
    case "PVP":
        return 150
    default:
        return 0
    }
}
```

#### 📂 Files Involved

-  Combat Service: `internal/modules/combat/service.go` (Method: `_EndMatch()`)
-  Character Service: `internal/modules/character/service.go` (Method: `GrantExp()`)

---

## 🎯 Priority Roadmap

### 🔴 HIGH Priority (ต้องทำก่อน)

**1. Complete Level Up System**

-  [ ] Implement level threshold calculation
-  [ ] Auto-increment level on EXP threshold
-  [ ] Recalculate MaxHP/MaxMP on level up
-  [ ] Grant talent points on level up
-  [ ] Create API for talent point allocation

**2. Implement Mastery EXP (MXP)**

-  [ ] Calculate MXP gain from spell usage
-  [ ] Track MXP in combat logs
-  [ ] Update mastery levels on MXP threshold
-  [ ] Apply mastery bonus to spell damage

**3. Complete Match Victory Flow**

-  [ ] Add match timeout (max turns)
-  [ ] Implement forfeit action
-  [ ] Handle disconnection/reconnection
-  [ ] Add rewards calculation (items/cards)

---

### 🟡 MEDIUM Priority (ทำต่อได้)

**4. Story Mode Implementation**

-  [ ] Load stage data from PVE repository
-  [ ] Load stage enemies
-  [ ] Apply stage modifiers
-  [ ] Track stage completion
-  [ ] Unlock next stages

**5. PVP Mode Implementation**

-  [ ] Create opponent combatant from player character
-  [ ] Load opponent's deck
-  [ ] Handle turn-based multiplayer
-  [ ] Implement matchmaking (optional)

**6. Spell Unlock System**

-  [ ] Define spell unlock requirements (level, mastery)
-  [ ] Check unlocked spells on character load
-  [ ] Filter available spells in deck builder
-  [ ] Show unlock progress in UI

---

### 🟢 LOW Priority (ไม่เร่งด่วน)

**7. Advanced AI Strategies**

-  [ ] Smart target selection
-  [ ] Spell priority system
-  [ ] Defensive behavior when low HP
-  [ ] Combo detection

**8. Match Cleanup System**

-  [ ] Implement stale match cleanup (cron job)
-  [ ] AbortMatch API for manual cleanup
-  [ ] Handle abandoned matches

**9. Additional Features**

-  [ ] Status effect actions (dispel, cleanse)
-  [ ] Match replay system
-  [ ] Combat statistics tracking
-  [ ] Leaderboards

---

## 📊 Completion Metrics

| Category              | Total Tasks | Completed | In Progress | Not Started | % Complete |
| --------------------- | ----------- | --------- | ----------- | ----------- | ---------- |
| **Core Flow**         | 7           | 5         | 1           | 1           | **85%**    |
| Account & Auth        | 1           | 1         | 0           | 0           | 100%       |
| Character Creation    | 1           | 1         | 0           | 0           | 100%       |
| Deck Management       | 1           | 1         | 0           | 0           | 100%       |
| Match Creation        | 1           | 0         | 1           | 0           | 80%        |
| Combat Actions        | 1           | 1         | 0           | 0           | 95%        |
| Match Victory         | 1           | 0         | 1           | 0           | 70%        |
| Rewards & Progression | 1           | 0         | 1           | 0           | 50%        |
| **Extended Features** | 15          | 0         | 3           | 12          | **20%**    |
| Level Up System       | 2           | 0         | 1           | 1           | 50%        |
| Mastery System        | 2           | 0         | 0           | 2           | 0%         |
| Story Mode            | 3           | 0         | 0           | 3           | 0%         |
| PVP Mode              | 3           | 0         | 0           | 3           | 0%         |
| Spell Unlock          | 2           | 0         | 0           | 2           | 0%         |
| Advanced AI           | 2           | 0         | 1           | 1           | 40%        |
| Match Cleanup         | 1           | 0         | 1           | 0           | 60%        |

**Grand Total:** 22 tasks, 5 complete, 4 in progress, 13 not started  
**Overall Completion:** 🟢 **68%** (Core systems functional, extended features pending)

---

## 🔍 Testing Checklist

### ✅ Ready to Test

-  [x] Account creation and login
-  [x] Character creation
-  [x] Deck creation and management
-  [x] TRAINING match creation
-  [x] Spell casting (INSTANT, CHARGE, OVERCHARGE)
-  [x] Damage calculation with talents
-  [x] Effect application (all types)
-  [x] Match victory detection
-  [x] EXP gain on victory

### ⏳ Partially Testable

-  [ ] Level up (EXP increments but no level up yet)
-  [ ] AI opponent (works but not smart)
-  [ ] Match timeout (needs implementation)

### ❌ Not Testable Yet

-  [ ] STORY mode
-  [ ] PVP mode
-  [ ] Mastery EXP gain
-  [ ] Talent point allocation
-  [ ] Spell unlock
-  [ ] Item/Card rewards
-  [ ] Forfeit action

---

## 📚 Related Documentation

-  [GAME_MECHANICS_DOCUMENTATION.md](GAME_MECHANICS_DOCUMENTATION.md) - Complete mechanics reference
-  [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md) - Development progress tracking
-  [docs/archive/combat-README.md](docs/archive/combat-README.md) - Combat system architecture
-  [docs/archive/MATCH_TYPES_GUIDE.md](docs/archive/MATCH_TYPES_GUIDE.md) - Match type implementation guide

---

## 🎮 Quick Start Guide (For Testers)

**Step-by-step testing flow:**

```bash
# 1. Register account
POST /api/v1/players/register
{ "username": "test", "email": "test@example.com", "password": "test1234" }

# 2. Login
POST /api/v1/players/login
{ "username": "test", "password": "test1234" }
# → Save access_token

# 3. Create character
POST /api/v1/characters/
Authorization: Bearer <access_token>
{ "name": "TestChar", "gender": "MALE", "elementId": 1, "masteryId": 1 }
# → Save character_id

# 4. Create deck
POST /api/v1/decks/
Authorization: Bearer <access_token>
{
  "character_id": 1,
  "deck_name": "Test Deck",
  "slots": [
    { "slot_num": 1, "element_id": 5 },
    { "slot_num": 2, "element_id": 6 },
    { "slot_num": 3, "element_id": 7 },
    { "slot_num": 4, "element_id": 8 },
    { "slot_num": 5, "element_id": 9 },
    { "slot_num": 6, "element_id": 10 },
    { "slot_num": 7, "element_id": 11 },
    { "slot_num": 8, "element_id": 12 }
  ]
}
# → Save deck_id

# 5. Create match
POST /api/v1/combat/
Authorization: Bearer <access_token>
{
  "character_id": 1,
  "match_type": "TRAINING",
  "deck_id": 1,
  "training_enemies": [{ "enemy_id": 1 }]
}
# → Save match_id

# 6. Cast spell
POST /api/v1/combat/:match_id/actions
Authorization: Bearer <access_token>
{
  "action_type": "CAST_SPELL",
  "cast_mode": "INSTANT",
  "spell_id": 101,
  "target_id": "<enemy_combatant_id>"
}

# 7. Repeat until match ends
# Check match.status == "COMPLETED"
# Check character's EXP increased
```

---

**Last Updated:** November 1, 2025  
**Maintainer:** nipon.k  
**Status:** 🟢 Active Development
