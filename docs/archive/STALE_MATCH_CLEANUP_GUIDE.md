# 🧹 Stale Match Cleanup System

> ⚠️ **ARCHIVED:** November 1, 2025  
> **Status:** Design document - Not yet implemented  
> **Note:** Feature planned for future development

---

## 📋 ภาพรวม

ระบบจัดการ Match ค้าง (Stale/Abandoned Matches) เพื่อป้องกันปัญหา:

-  Match ที่ค้างในสถานะ `IN_PROGRESS` นานเกินไป
-  ผู้เล่น disconnect แล้วไม่กลับมาเล่นต่อ
-  Database bloat จาก match ที่ไม่มีคนเล่นแล้ว
-  ป้องกันการเปิดหลายห้องพร้อมกัน (1 character = 1 active match)

---

## 🔧 โครงสร้างระบบ

### 1. **Database Schema**

```go
type CombatMatch struct {
    ID         uuid.UUID   `gorm:"type:uuid;primaryKey"`
    Status     MatchStatus `gorm:"type:varchar(20);not null"` // IN_PROGRESS, FINISHED, ABORTED
    CreatedAt  time.Time
    UpdatedAt  time.Time   `gorm:"index"` // ✅ ใช้สำหรับตรวจจับ match ค้าง (GORM auto-update)
    FinishedAt *time.Time
}
```

**UpdatedAt** = GORM อัปเดตอัตโนมัติทุกครั้งที่มีการเปลี่ยนแปลงใดๆ ใน match:

-  Player action (CAST_SPELL, END_TURN)
-  AI processing
-  Effect duration decay
-  Combatant stats change

**Match ค้าง** = `UpdatedAt` ไม่เปลี่ยนเลยนานเกินกำหนด (หมายถึงไม่มีการเคลื่อนไหวใดๆ ในระบบ)

---

### 2. **Repository Methods**

```go
type CombatRepository interface {
    // 🔍 หา match ที่ไม่มีความเคลื่อนไหวเกิน X นาที
    FindStaleMatches(inactiveMinutes int) ([]*domain.CombatMatch, error)

    // 🗑️ Abort match ค้างทั้งหมด (bulk operation)
    AbortStaleMatches(inactiveMinutes int) (int64, error)

    // 👤 หา active match ของผู้เล่น (ป้องกันเปิดหลายห้อง)
    FindPlayerActiveMatch(characterID uint) (*domain.CombatMatch, error)

    // ❌ Abort match เฉพาะ ID (forfeit/disconnect)
    AbortMatchByID(matchID string, reason string) (*domain.CombatMatch, error)
}
```

---

### 3. **Service Methods**

```go
type CombatService interface {
    // 🧹 ทำความสะอาด match ค้าง (สำหรับ Cron Job)
    CleanupStaleMatches(inactiveMinutes int) (int64, error)

    // ❌ Abort match เฉพาะ (Forfeit/Disconnect)
    AbortMatch(matchID string, reason string) error

    // 🔍 ตรวจสอบว่าผู้เล่นมี active match อยู่หรือเปล่า
    GetPlayerActiveMatch(characterID uint) (*domain.CombatMatch, error)
}
```

---

## 🔄 Flow การทำงาน

### **Flow 1: ป้องกันการเปิดหลายห้อง**

```
POST /combat/matches (CreateMatch)
  ↓
1. ตรวจสอบ character ownership
  ↓
2. 🆕 GetPlayerActiveMatch(characterID)
  ├─ มี active match อยู่แล้ว → 409 MATCH_ALREADY_ACTIVE
  └─ ไม่มี → สร้าง match ใหม่
  ↓
3. CreateMatch()
  └─ GORM sets CreatedAt, UpdatedAt = NOW()
```

### **Flow 2: อัปเดต Timestamp อัตโนมัติ**

```
POST /combat/matches/:id/actions (PerformAction)
  ↓
1. Validate match & turn
  ↓
2. Execute action (CAST_SPELL/END_TURN)
  ↓
3. AI processing (ถ้ามี)
  ↓
4. UpdateMatch() → GORM อัปเดต UpdatedAt อัตโนมัติ
  └─ ไม่ต้อง manual update ใดๆ!
```

### **Flow 3: Cleanup Job (Cron)**

```
Cron Job (ทุก 5-10 นาที)
  ↓
CleanupStaleMatches(30) // abort match ที่ไม่มีความเคลื่อนไหวเกิน 30 นาที
  ↓
SQL: UPDATE combat_matches
     SET status = 'ABORTED', finished_at = NOW()
     WHERE status = 'IN_PROGRESS'
       AND last_activity_at < NOW() - INTERVAL '30 minutes'
  ↓
Return: จำนวน match ที่ถูก abort
```

### **Flow 4: Manual Abort (Forfeit)**

```
POST /combat/matches/:id/abort
  ↓
AbortMatch(matchID, "player_forfeit")
  ↓
1. หา match จาก DB
2. ตรวจสอบว่ายัง IN_PROGRESS อยู่หรือเปล่า
3. UPDATE status = 'ABORTED', finished_at = NOW()
  ↓
Return: updated match
```

---

## 📝 ตัวอย่างการใช้งาน

### **1. ตรวจสอบ Active Match ก่อนสร้างใหม่**

```go
// ใน CreateMatch()
activeMatch, err := s.combatRepo.FindPlayerActiveMatch(req.CharacterID)
if err != nil {
    return nil, apperrors.SystemError("failed to check active match")
}
if activeMatch != nil {
    return nil, apperrors.New(409, "MATCH_ALREADY_ACTIVE",
        fmt.Sprintf("character already has an active match: %s", activeMatch.ID))
}
```

### **2. Cleanup Cron Job**

```go
// ใน main.go หรือ scheduler package
func setupCleanupJob(combatService combat.CombatService) {
    ticker := time.NewTicker(10 * time.Minute) // ทุก 10 นาที
    go func() {
        for range ticker.C {
            count, err := combatService.CleanupStaleMatches(30) // abort หลัง 30 นาที
            if err != nil {
                log.Printf("Cleanup failed: %v", err)
            } else if count > 0 {
                log.Printf("Aborted %d stale matches", count)
            }
        }
    }()
}
```

### **3. Forfeit Match (REST API)**

```go
// ใน handler.go
func (h *combatHandler) AbortMatch(c *gin.Context) {
    matchID := c.Param("id")

    err := h.combatService.AbortMatch(matchID, "player_forfeit")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "match aborted successfully"})
}
```

---

## ⚙️ Configuration

### **แนะนำค่า Timeout**

| Match Type   | Inactive Timeout | Cleanup Interval |
| ------------ | ---------------- | ---------------- |
| **TRAINING** | 30 minutes       | 10 minutes       |
| **STORY**    | 30 minutes       | 10 minutes       |
| **PVP**      | 15 minutes       | 5 minutes        |

**เหตุผล:**

-  **TRAINING**: ผู้เล่นอาจทดลองศึกษา AI → ให้เวลานานกว่า
-  **STORY**: เล่นปกติ → 30 นาที
-  **PVP**: ผู้คนรอกันอยู่ → abort เร็วกว่า

---

## 🗄️ SQL Queries ที่ใช้

### **1. หา Stale Matches**

```sql
SELECT * FROM combat_matches
WHERE status = 'IN_PROGRESS'
  AND updated_at < NOW() - INTERVAL '30 minutes'
ORDER BY updated_at ASC;
```

### **2. Abort Stale Matches (Bulk)**

```sql
UPDATE combat_matches
SET status = 'ABORTED',
    finished_at = NOW(),
    updated_at = NOW()
WHERE status = 'IN_PROGRESS'
  AND updated_at < NOW() - INTERVAL '30 minutes';
```

### **3. หา Active Match ของผู้เล่น**

```sql
SELECT cm.* FROM combat_matches cm
JOIN combatants c ON c.match_id = cm.id
WHERE cm.status = 'IN_PROGRESS'
  AND c.character_id = ?
LIMIT 1;
```

### **4. ดู Statistics**

````sql
-- จำนวน active matches
SELECT COUNT(*) FROM combat_matches WHERE status = 'IN_PROGRESS';

```sql
-- Match ที่ค้างนานที่สุด
SELECT id, updated_at, NOW() - updated_at AS idle_duration
FROM combat_matches
WHERE status = 'IN_PROGRESS'
ORDER BY updated_at ASC
LIMIT 10;
````

````

---

## 🎯 Best Practices

### ✅ DO

1. **เรียก GetPlayerActiveMatch ก่อน CreateMatch เสมอ**

   -  ป้องกันการเปิดหลายห้องพร้อมกัน

2. **ให้ GORM จัดการ UpdatedAt อัตโนมัติ**

   -  UpdateMatch() จะอัปเดต UpdatedAt เอง ไม่ต้อง manual update

3. **ใช้ Cron Job สำหรับ cleanup**

   -  ไม่ควรรอให้ผู้เล่นกดปุ่มถึงจะ cleanup

4. **Log การ abort**
   -  เก็บจำนวน match ที่ถูก abort เพื่อ monitor

### ❌ DON'T

1. **อย่า hard-delete match**

   -  ใช้ ABORTED status แทน (เพื่อ audit trail)

2. **อย่าลืม index UpdatedAt**

   -  Query จะช้ามากถ้าไม่มี index

3. **อย่า cleanup บ่อยเกินไป**
   -  5-10 นาที/ครั้ง เพียงพอ

---

## 🧪 การทดสอบ

### **Test Case 1: ป้องกันการเปิดหลายห้อง**

```bash
# สร้าง match แรก
curl -X POST http://localhost:8080/combat/matches \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "character_id": 1,
    "match_type": "TRAINING",
    "training_enemies": [{"enemy_id": 1}]
  }'
# Response: 201 Created

# พยายามสร้าง match ที่ 2 (ควรถูกปฏิเสธ)
curl -X POST http://localhost:8080/combat/matches \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "character_id": 1,
    "match_type": "TRAINING",
    "training_enemies": [{"enemy_id": 2}]
  }'
# Response: 409 MATCH_ALREADY_ACTIVE
````

### **Test Case 2: Cleanup Stale Matches**

```bash
# 1. สร้าง match
# 2. รอ 31 นาที (หรือแก้ updated_at ใน DB)
UPDATE combat_matches
SET updated_at = NOW() - INTERVAL '31 minutes'
WHERE id = '...';

# 3. เรียก cleanup
curl -X POST http://localhost:8080/admin/cleanup-matches \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# Response: {"aborted_count": 1}

# 4. ตรวจสอบใน DB
SELECT status FROM combat_matches WHERE id = '...';
# Result: ABORTED
```

### **Test Case 3: Forfeit Match**

```bash
curl -X POST http://localhost:8080/combat/matches/{match_id}/abort \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"reason": "player_forfeit"}'
# Response: 200 OK

# ตรวจสอบ status
curl http://localhost:8080/combat/matches/{match_id}
# Response: {"status": "ABORTED", "finished_at": "2025-10-28T..."}
```

---

## 📊 Monitoring

### **Dashboard Metrics**

```sql
-- Active matches count
SELECT COUNT(*) as active_matches
FROM combat_matches
WHERE status = 'IN_PROGRESS';

-- Stale matches (idle > 30 min)
SELECT COUNT(*) as stale_matches
FROM combat_matches
WHERE status = 'IN_PROGRESS'
  AND updated_at < NOW() - INTERVAL '30 minutes';

-- Average match duration
SELECT AVG(EXTRACT(EPOCH FROM (finished_at - created_at))/60) as avg_minutes
FROM combat_matches
WHERE status = 'FINISHED'
  AND finished_at IS NOT NULL;

-- Abort rate
SELECT
  (COUNT(CASE WHEN status = 'ABORTED' THEN 1 END)::float / COUNT(*) * 100) as abort_rate_percent
FROM combat_matches
WHERE created_at > NOW() - INTERVAL '7 days';
```

---

## 🔮 Future Enhancements

1. **Dynamic Timeout ตาม Match Type**

   ```go
   func getTimeoutMinutes(matchType MatchType) int {
       switch matchType {
       case MatchTypePVP: return 15
       case MatchTypeStory: return 30
       case MatchTypeTraining: return 60
       default: return 30
       }
   }
   ```

2. **Warning ก่อน Auto-Abort**

   -  ส่ง notification เตือนผู้เล่นก่อน 5 นาที

3. **Reconnect System**

   -  เก็บ match ไว้ 5 นาทีหลัง disconnect

4. **Analytics Dashboard**
   -  แสดงกราฟจำนวน match ค้างตามเวลา

---

## ✅ Checklist การ Deploy

-  [ ] เพิ่ม index สำหรับ `updated_at` (GORM จะสร้างให้อัตโนมัติ)
-  [ ] เพิ่ม cleanup endpoint ใน admin handler
-  [ ] Setup cron job (10 นาที/ครั้ง)
-  [ ] ทดสอบ FindPlayerActiveMatch ใน CreateMatch
-  [ ] ทดสอบ abort ผ่าน API
-  [ ] Monitor abort rate ใน dashboard
-  [ ] Document timeout values ใน README

---

**Version:** 1.0  
**Last Updated:** 2025-10-28  
**Author:** Combat System Team
