# Match Types Implementation Guide

> ⚠️ **ARCHIVED:** November 1, 2025  
> **Status:** Design document - Partially implemented  
> **Note:** TRAINING mode complete, STORY/PVP modes pending

---

**Date:** 28 ตุลาคม 2025  
**Status:** ✅ Ready for Implementation

---

## 📋 Overview

ระบบ Combat รองรับ **3 ประเภทการต่อสู้**:

1. **TRAINING** - โหมดฝึกซ้อม (เลือกศัตรูเอง)
2. **STORY** - โหมดเนื้อเรื่อง (ด่านที่กำหนด)
3. **PVP** - ต่อสู้กับผู้เล่นอื่น

---

## 🏗️ Architecture Changes

### 1. Domain Model (`combat_match.go`)

**เพิ่ม MatchType enum:**

```go
type MatchType string

const (
    MatchTypeTraining MatchType = "TRAINING"   // โหมดฝึกซ้อม
    MatchTypeStory    MatchType = "STORY"      // โหมดเนื้อเรื่อง
    MatchTypePVP      MatchType = "PVP"        // PvP
)
```

**เพิ่มฟิลด์ใน CombatMatch:**

```go
type CombatMatch struct {
    ID          uuid.UUID
    MatchType   MatchType  // NEW: ประเภทการต่อสู้
    StageID     *uint      // NEW: ID ของด่าน (สำหรับ STORY)
    Status      MatchStatus
    // ... other fields
}
```

### 2. Request DTO (`handler.go`)

**เพิ่มฟิลด์ใน CreateMatchRequest:**

```go
type CreateMatchRequest struct {
    CharacterID     uint
    MatchType       string  // "TRAINING", "STORY", "PVP"

    // Optional fields (ขึ้นอยู่กับ MatchType)
    StageID         *uint   // Required for STORY
    OpponentID      *uint   // Required for PVP
    TrainingEnemies []TrainingEnemyInput  // Required for TRAINING    DeckID          *uint
    Modifiers       *domain.MatchModifiers
}
```

### 3. Service Logic (`service.go`)

**CreateMatch ปรับเป็น switch-case:**

```go
func (s *combatService) CreateMatch(...) {
    // 1-4. Setup player combatant (เหมือนเดิม)

    // 5. สร้างศัตรูตาม MatchType
    switch req.MatchType {
    case "TRAINING":
        // โหลดศัตรูจาก req.TrainingEnemies

    case "STORY":
        // โหลดศัตรูจาก Stage
        // 3. Uncomment code ใน service.go (lines 144-162)
```

#### **Priority 2: Run Migration**

```bash
psql -U your_user -d your_database -f db/migrations/add_match_type_and_stage_id.sql
```

#### **Priority 3: Test All Modes**

-  Test TRAINING mode ✅
-  Test PVP mode ✅
-  Test STORY mode (after PveRepository ready) ⏳ case "PVP":
   // สร้าง combatant ของผู้เล่นฝ่ายตรงข้าม
   }

       // 6-8. Setup match (เหมือนเดิม)

   }

````

---

## 📊 Implementation Status

### ✅ TRAINING Mode (Completed)

**Request Example:**

```json
{
   "character_id": 1,
   "match_type": "TRAINING",
   "training_enemies": [{ "enemy_id": 1 }, { "enemy_id": 2 }],
   "deck_id": 5
}
````

**Features:**

-  ✅ เลือกศัตรูได้เอง (1 หรือหลายตัว)
-  ✅ ใช้ได้กับ enemy ทุกตัวในระบบ
-  ✅ เหมาะสำหรับทดสอบและฝึกซ้อม

**Use Cases:**

-  ทดสอบ deck ใหม่
-  ฝึกฝนการเล่น
-  ทดสอบ combo/strategy

---

### 🚧 PVE_STORY Mode (Partial - Need PveRepository)

**Request Example:**

```json
{
   "character_id": 1,
   "match_type": "PVE_STORY",
   "stage_id": 101,
   "deck_id": 5
}
```

**Current Status:**

-  ✅ Request validation
-  ✅ Database schema ready
-  ⏳ **Pending: PveRepository methods**

**Required Methods:**

```go
type PveRepository interface {
    FindStageByID(stageID uint) (*domain.Stage, error)
    FindStageEnemiesByStageID(stageID uint) ([]*domain.StageEnemy, error)
}
```

**Implementation Steps:**

1. **Add methods to PveRepository interface** (`internal/modules/pve/repository.go`):

```go
type PveRepository interface {
    FindAllActiveRealms() ([]domain.Realm, error)

    // NEW: Add these methods
    FindStageByID(stageID uint) (*domain.Stage, error)
    FindStageEnemiesByStageID(stageID uint) ([]*domain.StageEnemy, error)
}
```

2. **Implement in Postgres repository** (`internal/adapters/storage/postgres/pve_repository.go`):

```go
func (r *pveRepository) FindStageByID(stageID uint) (*domain.Stage, error) {
    var stage domain.Stage
    err := r.db.
        Preload("Chapter.Realm").
        First(&stage, stageID).Error
    return &stage, err
}

func (r *pveRepository) FindStageEnemiesByStageID(stageID uint) ([]*domain.StageEnemy, error) {
    var stageEnemies []*domain.StageEnemy
    err := r.db.
        Where("stage_id = ?", stageID).
        Order("position ASC").
        Find(&stageEnemies).Error
    return stageEnemies, err
}
```

3. **Uncomment code in service.go** (lines 144-162):

```go
case "PVE_STORY":
    if req.StageID == nil {
        return nil, apperrors.InvalidFormatError("stage_id is required for STORY mode", nil)
    }

    stageData, err := s.pveRepo.FindStageByID(*req.StageID)
    if err != nil || stageData == nil {
        return nil, apperrors.NotFoundError(fmt.Sprintf("stage with id %d not found", *req.StageID))
    }

    stageEnemies, err := s.pveRepo.FindStageEnemiesByStageID(*req.StageID)
    if err != nil {
        return nil, apperrors.SystemError("failed to load stage enemies")
    }

    for _, stageEnemy := range stageEnemies {
        enemyData, err := s.enemyRepo.FindByID(stageEnemy.EnemyID)
        if err != nil || enemyData == nil {
            continue // Skip if enemy not found
        }

        enemyCombatantID, _ := uuid.NewV7()
        enemyCombatant := &domain.Combatant{
            ID:         enemyCombatantID,
            EnemyID:    &enemyData.ID,
            Initiative: enemyData.Initiative,
            CurrentHP:  enemyData.MaxHP,
            CurrentMP:  9999,
            CurrentAP:  0,
        }
        combatants = append(combatants, enemyCombatant)
    }
```

**Features (When Complete):**

-  🎯 เล่นตามเนื้อเรื่อง
-  🎯 ศัตรูถูกกำหนดโดยด่าน
-  🎯 มีรางวัลเมื่อเคลียร์ครั้งแรก
-  🎯 Unlock ด่านต่อไปเมื่อชนะ

**Use Cases:**

-  เล่นเนื้อเรื่องหลัก
-  ปลดล็อกด่านใหม่
-  รับรางวัล first clear

---

### ✅ PVP Mode (Completed - Basic)

**Request Example:**

```json
{
   "character_id": 1,
   "match_type": "PVP",
   "opponent_id": 2,
   "deck_id": 5
}
```

**Current Features:**

-  ✅ สร้าง combatant ของผู้เล่น 2 คน
-  ✅ คำนวณ stats จาก character (HP, MP, Initiative)
-  ⏳ **TODO: โหลด deck ของฝ่ายตรงข้าม**

**Future Enhancements:**

-  [ ] Matchmaking system
-  [ ] Rating/Ranking system
-  [ ] PvP rewards
-  [ ] Replay system

**Use Cases:**

-  ต่อสู้กับเพื่อน
-  Ranked matches
-  Tournament mode

---

## 🗄️ Database Migration

**Method:** GORM AutoMigrate (ไม่ต้องรัน SQL เอง)

**Domain Model Changes:**

```go
type CombatMatch struct {
    ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
    MatchType   MatchType      `gorm:"type:varchar(20);not null;default:'TRAINING'"`
    StageID     *uint          `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    Stage       *Stage         `gorm:"foreignKey:StageID;references:ID"`
    Status      MatchStatus    `gorm:"type:varchar(20);not null"`
    // ... other fields
}
```

**GORM จะสร้างให้อัตโนมัติ:**

-  ✅ คอลัมน์ `match_type` VARCHAR(20) NOT NULL DEFAULT 'TRAINING'
-  ✅ คอลัมน์ `stage_id` INTEGER (nullable)
-  ✅ Foreign Key: `stage_id` → `stages.id` (ON DELETE SET NULL, ON UPDATE CASCADE)
-  ✅ Index: `idx_combat_matches_stage_id`
-  ✅ Relation: `Stage` จะถูก preload ได้เมื่อต้องการ

**Migration Order ใน `migrate.go`:**

```go
&domain.Stage{},        // ต้องมาก่อน
&domain.CombatMatch{},  // ใช้ StageID เป็น foreign key
```

**ไม่ต้องทำอะไร** - แค่ run application แล้ว GORM จะจัดการเอง! 🎉

---

## 🧪 Testing Guide

### 1. TRAINING Mode

```bash
curl -X POST http://localhost:8080/api/v1/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "character_id": 1,
    "match_type": "TRAINING",
    "training_enemies": [{"enemy_id": 1}],
    "deck_id": 5
  }'
```

**Expected:** ✅ Match created successfully

### 2. PVE_STORY Mode

```bash
curl -X POST http://localhost:8080/api/v1/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "character_id": 1,
    "match_type": "PVE_STORY",
    "stage_id": 101,
    "deck_id": 5
  }'
```

**Expected:** ⏳ 501 NOT_IMPLEMENTED (until PveRepository is ready)

### 3. PVP Mode

```bash
curl -X POST http://localhost:8080/api/v1/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "character_id": 1,
    "match_type": "PVP",
    "opponent_id": 2,
    "deck_id": 5
  }'
```

**Expected:** ✅ Match created successfully

---

## ✅ Validation Rules

### TRAINING Mode

-  ✅ `training_enemies` is required
-  ✅ `training_enemies` must not be empty
-  ✅ Each enemy must exist in database

### PVE_STORY Mode

-  ✅ `stage_id` is required
-  ✅ Stage must exist in database
-  ⏳ Stage must be unlocked (future enhancement)

### PVP Mode

-  ✅ `opponent_id` is required
-  ✅ Opponent character must exist
-  ⏳ Cannot play against yourself (future enhancement)
-  ⏳ Matchmaking validation (future enhancement)

---

## 📝 Next Steps

### Priority 1: Complete PVE_STORY

1. ✅ Add PveRepository interface methods
2. ✅ Implement Postgres repository methods
3. ✅ Uncomment service.go code
4. ✅ Test with actual stage data

### Priority 2: Enhance PVP

1. Load opponent's deck
2. Add matchmaking system
3. Add rating/ranking
4. Add replay system

### Priority 3: Game Features

1. First clear rewards (PVE_STORY)
2. Stage unlock system
3. PvP rewards
4. Leaderboards

---

## 🎯 Summary

### What's Ready Now

-  ✅ **TRAINING** - Fully functional
-  ✅ **PVP** - Basic functionality (no deck loading)
-  ✅ Database schema
-  ✅ Request validation
-  ✅ Error handling

### What's Needed

-  ⏳ **PVE_STORY** - Need PveRepository implementation
-  ⏳ **PVP Deck Loading** - Load opponent deck
-  ⏳ **Rewards System** - First clear rewards
-  ⏳ **Unlock System** - Progressive stage unlocking

### Code Quality

-  ✅ Follows Clean Architecture
-  ✅ Proper error handling
-  ✅ Type-safe with domain models
-  ✅ Extensible design
-  ✅ Well-documented

---

**Ready to implement!** 🚀
