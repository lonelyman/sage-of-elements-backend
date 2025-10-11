package game_data

import "time"

// CacheRepository คือ "สัญญา" สำหรับการติดต่อกับ Cache
type CacheRepository interface {
	GetMasterData() (*MasterDataResponse, error)
	SetMasterData(data *MasterDataResponse, expiration time.Duration) error
}
