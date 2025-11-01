# Combat System Documentation

> ⚠️ **ARCHIVED:** November 1, 2025  
> **Reason:** Consolidated into main documentation  
> **See Instead:** [GAME_MECHANICS_DOCUMENTATION.md](../../GAME_MECHANICS_DOCUMENTATION.md)

---

**Package:** `internal/modules/combat`  
**Last Updated:** 29 ตุลาคม 2025  
**Status:** ✅ Production Ready (Archived for reference)

> 📚 **Full Documentation:** See [GAME_MECHANICS_DOCUMENTATION.md](../../GAME_MECHANICS_DOCUMENTATION.md)  
> 📊 **Implementation Status:** See [IMPLEMENTATION_STATUS.md](../../IMPLEMENTATION_STATUS.md)

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [File Organization](#file-organization)
4. [Key Systems](#key-systems)
5. [Naming Conventions](#naming-conventions)
6. [Refactoring History](#refactoring-history)

---

## 🎯 Overview

Combat System เป็นระบบต่อสู้แบบเทิร์นเบส (Turn-based Combat) ที่รองรับ:

-  ⚔️ **Player vs Enemy Combat** - การต่อสู้ระหว่างผู้เล่นกับศัตรู
-  🎴 **Spell Casting System** - ระบบร่ายเวทมนตร์พร้อม element และ mastery
-  🤖 **AI Decision Making** - AI ที่ตัดสินใจโจมตีตาม priority rules
-  💫 **Effect System** - ระบบ buffs, debuffs, DoT, synergies
-  🔄 **Turn Management** - การจัดการเทิร์นและทรัพยากร (AP, MP)

---

## 🏗️ Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────┐
│  Handler (HTTP)                             │
│  - Validate requests                        │
│  - Call service methods                     │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  Service (Business Logic)                   │
│  - CreateMatch()                            │
│  - PerformAction()                          │
│  - processAITurn()                          │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  Repository (Data Access)                   │
│  - GetMatch()                               │
│  - SaveMatch()                              │
│  - GetCharacter(), GetEnemy(), etc.         │
└─────────────────────────────────────────────┘
```

### Combat Flow

```
1. Player Action (PerformAction)
   ├─ Validate player permissions
   ├─ Execute spell casting
   │  ├─ Preparation & Validation
   │  ├─ Calculation (damage/heal/effects)
   │  ├─ Application (apply to target)
   │  └─ Update combatant state
   ├─ Process effects (tick, expiry)
   ├─ End player turn
   └─ Start next turn (AI or player)

2. AI Turn (processAITurn)
   ├─ Validate AI combatant
   ├─ Prepare decision context
   ├─ Select next action (SelectNextAction)
   │  ├─ Evaluate conditions
   │  ├─ Check resources (AP/MP)
   │  └─ Determine target
   ├─ Execute action loop
   │  ├─ Execute AI action
   │  ├─ Deduct resources
   │  └─ Apply ability effects
   └─ End AI turn

3. Turn Management
   ├─ End current turn (endTurn)
   ├─ Rotate to next combatant
   ├─ Start new turn (startNewTurn)
   │  ├─ Process turn effects (DoT/HoT)
   │  ├─ Regenerate AP
   │  └─ Regenerate MP (players only)
   └─ Check match end condition
```

---

## 📁 File Organization

### Core Files (Public APIs)

| File            | Purpose                     | Key Functions                               |
| --------------- | --------------------------- | ------------------------------------------- |
| `handler.go`    | HTTP handlers               | HandleCreateMatch, HandlePerformAction      |
| `service.go`    | Business logic entry points | CreateMatch, PerformAction                  |
| `repository.go` | Data access                 | GetMatch, SaveMatch, GetCharacter, GetEnemy |

### Spell System

| File                     | Purpose                 | Key Functions                                            |
| ------------------------ | ----------------------- | -------------------------------------------------------- |
| `spell_resolver.go`      | Spell resolution        | ResolveSpell, findFallbackSpell                          |
| `spell_cast_executor.go` | Main orchestrator       | ExecuteSpellCast                                         |
| `spell_preparation.go`   | Validation & setup      | PrepareAndValidateCast                                   |
| `spell_calculation.go`   | Damage/heal calculation | CalculateInitialEffectValues, CalculateCombinedModifiers |
| `spell_application.go`   | Apply effects           | ApplyCalculatedEffects                                   |

### Effect System

| File                | Purpose            | Effects Handled                                                                                           |
| ------------------- | ------------------ | --------------------------------------------------------------------------------------------------------- |
| `effect_manager.go` | Core orchestration | processEffectTicksAndExpiry, recalculateStats, applyEffect                                                |
| `effect_direct.go`  | Direct effects     | Damage (1101), Heal (1103), Shield (1102), MP Damage (1104)                                               |
| `effect_buffs.go`   | Buff effects       | HP Regen (2101), MP Regen (2102), Evasion (2201), Damage Up (2202), Retaliation (2203), Defense Up (2204) |
| `effect_debuffs.go` | Debuff effects     | Slow (4101), Vulnerable (4102), Ignite (4201)                                                             |
| `effect_synergy.go` | Synergy stances    | Stance S (3101), Stance L (3102), Stance G (3103), Stance P (3104)                                        |

### AI System

| File              | Purpose          | Key Functions                                             |
| ----------------- | ---------------- | --------------------------------------------------------- |
| `ai_manager.go`   | AI orchestration | processAITurn                                             |
| `ai_decision.go`  | Decision making  | SelectNextAction, \_EvaluateCondition, \_CanAffordAbility |
| `ai_execution.go` | Action execution | ExecuteAIAction, \_DeductResources, \_ApplyAbilityEffects |

### Turn & Match Management

| File              | Purpose         | Key Functions                                               |
| ----------------- | --------------- | ----------------------------------------------------------- |
| `turn_manager.go` | Turn flow       | endTurn, startNewTurn, checkMatchEndCondition               |
| `match_utils.go`  | Match utilities | findCombatantByID, findPlayerCombatant, findAliveCombatants |

### Utilities

| File            | Purpose             | Key Functions                                                     |
| --------------- | ------------------- | ----------------------------------------------------------------- |
| `calculator.go` | Legacy calculations | calculateEffectValue (deprecated), calculateTalentBonusFromRecipe |

---

## 🔑 Key Systems

### 1. Talent-Based Calculations (New! Oct 29, 2025) ⭐

**Talent Secondary Effects** - เพิ่มความหลากหลายให้กับ builds:

#### **1.1 Heal Bonus (Talent L)** ✅

```go
// Formula: (Base + (Talent L / 10)) × Power Modifier
// Effect: Increases healing effectiveness

Config: TALENT_HEAL_DIVISOR = 10
Implementation: spell_calculation.go → _CalculateHealTalentBonus()

Example:
  Base Heal: 50
  Talent L: 93
  Heal Bonus: 93 / 10 = 9.3
  Final: (50 + 9.3) × 1.2 (CHARGE) = 71.16 HP
```

#### **1.2 Improvisation - Multi-Cast (Talent G)** ✅

```go
// Formula: Chance = min(Talent G / 5, Cap)
// Effect: Chance to cast spell twice without consuming AP/MP

Configs:
  TALENT_G_MULTICAST_DIVISOR = 5
  TALENT_G_MULTICAST_CAP_TRAINING = 30%
  TALENT_G_MULTICAST_CAP_STORY = 25%
  TALENT_G_MULTICAST_CAP_PVP = 20%

Implementation:
  spell_calculation.go → _ShouldTriggerMultiCast()
  spell_cast_executor.go → STEP 6 (after successful cast)

Example:
  Talent G: 100
  Base Chance: 100 / 5 = 20%
  PVP Cap: 20%
  Final Chance: 20% (hit cap)
  → Roll: 15 → Multi-Cast triggered! 🎲✨
```

#### **1.3 DoT Duration Scaling (Talent P)** ✅

```go
// Formula: Final Duration = Base Duration + floor(Talent P / 30)
// Effect: Extends DoT/HoT/Buff/Debuff duration

Config: TALENT_P_DURATION_DIVISOR = 30
Implementation:
  spell_calculation.go → _CalculateDurationBonus()
  spell_application.go → All duration-based effects

Example:
  Base Duration: 3 turns (BURN)
  Talent P: 90
  Bonus Turns: floor(90 / 30) = 3
  Final Duration: 3 + 3 = 6 turns
  Total Damage: 20 × 6 = 120 (vs 60 without Talent P)
```

**Impact on Builds:**

-  🔥 **S-Build (Damage):** High HP, consistent damage
-  💧 **L-Build (Support):** High MP + Heal, support-oriented
-  ⚡ **G-Build (RNG/Burst):** Initiative + Multi-Cast, explosive plays
-  🌿 **P-Build (DoT/Control):** Extended effects, long-game strategy

---

### 2. Effect ID System (1000-based Hierarchy)

```
Effect IDs are organized in thousands:

1000s - Direct Effects
  ├─ 1101: Damage
  ├─ 1102: Shield
  ├─ 1103: Heal
  └─ 1104: MP Damage

2000s - Buff Effects
  ├─ 2101: HP Regeneration
  ├─ 2102: MP Regeneration
  ├─ 2201: Evasion Up
  ├─ 2202: Damage Up
  ├─ 2203: Retaliation
  └─ 2204: Defense Up

3000s - Synergy Effects
  ├─ 3101: Stance S (Evasion + Damage)
  ├─ 3102: Stance L (HP Regen + Defense)
  ├─ 3103: Stance G (Damage + Crit)
  └─ 3104: Stance P (Shield + Reflect)

4000s - Debuff Effects
  ├─ 4101: Slow (Reduce AP regen)
  ├─ 4102: Vulnerable (Take more damage)
  └─ 4201: Ignite (Fire DoT)
```

### 3. Spell Casting Workflow

```go
// 1. Preparation & Validation
PrepareAndValidateCast()
  ├─ Fetch spell data
  ├─ Find target
  ├─ Validate targeting
  ├─ Calculate final costs (AP/MP)
  ├─ Validate resources
  └─ Consume charges (if any)

// 2. Initial Calculations
CalculateInitialEffectValues()
  ├─ Get base value
  ├─ Calculate mastery bonus (Level²)
  └─ Calculate talent bonus
      ├─ HEAL (1103): Use Talent L / 10 ⭐
      └─ Others: Use Σ(Ingredient Talents) / 10

// 3. Modifier Calculations
CalculateCombinedModifiers()
  ├─ Elemental modifier (1.3x advantage, 0.8x disadvantage)
  ├─ Buff/Debuff modifier
  └─ Power modifier (casting mode: 1.0x/1.2x/1.5x)

// 4. Apply Effects
ApplyCalculatedEffects()
  ├─ Determine effect target (SELF/OPPONENT)
  ├─ Apply specific effect (damage/heal/buff/debuff)
  │  └─ Duration-based effects: Apply Talent P bonus ⭐
  ├─ Check evasion
  ├─ Apply shields
  ├─ Update stats
  └─ Return detailed results

// 5. Multi-Cast Check (NEW!) ⭐
_ShouldTriggerMultiCast()
  ├─ Check Talent G value
  ├─ Calculate chance (Talent G / 5)
  ├─ Apply cap by match type
  ├─ Roll random (0-100)
  └─ If triggered: Re-execute steps 2-4
```

### 4. AI Decision Flow

```go
SelectNextAction()
  ├─ Loop through AI rules (by priority)
  ├─ For each rule:
  │  ├─ Evaluate condition (ALWAYS/TURN_IS/SELF_HP_BELOW)
  │  ├─ Check ability exists
  │  ├─ Validate resources (AP/MP)
  │  ├─ Determine target (PLAYER/SELF)
  │  └─ Return selected action
  └─ Return nil if no valid action
```

### 5. Turn Management

```go
// End Turn
endTurn()
  ├─ Move to next combatant in rotation
  └─ Return updated match

// Start New Turn
startNewTurn()
  ├─ Process turn effects (DoT/HoT ticks)
  ├─ Regenerate AP (all combatants)
  ├─ Regenerate MP (players only)
  └─ Return updated match

// Check Match End
checkMatchEndCondition()
  ├─ Separate teams (players vs enemies)
  ├─ Check if player team defeated
  ├─ Check if enemy team defeated
  ├─ End match if either team defeated
  └─ Return match
```

---

## 🎨 Naming Conventions

### Function Visibility & Hierarchy

```go
// Level 1: Public API (PascalCase)
// - Can be called from outside the package
// - Example: Called from handler.go
func (s *combatService) CreateMatch(...)
func (s *combatService) PerformAction(...)
func (s *combatService) ResolveSpell(...)

// Level 2: Internal Functions (camelCase)
// - Private to package
// - Main workflow functions
func (s *combatService) processAITurn(...)
func (s *combatService) endTurn(...)
func (s *combatService) findCombatantByID(...)

// Level 3: Sub-Helper Functions (_camelCase)
// - Private to package
// - Helper functions called by main workflows
// - "Don't call directly" convention
func (s *combatService) _ValidateAICombatant(...)
func (s *combatService) _EvaluateCondition(...)
func (s *combatService) _CheckAlways(...)
func (s *combatService) _CanAffordAbility(...)
```

**Why use `_` prefix?**

1. **Clear Hierarchy** - Shows function is a sub-helper
2. **Don't Call Directly** - Indicates it's meant to be called by specific parent functions
3. **Better Organization** - Easy to identify helper vs main functions
4. **Code Navigation** - Quickly see function relationships

**Example Hierarchy:**

```go
SelectNextAction()              // Public API
  └─ _EvaluateCondition()       // Sub-helper
      └─ _CheckAlways()         // Sub-sub-helper
      └─ _CheckTurnIs()         // Sub-sub-helper
      └─ _CheckSelfHPBelow()    // Sub-sub-helper
  └─ _CanAffordAbility()        // Sub-helper
  └─ _DetermineTarget()         // Sub-helper
```

---

## 📚 Refactoring History

### Phase 1: Effect Manager Refactoring (Oct 26-27, 2025)

**Problem:**

-  `effect_manager.go` was 1377 lines - monolithic and hard to maintain

**Solution:**

-  Split into 5 specialized files by effect category
-  Migrated old Effect IDs (1-399) to new 1000-based structure

**Files Created:**

-  `effect_manager.go` (248 lines) - Core orchestration
-  `effect_direct.go` (475 lines) - Direct effects
-  `effect_buffs.go` (330 lines) - Buff effects
-  `effect_debuffs.go` (145 lines) - Debuff effects
-  `effect_synergy.go` (250 lines) - Synergy stances

**Impact:**

-  ✅ 73% reduction in main file size
-  ✅ Clear separation of concerns
-  ✅ Easy to add new effects
-  ✅ 29 Effect ID references updated

### Phase 2: Spell Casting Refactoring (Oct 27, 2025)

**Problem:**

-  `executeCastSpell` was 400+ lines with mixed responsibilities
-  Hard to test, debug, and extend

**Solution:**

-  Split into 5 workflow files following Clean Architecture

**Files Created:**

-  `spell_cast_executor.go` - Main orchestrator
-  `spell_preparation.go` - Validation & setup
-  `spell_calculation.go` - Damage/heal calculations
-  `spell_application.go` - Effect application
-  Removed `action_executor.go` (deprecated)

**Impact:**

-  ✅ Single Responsibility Principle
-  ✅ Testable components
-  ✅ Clear workflow steps
-  ✅ Better error handling

### Phase 3: AI System Refactoring (Oct 27, 2025)

**Problem:**

-  `ai_manager.go` mixed decision logic with execution
-  Hard to add new AI conditions or target types

**Solution:**

-  Split into 3 specialized files

**Files Created:**

-  `ai_manager.go` (220 lines) - Orchestration
-  `ai_decision.go` (180 lines) - Decision making & conditions
-  `ai_execution.go` (120 lines) - Action execution

**Impact:**

-  ✅ Clear separation: Decision vs Execution
-  ✅ Easy to add new conditions
-  ✅ Easy to add new target types
-  ✅ Better logging and debugging

### Phase 4: Talent Secondary Effects (Oct 29, 2025)

**Problem:**

-  Talent builds lacked distinctive features
-  L-builds (support) had no healing advantage
-  G-builds (speed) had no unique mechanics
-  P-builds (DoT) couldn't extend effect duration

**Solution:**

-  Implemented 3 Talent-based bonuses
-  Each talent now has unique secondary effect

**Implementation:**

-  **Heal Bonus (Talent L):**

   -  Added `_CalculateHealTalentBonus()` in `spell_calculation.go`
   -  Modified `_CalculateTalentBonus()` to handle HEAL (1103)
   -  Modified `CalculateInitialEffectValues()` to skip Mastery for HEAL
   -  Config: `TALENT_HEAL_DIVISOR = 10`

-  **Multi-Cast (Talent G):**

   -  Added `_ShouldTriggerMultiCast()` in `spell_calculation.go`
   -  Modified `ExecuteSpellCast()` in `spell_cast_executor.go` (STEP 6)
   -  Configs: `TALENT_G_MULTICAST_DIVISOR`, `TALENT_G_MULTICAST_CAP_*`

-  **DoT Duration (Talent P):**
   -  Added `_CalculateDurationBonus()` in `spell_calculation.go`
   -  Modified `spell_application.go` for Shield/Buff/Debuff/Synergy effects
   -  Config: `TALENT_P_DURATION_DIVISOR = 30`

**Impact:**

-  ✅ Build diversity - Each talent has unique identity
-  ✅ L-builds viable - Healing effectiveness scales
-  ✅ G-builds exciting - RNG-based burst potential
-  ✅ P-builds strategic - Long-game DoT/Control
-  ✅ Complete system - All 4 talents (S, L, G, P) have secondary effects

### Phase 5: Turn Manager Refactoring (Oct 28, 2025)

**Problem:**

-  `turn_manager.go` had mixed concerns
-  Config values hardcoded

**Solution:**

-  Reorganized into 5 clear sections
-  Made config-driven (AP, MP regen)

**Sections:**

1. Turn Rotation - `endTurn()`
2. Turn Initialization - `startNewTurn()`
3. Config Helpers - `_GetAPPerTurn()`, `_GetMaxAP()`, `_GetMPRegenPercent()`
4. Match End Condition - `checkMatchEndCondition()`
5. Match End Helpers - `_SeparateTeams()`, `_IsTeamDefeated()`, `_EndMatch()`

**Impact:**

-  ✅ Config-driven design
-  ✅ Easy to modify game balance
-  ✅ Clear section organization
-  ✅ Reusable helper functions

### Phase 6: Match Utils Enhancement (Oct 28, 2025)

**Problem:**

-  `helper.go` had generic name and incomplete utilities
-  Missing filtering and counting functions

**Solution:**

-  Renamed to `match_utils.go`
-  Added comprehensive utility functions
-  Used utilities in existing code

**Changes:**

-  ✅ Renamed `helper.go` → `match_utils.go`
-  ✅ Added `findPlayerCombatants()`
-  ✅ Added `findEnemyCombatants()`
-  ✅ Added `findAliveCombatants()`
-  ✅ Removed 5 unused functions
-  ✅ Applied utilities in `turn_manager.go`, `spell_preparation.go`

**Final Utilities (6 functions):**

1. `findCombatantByID()` - Find by UUID
2. `findPlayerCombatant()` - Find first player
3. `findPlayerCombatants()` - Find all players
4. `findEnemyCombatants()` - Find all enemies
5. `findAliveCombatants()` - Find alive combatants (HP > 0)

**Impact:**

-  ✅ Clear file purpose
-  ✅ DRY principle (Don't Repeat Yourself)
-  ✅ Only useful utilities remain
-  ✅ 37.5% code reduction

---

## 🎯 Current State Summary

### Statistics

```
Total Files: 20+
Total Lines: ~8000+ (well-organized)

Core System:
  - 3 entry point files (handler, service, repository)
  - 5 spell system files
  - 5 effect system files
  - 3 AI system files
  - 2 turn management files
  - 2 utility files
```

### Code Quality

-  ✅ **Clean Architecture** - Clear separation of layers
-  ✅ **Single Responsibility** - Each file has one clear purpose
-  ✅ **DRY Principle** - Utilities shared, no duplication
-  ✅ **Testable** - Functions are small and focused
-  ✅ **Maintainable** - Easy to find and modify code
-  ✅ **Extensible** - Easy to add new features
-  ✅ **Well-Documented** - Section headers and comments
-  ✅ **Build Diversity** - All talent builds are viable and unique

### Ready for Production

-  ✅ Compilable - All code compiles successfully
-  ✅ Organized - Clear file structure
-  ✅ Consistent - Following naming conventions
-  ✅ Efficient - No unused code

---

## 🚀 Future Enhancements

### Planned Features

1. **Advanced AI Conditions**

   -  TARGET_HP_BELOW - Target specific enemies
   -  ALLY_COUNT - Check team composition
   -  TURN_MOD - Actions on specific turn intervals

2. **Team Mechanics**

   -  Multi-player support
   -  Team-based buffs
   -  Combo attacks

3. **Victory Conditions**

   -  Time limits
   -  Objective-based
   -  Score-based

4. **Combat Analytics**
   -  Damage meters
   -  Healing meters
   -  Effect uptime tracking

### Development Guidelines

When adding new features:

1. **Follow naming conventions** - Use appropriate visibility levels
2. **Keep files focused** - One responsibility per file
3. **Use utilities** - Check `match_utils.go` before writing new helpers
4. **Add documentation** - Update this README when making significant changes
5. **Test thoroughly** - Write unit tests for new functions

---

## 📖 Related Documentation

-  Effect ID mapping: See `effect_manager.go` header
-  AI Conditions: See `ai_decision.go` header
-  Spell workflow: See `spell_cast_executor.go` header
-  Turn flow: See `turn_manager.go` header

---

**Maintained by:** Combat System Team  
**Questions?** Check code comments or ask in #combat-dev
