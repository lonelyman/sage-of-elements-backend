# Effect ID Migration - 1000-Based Structure

> ⚠️ **ARCHIVED:** November 1, 2025  
> **Status:** Migration completed successfully on October 26, 2025  
> **For Current Effect IDs:** See [GAME_MECHANICS_DOCUMENTATION.md](../../GAME_MECHANICS_DOCUMENTATION.md)

---

## Overview

This document tracks the migration from the old effect ID system (1-399) to the new 1000-based structure for better organization and scalability.

## Migration Date

October 26, 2025

## New ID Structure

### 1000s: Direct Effects (กระทำโดยตรง)

**1100-1199: HP/MP/Resource Manipulation**

-  `1101` - DAMAGE (💥 สร้างความเสียหาย HP)
-  `1102` - SHIELD (🛡️ สร้างโล่)
-  `1103` - HEAL (❤️ ฟื้นฟู HP)
-  `1104` - MP_DAMAGE (💧 สร้างความเสียหาย MP)

### 2000s: Buffs (เสริมพลัง - ติดตัวเป้าหมาย)

**2100-2199: Regeneration Buffs**

-  `2101` - BUFF_HP_REGEN (💖 ฟื้นฟู HP ต่อเนื่อง)
-  `2102` - BUFF_MP_REGEN (💙 ฟื้นฟู MP ต่อเนื่อง)

**2200-2299: Combat Stat Buffs**

-  `2201` - BUFF_EVASION (💨 เพิ่มโอกาสหลบหลีก)
-  `2202` - BUFF_DMG_UP (🔥 เพิ่มความเสียหายที่ทำ)
-  `2203` - BUFF_RETALIATION (✨ สะท้อนความเสียหาย)
-  `2204` - BUFF_DEFENSE_UP (💪 ลดความเสียหาย HP ที่ได้รับ)

### 3000s: Synergy Buffs (เสริมพลัง - เฉพาะทาง)

**3100-3199: Stance Buffs**

-  `3101` - STANCE_S (🌟 สถานะเสริมพลัง S)
-  `3102` - STANCE_L (🌟 สถานะเสริมพลัง L)
-  `3103` - STANCE_G (🌟 สถานะเสริมพลัง G)
-  `3104` - STANCE_P (🌟 สถานะเสริมพลัง P)

### 4000s: Debuffs (ลดทอน - ติดตัวเป้าหมาย)

**4100-4199: Stat Debuffs**

-  `4101` - DEBUFF_SLOW (🐢 ลดค่า Initiative)
-  `4102` - DEBUFF_VULNERABLE (🎯 ทำให้ได้รับความเสียหายแรงขึ้น)

**4200-4299: Damage Over Time (DoT) Debuffs**

-  `4201` - DEBUFF_IGNITE (🔥 สร้างความเสียหายต่อเนื่อง - เผาไหม้)

### 5000s+: Reserved for Future Expansion

-  5000 - Utility effects
-  6000 - Crowd Control effects
-  7000 - Special/Ultimate effects

## ID Mapping Table

| Old ID | New ID | Effect Name          | Type         | Category         |
| ------ | ------ | -------------------- | ------------ | ---------------- |
| 1      | 1101   | DAMAGE               | Damage       | Direct           |
| 2      | 1102   | SHIELD               | Shield       | Direct           |
| 3      | 1103   | HEAL                 | Heal         | Direct           |
| 5      | 1104   | MP_DAMAGE (DRAIN_MP) | Resource     | Direct           |
| 100    | 2101   | BUFF_HP_REGEN        | Buff         | Regeneration     |
| 101    | 2102   | BUFF_MP_REGEN        | Buff         | Regeneration     |
| 102    | 2201   | BUFF_EVASION         | Buff         | Combat Stat      |
| 103    | 2202   | BUFF_DMG_UP          | Buff         | Combat Stat      |
| 104    | 2203   | BUFF_RETALIATION     | Buff         | Combat Stat      |
| 110    | 2204   | BUFF_DEFENSE_UP      | Buff         | Combat Stat      |
| 200    | 3101   | STANCE_S             | Synergy Buff | Stance           |
| 201    | 3102   | STANCE_L             | Synergy Buff | Stance           |
| 202    | 3103   | STANCE_G             | Synergy Buff | Stance           |
| 203    | 3104   | STANCE_P             | Synergy Buff | Stance           |
| 301    | 4101   | DEBUFF_SLOW          | Debuff CC    | Stat Debuff      |
| 302    | 4102   | DEBUFF_VULNERABLE    | Debuff       | Stat Debuff      |
| 306    | 4201   | DEBUFF_IGNITE        | Debuff DOT   | Damage Over Time |

## Removed Effects (Unused)

These effects were defined in the old schema but never referenced in code:

-  ❌ `4` - TRUE_DAMAGE
-  ❌ `6` - GAIN_AP
-  ❌ `7` - CLEANSE
-  ❌ `105` - BUFF_MAX_HP
-  ❌ `106` - BUFF_CC_RESIST
-  ❌ `108` - BUFF_PENETRATION
-  ❌ `300` - DEBUFF_REDUCE_ARMOR
-  ❌ `303` - DEBUFF_ROOT
-  ❌ `304` - DEBUFF_AP_DRAIN
-  ❌ `305` - DEBUFF_STUN
-  ❌ `308` - DEBUFF_CORROSION

## Files Updated

### Database Seeder

-  ✅ `seeder.go::seedEffects()` - Updated effect definitions
-  ✅ `seeder.go::seedSpells()` - Updated all spell effect references
-  ✅ `seeder.go::seedEnemies()` - Updated all enemy ability effect references

### Spell Definitions Updated

| Spell ID | Spell Name      | Old Effect IDs | New Effect IDs |
| -------- | --------------- | -------------- | -------------- |
| 1        | EarthSlam       | 1, 200         | 1101, 3101     |
| 2        | StoneSkin       | 2, 200         | 1102, 3101     |
| 3        | Reinforce       | 110            | 2204           |
| 4        | Tremor          | 301            | 4101           |
| 5        | AquaShot        | 1              | 1101           |
| 6        | SoothingMist    | 100, 201       | 2101, 3102     |
| 7        | Meditate        | 101            | 2102           |
| 8        | MinorHeal       | 3              | 1103           |
| 9        | WindSlash       | 1              | 1101           |
| 10       | Blur            | 102            | 2201           |
| 11       | SwiftStep       | 202            | 3103           |
| 12       | Gust            | 301            | 4101           |
| 13       | PlasmaBolt      | 1              | 1101           |
| 14       | StaticField     | 2, 104         | 1102, 2203     |
| 15       | Empower         | 103, 203       | 2202, 3104     |
| 16       | Analyze         | 302            | 4102           |
| 17       | EntanglingRoots | 301            | 4101           |
| 18       | ManaBurn        | 5              | 1104           |
| 21       | Fireball        | 1, 306         | 1101, 4201     |

### Enemy Abilities Updated

| Enemy ID | Ability Name | Old Effect IDs | New Effect IDs |
| -------- | ------------ | -------------- | -------------- |
| 1        | P_PUNCH      | 1              | 1101           |
| 1        | P_TREMOR     | 1, 301         | 1101, 4101     |
| 1        | P_OVERCHARGE | 103            | 2202           |
| 2        | S_SLAP       | 1              | 1101           |
| 2        | S_HARDEN     | 110            | 2204           |
| 2        | S_QUAKE      | 1              | 1101           |
| 3        | L_SPLASH     | 1              | 1101           |
| 3        | L_REGEN      | 100            | 2101           |
| 3        | L_DROWN      | 302            | 4102           |
| 4        | G_WIND_SLASH | 1              | 1101           |
| 4        | G_EVADE      | 102            | 2201           |
| 4        | G_TORNADO    | 1              | 1101           |

## Code Module Updates Required

⚠️ **IMPORTANT**: The following modules need to be updated to use new effect IDs:

### 1. Effect Manager (`effect_manager.go`)

-  Update all hardcoded effect ID constants
-  Search for: `EffectID: 1`, `EffectID: 2`, etc.
-  Replace with new 1000-based IDs

### 2. Spell Application (`spell_application.go`)

-  Update effect ID checks in switch statements
-  Update constant comparisons

### 3. Combat Modules

-  `spell_calculation.go` - May have effect ID references
-  `calculator.go` - May have effect type checks
-  Any other files that reference effect IDs directly

## Database Migration Steps

### Development Environment

1. ✅ Update `seeder.go` with new effect IDs
2. ⏳ Drop and recreate database tables
3. ⏳ Run seeder to populate with new data
4. ⏳ Test all combat scenarios

### Production Environment (When Ready)

```sql
-- Step 1: Create mapping table for migration
CREATE TABLE effect_id_mapping (
    old_id INT PRIMARY KEY,
    new_id INT NOT NULL
);

-- Step 2: Insert mapping data
INSERT INTO effect_id_mapping VALUES
(1, 1101), (2, 1102), (3, 1103), (5, 1104),
(100, 2101), (101, 2102), (102, 2201), (103, 2202), (104, 2203), (110, 2204),
(200, 3101), (201, 3102), (202, 3103), (203, 3104),
(301, 4101), (302, 4102), (306, 4201);

-- Step 3: Backup current data
CREATE TABLE effects_backup AS SELECT * FROM effects;
CREATE TABLE spell_effects_backup AS SELECT * FROM spell_effects;
CREATE TABLE combatant_active_effects_backup AS
    SELECT * FROM combatants; -- JSON field backup

-- Step 4: Update effect IDs in effects table
UPDATE effects e
SET id = (SELECT new_id FROM effect_id_mapping WHERE old_id = e.id)
WHERE id IN (SELECT old_id FROM effect_id_mapping);

-- Step 5: Update spell_effects junction table
UPDATE spell_effects se
SET effect_id = (SELECT new_id FROM effect_id_mapping WHERE old_id = se.effect_id)
WHERE effect_id IN (SELECT old_id FROM effect_id_mapping);

-- Step 6: Update active effects in combatants (JSONB field)
-- This requires application-level migration due to JSON structure

-- Step 7: Clean up
DROP TABLE effect_id_mapping;
```

## Testing Checklist

-  [ ] All spells cast successfully with new effect IDs
-  [ ] All enemy abilities work correctly
-  [ ] Damage calculation uses correct effect IDs
-  [ ] Buff/debuff application works
-  [ ] Shield effect works correctly
-  [ ] Healing effects work correctly
-  [ ] DoT effects (IGNITE) calculate properly
-  [ ] Stance effects apply correctly
-  [ ] MP damage effect works
-  [ ] Defense/Evasion buffs function properly
-  [ ] Vulnerable debuff increases damage correctly
-  [ ] Slow debuff affects initiative

## Benefits of New Structure

1. **Better Organization**: Effect grouped by category (1000s, 2000s, 3000s, etc.)
2. **Scalability**: Room for 100 effects per subcategory
3. **Clarity**: ID range instantly tells you the effect type
4. **Future-Proof**: Reserved ranges for new effect types
5. **Clean Schema**: Removed 12 unused effects reducing database bloat

## Rollback Plan

If issues arise:

1. Restore database from backup
2. Revert `seeder.go` changes
3. Rebuild and redeploy old version
4. Investigate issues before re-attempting migration

---

**Status**: ✅ Schema Updated, ⏳ Code Migration Pending, ⏳ Testing Pending
**Last Updated**: October 26, 2025
