package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/player" // Import interface จาก module

	"gorm.io/gorm"
)

// playerRepository คือ struct ที่เก็บการเชื่อมต่อ Database และ implement PlayerRepository interface
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository คือฟังก์ชันสำหรับสร้าง Repository ขึ้นมาใช้งาน (Constructor)
func NewPlayerRepository(db *gorm.DB) player.PlayerRepository {
	return &playerRepository{
		db: db,
	}
}

// Save ทำหน้าที่บันทึกผู้เล่นใหม่ลงในฐานข้อมูล (INSERT a new player)
func (r *playerRepository) Save(player *domain.Player) (*domain.Player, error) {
	// ใช้คำสั่ง .Create ของ GORM เพื่อ INSERT ข้อมูล
	// GORM จะจัดการเรื่องการใส่ ID และ CreatedAt ให้โดยอัตโนมัติ
	result := r.db.Create(player)
	if result.Error != nil {
		return nil, result.Error
	}

	return player, nil
}

// FindByEmail ค้นหาผู้เล่นด้วย Email
func (r *playerRepository) FindByEmail(email string) (*domain.Player, error) {
	var player domain.Player
	result := r.db.Where("email = ?", email).First(&player)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ไม่เจอ ถือว่าปกติ
		}
		return nil, result.Error
	}
	return &player, nil
}

// FindByUsername ทำหน้าที่ค้นหาผู้เล่นด้วย Username (SELECT ... WHERE)
func (r *playerRepository) FindByUsername(username string) (*domain.Player, error) {
	var player domain.Player
	// ใช้ .Where และ .First เพื่อ SELECT ข้อมูล
	result := r.db.Where("username = ?", username).First(&player)

	// ตรวจสอบ Error: GORM จะ return 'ErrRecordNotFound' ถ้าหาไม่เจอ
	if result.Error != nil {
		// ถ้าเป็น Error "หาไม่เจอ" เราไม่ถือว่าเป็น Error ของโปรแกรม แต่คือ "ไม่พบข้อมูล"
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		// แต่ถ้าเป็น Error อื่นๆ (เช่น DB down) ให้ส่ง Error กลับไป
		return nil, result.Error
	}

	return &player, nil
}

func (r *playerRepository) FindByID(id uint) (*domain.Player, error) {
	var player domain.Player
	// ใช้ .First ของ GORM ซึ่งจะค้นหาด้วย Primary Key โดยอัตโนมัติ
	result := r.db.First(&player, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ไม่พบข้อมูล
		}
		return nil, result.Error // Error อื่นๆ
	}

	return &player, nil
}

// FindAuthByPlayerIDAndProvider ค้นหาวิธีการยืนยันตัวตนของผู้เล่น
func (r *playerRepository) FindAuthByPlayerIDAndProvider(playerID uint, provider string) (*domain.PlayerAuth, error) {
	var playerAuth domain.PlayerAuth
	result := r.db.Where("player_id = ? AND provider = ?", playerID, provider).First(&playerAuth)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ไม่เจอ ถือว่าปกติ
		}
		return nil, result.Error
	}
	return &playerAuth, nil
}

// CreateWithAuth สร้าง Player และ PlayerAuth พร้อมกันใน Transaction เดียว
func (r *playerRepository) CreateWithAuth(player *domain.Player, auth *domain.PlayerAuth) (*domain.Player, error) {
	// ใช้ Transaction เพื่อให้แน่ใจว่าถ้ามีอะไรผิดพลาดระหว่างทาง... จะไม่มีข้อมูลถูกสร้างขึ้นเลย
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. สร้าง Player ก่อน
		if err := tx.Create(player).Error; err != nil {
			return err
		}

		// 2. เอา ID ของ Player ที่เพิ่งสร้าง ไปใส่ใน PlayerAuth
		auth.PlayerID = player.ID

		// 3. สร้าง PlayerAuth
		if err := tx.Create(auth).Error; err != nil {
			return err
		}

		// ถ้าทุกอย่างสำเร็จ ให้ commit transaction
		return nil
	})

	if err != nil {
		return nil, err
	}

	return player, nil
}

// Update อัปเดตข้อมูล Player
func (r *playerRepository) Update(player *domain.Player) (*domain.Player, error) {
	result := r.db.Save(player)
	if result.Error != nil {
		return nil, result.Error
	}
	return player, nil
}

// UpdateAuth อัปเดตข้อมูล PlayerAuth
func (r *playerRepository) UpdateAuth(auth *domain.PlayerAuth) (*domain.PlayerAuth, error) {
	result := r.db.Save(auth)
	if result.Error != nil {
		return nil, result.Error
	}
	return auth, nil
}

// FindAuthByRefreshToken ค้นหา PlayerAuth record ด้วย Refresh Token ที่กำหนด
func (r *playerRepository) FindAuthByRefreshToken(token string) (*domain.PlayerAuth, error) {
	var playerAuth domain.PlayerAuth
	// ใช้ .Where เพื่อค้นหาแถวที่มี refresh_token ตรงกับที่ส่งมา
	result := r.db.Where("refresh_token = ?", token).First(&playerAuth)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ไม่พบ Token นี้ในระบบ, ถือว่าปกติ
		}
		return nil, result.Error // Error อื่นๆ
	}

	return &playerAuth, nil
}
