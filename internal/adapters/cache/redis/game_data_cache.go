package redis

import (
	"context"
	"encoding/json"
	"sage-of-elements-backend/internal/modules/game_data"
	"time"

	"github.com/redis/go-redis/v9"
)

const masterDataKey = "master_data:v1" // <-- เราใส่เวอร์ชันไว้ด้วย เผื่ออนาคตมีการเปลี่ยนโครงสร้างข้อมูล

type gameDataCacheRepository struct {
	client *redis.Client
}

// NewGameDataCacheRepository คือฟังก์ชันสำหรับสร้าง Cache Repository
func NewGameDataCacheRepository(client *redis.Client) game_data.CacheRepository {
	return &gameDataCacheRepository{client: client}
}

// GetMasterData พยายามดึงข้อมูล Master Data จาก Redis Cache
func (r *gameDataCacheRepository) GetMasterData() (*game_data.MasterDataResponse, error) {
	// ใช้ client.Get เพื่อดึงข้อมูลจาก key "master_data:v1"
	val, err := r.client.Get(context.Background(), masterDataKey).Result()

	// นี่คือหัวใจสำคัญ!
	if err == redis.Nil {
		return nil, nil // ถ้าเจอ error 'redis.Nil' แปลว่า "ไม่เจอ key" นี่คือ Cache Miss ซึ่งเป็นเรื่องปกติ
	} else if err != nil {
		return nil, err // ถ้าเป็น Error อื่นๆ (เช่น Redis down) ให้ส่ง Error กลับไป
	}

	// ถ้าเจอข้อมูล... ก็ทำการ Unmarshal JSON string กลับมาเป็น Struct
	var data game_data.MasterDataResponse
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// SetMasterData ทำการบันทึก Master Data ลงใน Redis Cache
func (r *gameDataCacheRepository) SetMasterData(data *game_data.MasterDataResponse, expiration time.Duration) error {
	// 1. แปลง Struct ของเราให้กลายเป็น JSON string
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 2. ใช้ client.Set เพื่อบันทึกข้อมูลลงใน Redis พร้อมตั้งเวลาหมดอายุ
	return r.client.Set(context.Background(), masterDataKey, bytes, expiration).Err()
}
