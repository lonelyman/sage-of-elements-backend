# 🎮 User Flow Scenarios: Complete Gameplay Journey

**ละเอียดทุก Screen ตั้งแต่สมัครสมาชิกถึงจบด่านแรก**  
**Date:** November 1, 2025  
**Target:** Frontend Developers, UI/UX Designers

---

## 📑 Table of Contents

1. [Scenario Overview](#scenario-overview)
2. [Screen-by-Screen Flow](#screen-by-screen-flow)
3. [API Mapping Summary](#api-mapping-summary)
4. [Error Handling Guide](#error-handling-guide)

---

## Scenario Overview

**เป้าหมาย:** ผู้เล่นใหม่สมัครสมาชิก → สร้างตัวละคร → จัด Deck → ต่อสู้ด่านแรก → ชนะและได้รางวัล

**ระยะเวลาโดยประมาณ:** 10-15 นาที (first-time user)

**จำนวน Screens:** 12 screens

**จำนวน API Calls:** ~15-20 calls

---

## Screen-by-Screen Flow

---

### 🔵 **Screen 1: Welcome / Landing Page**

**จุดประสงค์:** แนะนำเกมและให้เลือกระหว่างเข้าสู่ระบบหรือสมัครสมาชิก

#### UI Elements:

```
┌─────────────────────────────────────┐
│  🎮 SAGE OF THE ELEMENTS           │
│                                     │
│     [Epic Game Logo]                │
│                                     │
│  Master the elements,               │
│  forge your destiny                 │
│                                     │
│  [     Login      ]                 │
│  [ Register Account ]               │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

-  **None** (Static page)

#### User Actions:

-  Click "Register Account" → ไปหน้า Screen 2
-  Click "Login" → ไปหน้า Screen 2.5 (Login Screen)

---

### 🔵 **Screen 2: Registration Form**

**จุดประสงค์:** กรอกข้อมูลสมัครสมาชิก

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ← Back          Register           │
├─────────────────────────────────────┤
│                                     │
│  Create Your Account                │
│                                     │
│  Username:                          │
│  [___________________]              │
│  (min 4 characters)                 │
│                                     │
│  Email:                             │
│  [___________________]              │
│  (valid email address)              │
│                                     │
│  Password:                          │
│  [___________________] 👁           │
│  (min 8 characters)                 │
│                                     │
│  [ ] I agree to Terms & Conditions  │
│                                     │
│  [    Create Account    ]           │
│                                     │
│  Already have an account? Login     │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Register Account**

```http
POST /api/v1/players/register
Content-Type: application/json

{
  "username": "player123",
  "email": "player@example.com",
  "password": "securepass123"
}
```

**Response (Success):**

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

#### User Actions:

-  กรอกข้อมูล → Validate client-side
-  Click "Create Account" → Call API
-  Success → **Auto-login** → ไปหน้า Screen 3 (Character Creation)
-  Error → Show error message below input fields

#### Validation Rules:

-  ✅ Username: min 4 chars, unique
-  ✅ Email: valid format, unique
-  ✅ Password: min 8 chars
-  ✅ Terms checkbox: must be checked

---

### 🔵 **Screen 2.5: Login Form** (Alternative Flow)

**จุดประสงค์:** เข้าสู่ระบบสำหรับผู้เล่นเก่า

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ← Back            Login            │
├─────────────────────────────────────┤
│                                     │
│  Welcome Back!                      │
│                                     │
│  Username:                          │
│  [___________________]              │
│                                     │
│  Password:                          │
│  [___________________] 👁           │
│                                     │
│  [ ] Remember me                    │
│                                     │
│  [      Login      ]                │
│                                     │
│  Forgot password?                   │
│  Don't have an account? Register    │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Login**

```http
POST /api/v1/players/login
Content-Type: application/json

{
  "username": "player123",
  "password": "securepass123"
}
```

**Response:**

```json
{
   "success": true,
   "message": "Login successful",
   "data": {
      "accessToken": "eyJhbGciOiJIUzI1NiIs...",
      "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
   }
}
```

**2. Get Profile (after login)**

```http
GET /api/v1/players/me
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": {
      "id": 1,
      "username": "player123",
      "email": "player@example.com"
   }
}
```

**3. List Characters (check if has characters)**

```http
GET /api/v1/characters/
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": [
      // ถ้ามีตัวละครแล้ว → ไปหน้า Screen 9 (Character Selection)
      // ถ้ายังไม่มี → ไปหน้า Screen 3 (Character Creation)
   ]
}
```

#### Flow After Login:

-  ✅ Has Characters → ไปหน้า **Screen 9** (Character Selection)
-  ❌ No Characters → ไปหน้า **Screen 3** (Character Creation)

---

### 🟢 **Screen 3: Character Creation - Step 1 (Name & Gender)**

**จุดประสงค์:** กรอกชื่อและเลือกเพศของตัวละคร

#### UI Elements:

```
┌─────────────────────────────────────┐
│      Create Your Character          │
│           Step 1 of 3               │
│  [████████░░░░░░░░░░░░] 33%        │
├─────────────────────────────────────┤
│                                     │
│  Character Name:                    │
│  [___________________]              │
│  (min 3 characters)                 │
│                                     │
│  Choose Gender:                     │
│                                     │
│  ┌─────────┐  ┌─────────┐         │
│  │  MALE   │  │ FEMALE  │         │
│  │  [👨]   │  │  [👩]   │         │
│  └─────────┘  └─────────┘         │
│     ☑              ☐               │
│                                     │
│  ⓘ Gender is cosmetic only         │
│                                     │
│  [      Next Step →     ]          │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

-  **None** (Local state only)

#### User Actions:

-  กรอกชื่อ → Validate ความยาว
-  เลือกเพศ → เก็บค่าไว้
-  Click "Next Step" → ไปหน้า Screen 4

---

### 🟢 **Screen 4: Character Creation - Step 2 (Element Selection)**

**จุดประสงค์:** เลือกธาตุหลัก (Primary Element)

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ←  Create Your Character           │
│           Step 2 of 3               │
│  [████████████████░░░░░░] 66%      │
├─────────────────────────────────────┤
│                                     │
│  Choose Your Primary Element:       │
│                                     │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌─────┐│
│  │  S   │ │  L   │ │  G   │ │  P  ││
│  │ 🪨   │ │ 💧   │ │ 🌪️   │ │ ⚡  ││
│  │Solid │ │Liquid│ │ Gas  │ │Plasma││
│  └──────┘ └──────┘ └──────┘ └─────┘│
│     ☑        ☐       ☐       ☐     │
│                                     │
│  ╔════════════════════════════════╗│
│  ║ SOLIDITY (S)                   ║│
│  ║ • High HP and Defense          ║│
│  ║ • Tank/Bruiser playstyle       ║│
│  ║ • Talent S: +90 points         ║│
│  ║                                ║│
│  ║ Starting Stats:                ║│
│  ║ HP: 1,030 | MP: 175            ║│
│  ╚════════════════════════════════╝│
│                                     │
│  [      Next Step →     ]          │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Get All Elements (for display)**

```http
GET /api/v1/game-data/elements
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": [
      {
         "id": 1,
         "element_name": "S (Solidity)",
         "tier": 0,
         "is_primary": true
      },
      {
         "id": 2,
         "element_name": "L (Liquidity)",
         "tier": 0,
         "is_primary": true
      },
      {
         "id": 3,
         "element_name": "G (Gas)",
         "tier": 0,
         "is_primary": true
      },
      {
         "id": 4,
         "element_name": "P (Plasma)",
         "tier": 0,
         "is_primary": true
      }
   ]
}
```

#### User Actions:

-  API call on mount → Load primary elements
-  Click element → Show description panel
-  Click "Next Step" → เก็บ elementId ไว้ → ไปหน้า Screen 5

#### Element Stats Preview (Client-side calculation):

```javascript
// Talent Calculation
const baseTalent = 3;
const primaryBonus = 90;

// Stats Calculation
const STAT_HP_BASE = 100;
const STAT_HP_PER_TALENT_S = 10;
const STAT_MP_BASE = 100;
const STAT_MP_PER_TALENT_L = 25;

function calculateStats(elementId) {
   const talents = { S: 3, L: 3, G: 3, P: 3 };

   switch (elementId) {
      case 1:
         talents.S = 93;
         break; // Solidity
      case 2:
         talents.L = 93;
         break; // Liquidity
      case 3:
         talents.G = 93;
         break; // Gas
      case 4:
         talents.P = 93;
         break; // Plasma
   }

   const maxHP = STAT_HP_BASE + talents.S * STAT_HP_PER_TALENT_S;
   const maxMP = STAT_MP_BASE + talents.L * STAT_MP_PER_TALENT_L;

   return { maxHP, maxMP, talents };
}
```

---

### 🟢 **Screen 5: Character Creation - Step 3 (Mastery Selection)**

**จุดประสงค์:** เลือกศาสตร์หลัก (Primary Mastery)

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ←  Create Your Character           │
│           Step 3 of 3               │
│  [████████████████████████] 100%   │
├─────────────────────────────────────┤
│                                     │
│  Choose Your Primary Mastery:       │
│                                     │
│  ┌─────────┐ ┌─────────┐           │
│  │Creation │ │Destruct │           │
│  │   ✨    │ │   💥    │           │
│  └─────────┘ └─────────┘           │
│      ☑           ☐                  │
│                                     │
│  ┌─────────┐ ┌─────────┐           │
│  │Restorat │ │Transmut │           │
│  │   💚    │ │   🔄    │           │
│  └─────────┘ └─────────┘           │
│      ☐           ☐                  │
│                                     │
│  ╔════════════════════════════════╗│
│  ║ CREATION                       ║│
│  ║ • Summoning and buffs          ║│
│  ║ • Support/Utility spells       ║│
│  ║ • Shield and protective magic  ║│
│  ╚════════════════════════════════╝│
│                                     │
│  [    Create Character    ]         │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Get All Masteries (for display)**

```http
GET /api/v1/game-data/masteries
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": [
      {
         "id": 1,
         "mastery_name": "Creation",
         "description": "Focus on summoning and buffs"
      },
      {
         "id": 2,
         "mastery_name": "Destruction",
         "description": "Focus on dealing damage"
      },
      {
         "id": 3,
         "mastery_name": "Restoration",
         "description": "Focus on healing and support"
      },
      {
         "id": 4,
         "mastery_name": "Transmutation",
         "description": "Focus on transformation and debuffs"
      }
   ]
}
```

**2. Create Character (on submit)**

```http
POST /api/v1/characters/
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "name": "FireMage",
  "gender": "MALE",
  "elementId": 1,
  "masteryId": 2
}
```

**Response:**

```json
{
   "success": true,
   "message": "Character created successfully",
   "data": {
      "id": 1,
      "character_name": "FireMage",
      "gender": "MALE",
      "primary_element_id": 1,
      "level": 1,
      "exp": 0,
      "talent_s": 93,
      "talent_l": 3,
      "talent_g": 3,
      "talent_p": 3,
      "current_hp": 1023,
      "current_mp": 330,
      "masteries": [
         { "mastery_id": 1, "level": 1, "mxp": 0 },
         { "mastery_id": 2, "level": 1, "mxp": 0 },
         { "mastery_id": 3, "level": 1, "mxp": 0 },
         { "mastery_id": 4, "level": 1, "mxp": 0 }
      ],
      "tutorial": {
         "current_step": 0,
         "is_completed": false
      }
   }
}
```

#### User Actions:

-  Click mastery → Show description
-  Click "Create Character" → Call API
-  Success → Show loading animation → ไปหน้า Screen 6 (Tutorial/Welcome)

---

### 🟡 **Screen 6: Character Welcome / Tutorial Intro**

**จุดประสงค์:** ต้อนรับผู้เล่นและแนะนำเกม (Optional Tutorial)

#### UI Elements:

```
┌─────────────────────────────────────┐
│                                     │
│        Welcome, FireMage!           │
│                                     │
│     [Character Avatar/Portrait]     │
│                                     │
│  "Welcome to the world of           │
│   Elemental Mastery, young sage.    │
│   Your journey begins now..."       │
│                                     │
│  - Elder Sage                       │
│                                     │
│  [   Start Tutorial   ]             │
│  [   Skip Tutorial    ]             │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

-  **None** (Character data already loaded from Screen 5)

#### User Actions:

-  Click "Start Tutorial" → ไปหน้า Screen 7 (Deck Building Tutorial)
-  Click "Skip Tutorial" → Call Skip Tutorial API → ไปหน้า Screen 9 (Character Selection)

**Skip Tutorial API:**

```http
POST /api/v1/characters/1/tutorial/skip
Authorization: Bearer <accessToken>
```

---

### 🟡 **Screen 7: Deck Building Tutorial**

**จุดประสงค์:** สอนการสร้าง Deck

#### UI Elements:

```
┌─────────────────────────────────────┐
│  Tutorial: Building Your First Deck │
│                                     │
│  💡 "Elements are your power source │
│      Choose 8 elements to create    │
│      your battle deck"              │
│                                     │
│  Available Elements:                │
│  ┌────┐┌────┐┌────┐┌────┐         │
│  │ 🔥 ││ 💧 ││ 🌪️ ││ 🌍 │         │
│  │Fire││Water││Wind││Earth│         │
│  └────┘└────┘└────┘└────┘         │
│                                     │
│  Your Deck (0/8):                   │
│  [Empty][Empty][Empty][Empty]       │
│  [Empty][Empty][Empty][Empty]       │
│                                     │
│  Tap elements to add to deck →     │
│                                     │
│  [      Continue      ]             │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Get Elements (Tier 1 only)**

```http
GET /api/v1/game-data/elements
Authorization: Bearer <accessToken>
```

Filter client-side for `tier === 1` only

**Response:**

```json
{
   "success": true,
   "data": [
      { "id": 5, "element_name": "Fire", "tier": 1 },
      { "id": 6, "element_name": "Water", "tier": 1 },
      { "id": 7, "element_name": "Wind", "tier": 1 },
      { "id": 8, "element_name": "Earth", "tier": 1 },
      { "id": 9, "element_name": "Lightning", "tier": 1 },
      { "id": 10, "element_name": "Ice", "tier": 1 },
      { "id": 11, "element_name": "Light", "tier": 1 },
      { "id": 12, "element_name": "Dark", "tier": 1 }
   ]
}
```

**2. Create Deck**

```http
POST /api/v1/decks/
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "character_id": 1,
  "deck_name": "Starter Deck",
  "slots": [
    { "slot_num": 1, "element_id": 5 },
    { "slot_num": 2, "element_id": 5 },
    { "slot_num": 3, "element_id": 6 },
    { "slot_num": 4, "element_id": 6 },
    { "slot_num": 5, "element_id": 9 },
    { "slot_num": 6, "element_id": 9 },
    { "slot_num": 7, "element_id": 11 },
    { "slot_num": 8, "element_id": 12 }
  ]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "deck_name": "Starter Deck",
    "slots": [...]
  }
}
```

#### User Actions:

-  Click element → Add to deck slot
-  Click deck slot → Remove element
-  When 8 slots filled → Enable "Continue" button
-  Click "Continue" → Call Create Deck API → ไปหน้า Screen 8

---

### 🟡 **Screen 8: Tutorial - First Battle Intro**

**จุดประสงค์:** แนะนำการต่อสู้ครั้งแรก

#### UI Elements:

```
┌─────────────────────────────────────┐
│      Tutorial: Your First Battle    │
│                                     │
│  💡 "Time to test your skills!      │
│      Defeat this training dummy     │
│      to complete your tutorial"     │
│                                     │
│      [Enemy: Training Dummy]        │
│         HP: 300                     │
│         Easy Difficulty             │
│                                     │
│  Battle Basics:                     │
│  • Cast spells using your deck      │
│  • Elements + Mastery = Spell       │
│  • Reduce enemy HP to 0 to win      │
│                                     │
│  [    Enter Battle    ]             │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Get Tutorial Enemy Info**

```http
GET /api/v1/enemies/
Authorization: Bearer <accessToken>
```

Filter for tutorial enemy (e.g., enemy_id = 1)

**2. Create Training Match**

```http
POST /api/v1/combat/
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "character_id": 1,
  "match_type": "TRAINING",
  "deck_id": 1,
  "training_enemies": [
    { "enemy_id": 1 }
  ]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "01932f5d-8e9f-7890-abcd-ef1234567890",
    "match_type": "TRAINING",
    "status": "IN_PROGRESS",
    "current_turn": 1,
    "active_combatant_id": "01932f5d-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "combatants": [
      {
        "id": "01932f5d-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
        "character_id": 1,
        "current_hp": 1023,
        "current_mp": 330,
        "deck": [...]
      },
      {
        "id": "01932f5d-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
        "enemy_id": 1,
        "current_hp": 300,
        "current_mp": 9999
      }
    ]
  }
}
```

#### User Actions:

-  Click "Enter Battle" → Call Create Match API → ไปหน้า Screen 8.5 (Battle Screen)

---

### ⚔️ **Screen 8.5: Battle Screen (Tutorial Fight)**

**จุดประสงค์:** หน้าจอการต่อสู้

#### UI Elements:

```
┌─────────────────────────────────────┐
│  Turn 1          [ Tutorial Mode ]  │
├─────────────────────────────────────┤
│  Enemy: Training Dummy              │
│  HP: [████████████] 300/300         │
│  Status: Normal                     │
│                                     │
│       [Enemy Sprite]                │
│                                     │
│  ───────────────────────────────────│
│                                     │
│  You: FireMage                      │
│  HP: [████████████] 1023/1023       │
│  MP: [████████████] 330/330         │
│                                     │
│  Your Deck:                         │
│  [🔥][🔥][💧][💧][⚡][⚡][✨][🌑]  │
│   1   2   3   4   5   6   7   8    │
│                                     │
│  Select Element + Mastery:          │
│  Selected: 🔥 Fire                  │
│                                     │
│  Masteries:                         │
│  [Creation][Destruct][Restore][Trans]│
│      ☐        ☑        ☐       ☐   │
│                                     │
│  ⚡ Fireball (Fire + Destruction)   │
│  💥 DMG: 134 | 🔮 MP: 20            │
│                                     │
│  Cast Mode:                         │
│  ◉ Instant (1.0x) ○ Charge (1.5x)  │
│                                     │
│  [   Cast Spell   ] [  End Turn  ]  │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Resolve Spell (when selecting element + mastery)**

```http
GET /api/v1/combat/resolve-spell?element_id=5&mastery_id=2&caster_element_id=1
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": {
      "spell": {
         "id": 101,
         "spell_name": "Fireball",
         "base_damage": 100,
         "base_mp_cost": 20,
         "effects": [
            {
               "effect_id": 1001,
               "effect_name": "Direct Damage",
               "base_value": 100
            }
         ]
      },
      "element_requested": 5,
      "mastery_requested": 2,
      "caster_element_used": 1
   }
}
```

**Calculate actual damage client-side:**

```javascript
// TalentS = 93, Mastery Destruction Level = 1
const baseDamage = 100;
const masteryBonus = 1 * 1; // level²
const talentBonus = Math.floor(93 / 10); // 9
const castModifier = 1.0; // INSTANT

const finalDamage = (baseDamage + masteryBonus + talentBonus) * castModifier;
// = (100 + 1 + 9) * 1.0 = 110 DMG
```

**2. Cast Spell**

```http
POST /api/v1/combat/01932f5d-8e9f-7890-abcd-ef1234567890/actions
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "action_type": "CAST_SPELL",
  "cast_mode": "INSTANT",
  "spell_id": 101,
  "target_id": "01932f5d-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}
```

**Response:**

```json
{
   "success": true,
   "data": {
      "updatedMatch": {
         "current_turn": 2,
         "combatants": [
            {
               "id": "01932f5d-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
               "current_hp": 1023,
               "current_mp": 310,
               "deck": [
                  { "element_id": 5, "is_consumed": true },
                  { "element_id": 5, "is_consumed": false }
                  // ...
               ]
            },
            {
               "id": "01932f5d-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
               "current_hp": 190,
               "current_mp": 9999
            }
         ],
         "combat_logs": [
            {
               "turn": 1,
               "action_type": "CAST_SPELL",
               "damage_dealt": 110,
               "mp_consumed": 20
            }
         ]
      }
   }
}
```

**3. Enemy AI Turn (automatic)**

Server automatically processes enemy turn and returns updated match state.

#### User Actions:

-  Select deck slot (element) → Show available masteries
-  Select mastery → Call Resolve Spell API → Show spell preview
-  Select cast mode (Instant/Charge/Overcharge)
-  Click "Cast Spell" → Call Perform Action API → Update UI
-  AI enemy turn plays → Update UI
-  Repeat until enemy HP = 0
-  Victory detected → ไปหน้า Screen 8.6 (Victory Screen)

---

### 🎉 **Screen 8.6: Victory Screen (Tutorial Complete)**

**จุดประสงค์:** แสดงผลชนะและรางวัล

#### UI Elements:

```
┌─────────────────────────────────────┐
│                                     │
│           🎉 VICTORY! 🎉           │
│                                     │
│      You defeated Training Dummy!   │
│                                     │
│  ╔════════════════════════════════╗│
│  ║  Rewards:                      ║│
│  ║  ✨ 50 EXP                     ║│
│  ║  🎓 Tutorial Completed!        ║│
│  ║                                ║│
│  ║  Character Progress:           ║│
│  ║  Level: 1 → 1                  ║│
│  ║  EXP: 0/100 → 50/100           ║│
│  ║  [██████████░░░░░░] 50%       ║│
│  ╚════════════════════════════════╝│
│                                     │
│  [    Continue    ]                 │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**Get Updated Character Info**

```http
GET /api/v1/characters/1
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": {
      "id": 1,
      "character_name": "FireMage",
      "level": 1,
      "exp": 50,
      "current_hp": 1023,
      "current_mp": 310,
      "tutorial": {
         "current_step": 999,
         "is_completed": true
      }
   }
}
```

#### User Actions:

-  Click "Continue" → ไปหน้า Screen 9 (Character Selection / Home)

---

### 🏠 **Screen 9: Home / Character Selection**

**จุดประสงค์:** Hub หลักของเกม เลือกตัวละครและเมนูต่างๆ

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ☰ Menu              [👤 player123] │
├─────────────────────────────────────┤
│                                     │
│  Your Characters:                   │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  FireMage         Lv.1        │ │
│  │  [Avatar]                     │ │
│  │  HP: 1023/1023  MP: 310/330  │ │
│  │  Primary: S (Solidity)        │ │
│  │  EXP: 50/100 [████░░] 50%    │ │
│  │                               │ │
│  │  [    Select    ]             │ │
│  └───────────────────────────────┘ │
│                                     │
│  [ + Create New Character ]         │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. List All Characters**

```http
GET /api/v1/characters/
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": [
      {
         "id": 1,
         "character_name": "FireMage",
         "level": 1,
         "exp": 50,
         "primary_element_id": 1,
         "current_hp": 1023,
         "current_mp": 310
      }
   ]
}
```

#### User Actions:

-  Click character → ไปหน้า Screen 10 (Character Detail / Main Menu)
-  Click "+ Create New Character" → ไปหน้า Screen 3

---

### 🎯 **Screen 10: Main Menu (Character Selected)**

**จุดประสงค์:** เมนูหลักสำหรับเข้าถึงฟีเจอร์ต่างๆ

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ← Back          FireMage     Lv.1  │
├─────────────────────────────────────┤
│                                     │
│       [Character 3D Model]          │
│                                     │
│  HP: 1023/1023  MP: 310/330         │
│  EXP: [████████░░░░] 50/100         │
│                                     │
│  ┌─────────────┐ ┌─────────────┐   │
│  │   🎮 PVE    │ │   ⚔️  PVP   │   │
│  │   Story     │ │   Battle    │   │
│  └─────────────┘ └─────────────┘   │
│                                     │
│  ┌─────────────┐ ┌─────────────┐   │
│  │   🎴 Deck   │ │  🧪 Fusion  │   │
│  │  Builder    │ │   Craft     │   │
│  └─────────────┘ └─────────────┘   │
│                                     │
│  ┌─────────────┐ ┌─────────────┐   │
│  │ 💼 Inventory│ │  📊 Stats   │   │
│  │   Items     │ │  Profile    │   │
│  └─────────────┘ └─────────────┘   │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**Get Character Details**

```http
GET /api/v1/characters/1
Authorization: Bearer <accessToken>
```

#### User Actions:

-  Click "PVE Story" → ไปหน้า Screen 11 (Stage Selection)
-  Click "PVP Battle" → ไปหน้า PVP (Not implemented yet)
-  Click "Deck Builder" → ไปหน้า Deck Management
-  Click "Fusion" → ไปหน้า Crafting
-  Click "Inventory" → ไปหน้า Inventory
-  Click "Stats" → ไปหน้า Character Stats

---

### 📜 **Screen 11: Stage Selection (PVE Story Mode)**

**จุดประสงค์:** เลือกด่านที่จะเล่น

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ← Back        Story Mode            │
├─────────────────────────────────────┤
│                                     │
│  Realm: Fire Realm                  │
│  Progress: 0/10 Stages              │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  Stage 1: Ember Plains        │ │
│  │  ⭐☆☆                          │ │
│  │  Difficulty: Easy              │ │
│  │  Enemies: 1                    │ │
│  │  Reward: 100 EXP               │ │
│  │                                │ │
│  │  Enemy: Fire Imp               │ │
│  │  • HP: 500                     │ │
│  │  • Level: 2                    │ │
│  │                                │ │
│  │  [    Start Battle    ]        │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  Stage 2: ???                  │ │
│  │  🔒 Locked                     │ │
│  │  (Complete Stage 1 to unlock) │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Get Realms & Stages**

```http
GET /api/v1/pve/realms
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": [
      {
         "id": 1,
         "realm_name": "Fire Realm",
         "stages": [
            {
               "id": 1,
               "stage_name": "Ember Plains",
               "difficulty": 1,
               "reward_exp": 100,
               "enemies": [
                  {
                     "enemy_id": 2,
                     "position": 1
                  }
               ]
            },
            {
               "id": 2,
               "stage_name": "Volcanic Cavern",
               "difficulty": 2,
               "is_locked": true
            }
         ]
      }
   ]
}
```

**2. Get Enemy Details**

```http
GET /api/v1/enemies/
Authorization: Bearer <accessToken>
```

Filter for enemy_id = 2

#### User Actions:

-  Click "Start Battle" → ไปหน้า Screen 12 (Deck Selection)

---

### 🎴 **Screen 12: Deck Selection (Before Battle)**

**จุดประสงค์:** เลือก Deck ที่จะใช้ในการต่อสู้

#### UI Elements:

```
┌─────────────────────────────────────┐
│  ← Back      Choose Your Deck       │
├─────────────────────────────────────┤
│                                     │
│  Stage 1: Ember Plains              │
│  Enemy: Fire Imp (Lv.2)             │
│                                     │
│  Your Decks:                        │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ◉ Starter Deck                │ │
│  │   [🔥][🔥][💧][💧][⚡][⚡]    │ │
│  │   [✨][🌑]                     │ │
│  │   Fire/Water/Lightning         │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ○ Balanced Deck               │ │
│  │   [🔥][💧][🌪️][🌍][⚡][❄️]    │ │
│  │   [✨][🌑]                     │ │
│  │   All Elements                 │ │
│  └───────────────────────────────┘ │
│                                     │
│  [  Manage Decks  ] [  Continue  ]  │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**Get Character's Decks**

```http
GET /api/v1/decks/?character_id=1
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "deck_name": "Starter Deck",
      "slots": [
        { "slot_num": 1, "element_id": 5, "element": { "element_name": "Fire" } },
        { "slot_num": 2, "element_id": 5, "element": { "element_name": "Fire" } },
        // ... 8 slots total
      ]
    },
    {
      "id": 2,
      "deck_name": "Balanced Deck",
      "slots": [...]
    }
  ]
}
```

#### User Actions:

-  Select deck (radio button)
-  Click "Continue" → ไปหน้า Screen 13 (Battle Screen - Stage 1)
-  Click "Manage Decks" → ไปหน้า Deck Management

---

### ⚔️ **Screen 13: Battle Screen (Stage 1 Fight)**

**จุดประสงค์:** ต่อสู้กับศัตรูในด่านที่ 1

#### UI Elements:

```
┌─────────────────────────────────────┐
│  Turn 1      Stage 1: Ember Plains  │
├─────────────────────────────────────┤
│  Enemy: Fire Imp                    │
│  HP: [████████████] 500/500         │
│  Lv.2  🔥 Solidity                  │
│                                     │
│      [Fire Imp Sprite]              │
│                                     │
│  ───────────────────────────────────│
│                                     │
│  You: FireMage  Lv.1                │
│  HP: [████████████] 1023/1023       │
│  MP: [████████████] 310/330         │
│                                     │
│  Deck:                              │
│  [🔥][🔥][💧][💧][⚡][⚡][✨][🌑]  │
│                                     │
│  Actions: (Same as Screen 8.5)      │
│  [Select Element + Mastery + Cast]  │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Create Match (STORY Mode)** ⚠️ NOT IMPLEMENTED YET

```http
POST /api/v1/combat/
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "character_id": 1,
  "match_type": "STORY",
  "deck_id": 1,
  "stage_id": 1
}
```

**Fallback: Use TRAINING Mode**

```http
POST /api/v1/combat/
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "character_id": 1,
  "match_type": "TRAINING",
  "deck_id": 1,
  "training_enemies": [
    { "enemy_id": 2 }
  ]
}
```

**2-3. Same as Screen 8.5 (Resolve Spell + Cast Spell)**

Battle continues until victory or defeat...

---

### 🎉 **Screen 14: Victory Screen (Stage 1 Complete)**

**จุดประสงค์:** แสดงรางวัลหลังชนะด่าน

#### UI Elements:

```
┌─────────────────────────────────────┐
│                                     │
│        🏆 STAGE COMPLETE! 🏆       │
│                                     │
│     Ember Plains - Conquered!       │
│                                     │
│  ╔════════════════════════════════╗│
│  ║  Performance:                  ║│
│  ║  ⭐⭐⭐ Perfect!               ║│
│  ║                                ║│
│  ║  Rewards:                      ║│
│  ║  ✨ 100 EXP                    ║│
│  ║  🎴 Fire Element x2            ║│
│  ║  💎 10 Gems                    ║│
│  ║                                ║│
│  ║  Character Progress:           ║│
│  ║  Level: 1 → 2! 🎉             ║│
│  ║  EXP: 50/100 → 150/200         ║│
│  ║  [████████████░░] 75%         ║│
│  ║                                ║│
│  ║  🆕 Unlocked:                  ║│
│  ║  • Stage 2: Volcanic Cavern   ║│
│  ║  • New Spell: Flame Strike    ║│
│  ╚════════════════════════════════╝│
│                                     │
│  [  Continue  ] [  Replay Stage  ]  │
│                                     │
└─────────────────────────────────────┘
```

#### API Calls:

**1. Get Updated Character** (EXP already granted by server)

```http
GET /api/v1/characters/1
Authorization: Bearer <accessToken>
```

**Response:**

```json
{
   "success": true,
   "data": {
      "id": 1,
      "character_name": "FireMage",
      "level": 2,
      "exp": 150,
      "current_hp": 1033,
      "current_mp": 330
   }
}
```

**2. Get Updated Inventory** (if items rewarded)

```http
GET /api/v1/characters/1/inventory
Authorization: Bearer <accessToken>
```

#### User Actions:

-  Click "Continue" → ไปหน้า Screen 11 (Stage Selection) - Stage 2 unlocked
-  Click "Replay Stage" → Create new match with Stage 1

---

## 🎯 **END OF SCENARIO**

**ผู้เล่นได้ทำสำเร็จ:**

-  ✅ สมัครสมาชิก
-  ✅ สร้างตัวละคร (พร้อมเลือก Element & Mastery)
-  ✅ สร้าง Deck แรก
-  ✅ เรียนรู้การต่อสู้ผ่าน Tutorial
-  ✅ ชนะด่านแรก (Stage 1)
-  ✅ ได้รับรางวัล (EXP, Items)
-  ✅ ปลดล็อคด่านถัดไป

**ผู้เล่นพร้อมที่จะ:**

-  🎮 เล่นด่านต่อไป (Stage 2, 3, ...)
-  🎴 จัด Deck ใหม่
-  🧪 Fusion ธาตุใหม่
-  ⚔️ ท้าทาย PVP (when implemented)

---

## 📊 API Mapping Summary

### By Screen

| Screen | Screen Name         | API Calls                                                                         | Count |
| ------ | ------------------- | --------------------------------------------------------------------------------- | ----- |
| 1      | Welcome             | None                                                                              | 0     |
| 2      | Registration        | POST /players/register                                                            | 1     |
| 2.5    | Login               | POST /players/login<br>GET /players/me<br>GET /characters/                        | 3     |
| 3      | Character Create #1 | None                                                                              | 0     |
| 4      | Character Create #2 | GET /game-data/elements                                                           | 1     |
| 5      | Character Create #3 | GET /game-data/masteries<br>POST /characters/                                     | 2     |
| 6      | Tutorial Intro      | POST /characters/:id/tutorial/skip (optional)                                     | 0-1   |
| 7      | Deck Building       | GET /game-data/elements<br>POST /decks/                                           | 2     |
| 8      | Battle Intro        | GET /enemies/<br>POST /combat/                                                    | 2     |
| 8.5    | Battle (Tutorial)   | GET /combat/resolve-spell<br>POST /combat/:id/actions (multiple)                  | 2-10  |
| 8.6    | Victory (Tutorial)  | GET /characters/:id                                                               | 1     |
| 9      | Home                | GET /characters/                                                                  | 1     |
| 10     | Main Menu           | GET /characters/:id                                                               | 1     |
| 11     | Stage Selection     | GET /pve/realms<br>GET /enemies/                                                  | 2     |
| 12     | Deck Selection      | GET /decks/                                                                       | 1     |
| 13     | Battle (Stage 1)    | POST /combat/<br>GET /combat/resolve-spell<br>POST /combat/:id/actions (multiple) | 2-15  |
| 14     | Victory (Stage 1)   | GET /characters/:id<br>GET /characters/:id/inventory                              | 2     |

**Total API Calls:** ~25-45 calls (depending on battle length)

---

### By Module

| Module             | Endpoints Used                                                                                     | Total Calls |
| ------------------ | -------------------------------------------------------------------------------------------------- | ----------- |
| **Authentication** | POST /players/register<br>POST /players/login<br>GET /players/me                                   | 3-4         |
| **Character**      | POST /characters/<br>GET /characters/<br>GET /characters/:id<br>POST /characters/:id/tutorial/skip | 5-7         |
| **Game Data**      | GET /game-data/elements<br>GET /game-data/masteries                                                | 2-3         |
| **Deck**           | POST /decks/<br>GET /decks/                                                                        | 2-3         |
| **Combat**         | POST /combat/<br>GET /combat/resolve-spell<br>POST /combat/:id/actions                             | 10-25       |
| **Enemy**          | GET /enemies/                                                                                      | 2           |
| **PVE**            | GET /pve/realms                                                                                    | 1           |
| **Inventory**      | GET /characters/:id/inventory                                                                      | 1           |

---

## ⚠️ Error Handling Guide

### Common Error Scenarios

#### 1. Registration Errors

```json
// 409 Conflict - Username exists
{
   "success": false,
   "error": {
      "code": "USERNAME_EXISTS",
      "message": "Username already taken"
   }
}
```

**UI Response:** Show error below username field, suggest alternatives

---

#### 2. Login Errors

```json
// 401 Unauthorized - Wrong password
{
   "success": false,
   "error": {
      "code": "INVALID_CREDENTIALS",
      "message": "Invalid username or password"
   }
}
```

**UI Response:** Show generic error "Invalid credentials" (don't reveal which field is wrong for security)

---

#### 3. Token Expired

```json
// 401 Unauthorized
{
   "success": false,
   "error": {
      "code": "TOKEN_EXPIRED",
      "message": "Access token has expired"
   }
}
```

**UI Response:**

1. Call POST /players/refresh-token with refresh token
2. Retry original request with new access token
3. If refresh fails → Redirect to login

---

#### 4. Character Name Exists

```json
// 409 Conflict
{
   "success": false,
   "error": {
      "code": "CHARACTER_NAME_EXISTS",
      "message": "Character name already exists"
   }
}
```

**UI Response:** Show error, allow user to try different name

---

#### 5. Match Already Active

```json
// 409 Conflict
{
   "success": false,
   "error": {
      "code": "MATCH_ALREADY_ACTIVE",
      "message": "Character already has an active match"
   }
}
```

**UI Response:**

-  Show dialog: "You have an ongoing battle. Resume?"
-  [Resume] → Get active match and go to battle screen
-  [Abandon] → Call abort match API

---

#### 6. Insufficient Resources

```json
// 400 Bad Request
{
   "success": false,
   "error": {
      "code": "INSUFFICIENT_MP",
      "message": "Not enough MP to cast spell"
   }
}
```

**UI Response:** Show in-battle notification, disable cast button

---

### Error Handling Flow

```javascript
// Global API Error Handler
async function callAPI(endpoint, options) {
   try {
      const response = await fetch(endpoint, options);
      const data = await response.json();

      if (!response.ok) {
         if (response.status === 401) {
            // Token expired - try refresh
            const refreshed = await refreshAccessToken();
            if (refreshed) {
               // Retry original request
               return callAPI(endpoint, options);
            } else {
               // Refresh failed - logout
               redirectToLogin();
            }
         }

         // Other errors
         throw new APIError(data.error);
      }

      return data;
   } catch (error) {
      handleError(error);
   }
}
```

---

## 📝 Implementation Notes

### Client-Side State Management

**Recommended State Structure:**

```javascript
{
  auth: {
    accessToken: string,
    refreshToken: string,
    user: { id, username, email }
  },
  character: {
    current: { id, name, level, exp, hp, mp, talents, masteries },
    list: Character[]
  },
  deck: {
    current: { id, name, slots[] },
    list: Deck[]
  },
  battle: {
    match: Match,
    selectedElement: Element,
    selectedMastery: Mastery,
    resolvedSpell: Spell,
    castMode: "INSTANT" | "CHARGE" | "OVERCHARGE"
  },
  gameData: {
    elements: Element[],
    spells: Spell[],
    effects: Effect[],
    masteries: Mastery[],
    enemies: Enemy[]
  }
}
```

---

### Caching Strategy

**Cache on Client:**

-  ✅ Game Data (elements, spells, effects, masteries) - 1 hour
-  ✅ Character list - 5 minutes
-  ✅ Deck list - 5 minutes
-  ❌ Battle state - NO CACHE (always fetch fresh)

**Example:**

```javascript
// Cache game data after first fetch
const CACHE_DURATION = 3600000; // 1 hour
let gameDataCache = {
   elements: null,
   timestamp: 0,
};

async function getElements() {
   const now = Date.now();
   if (
      gameDataCache.elements &&
      now - gameDataCache.timestamp < CACHE_DURATION
   ) {
      return gameDataCache.elements;
   }

   const data = await callAPI("/api/v1/game-data/elements");
   gameDataCache.elements = data;
   gameDataCache.timestamp = now;
   return data;
}
```

---

### Performance Optimization

**Parallel API Calls:**

```javascript
// Screen 2.5 Login - Fetch multiple data in parallel
async function afterLogin(accessToken) {
   const [profile, characters, elements, masteries] = await Promise.all([
      callAPI("/api/v1/players/me", { token: accessToken }),
      callAPI("/api/v1/characters/", { token: accessToken }),
      callAPI("/api/v1/game-data/elements", { token: accessToken }),
      callAPI("/api/v1/game-data/masteries", { token: accessToken }),
   ]);

   // Determine next screen based on characters
   if (characters.data.length > 0) {
      navigateTo("/character-selection");
   } else {
      navigateTo("/character-create");
   }
}
```

---

### Offline Handling

**Show appropriate message:**

```javascript
if (!navigator.onLine) {
   showNotification("You are offline. Please check your connection.");
   disableBattleActions();
}

window.addEventListener("online", () => {
   showNotification("Connection restored!");
   enableBattleActions();
   syncPendingActions();
});
```

---

## 🎯 Next Steps

**After completing Stage 1, players can:**

1. **Continue Story Mode**

   -  Play Stage 2, 3, 4...
   -  Unlock new realms
   -  Face tougher enemies

2. **Manage Decks**

   -  Create multiple decks
   -  Edit existing decks
   -  Strategize for different enemies

3. **Fusion/Crafting**

   -  Combine T1 elements → T2
   -  Craft advanced elements
   -  Unlock powerful spells

4. **Character Progression**

   -  Level up (gain EXP)
   -  Allocate talent points (when implemented)
   -  Increase mastery levels

5. **PVP (When Implemented)**
   -  Challenge other players
   -  Climb rankings
   -  Earn PVP rewards

---

**Last Updated:** November 1, 2025  
**Scenario Version:** 1.0  
**For Questions:** Contact nipon.k
