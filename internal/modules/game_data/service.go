package game_data

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/applogger"
	"time"

	"golang.org/x/sync/errgroup"
)

// MasterDataResponse คือ DTO สำหรับ API game-data
type MasterDataResponse struct {
	Elements    []domain.Element    `json:"elements"`
	Masteries   []domain.Mastery    `json:"masteries"`
	Recipes     []domain.Recipe     `json:"recipes"`
	Spells      []domain.Spell      `json:"spells"`
	GameConfigs []domain.GameConfig `json:"gameConfigs"`
}

// Service คือ "สัญญา" สำหรับ Business Logic ของ Game Data
type Service interface {
	GetMasterData() (*MasterDataResponse, error)
}

// gameDataService คือ struct ที่จะเก็บ Logic การทำงานจริง
type gameDataService struct {
	appLogger    applogger.Logger
	gameDataRepo GameDataRepository
	cacheRepo    CacheRepository
}

// NewService คือฟังก์ชันสำหรับสร้าง Service ขึ้นมาใช้งาน
func NewGameDataService(appLogger applogger.Logger, gameDataRepo GameDataRepository, cacheRepo CacheRepository) Service {
	return &gameDataService{
		appLogger:    appLogger,
		gameDataRepo: gameDataRepo,
		cacheRepo:    cacheRepo,
	}
}

// GetMasterData คือฟังก์ชันที่เราจะมาใส่ Logic กันในขั้นตอนต่อไป!
func (s *gameDataService) GetMasterData() (*MasterDataResponse, error) {
	// --- ขั้นตอนที่ 1: ถาม Cache ก่อนเสมอ! ---
	cachedData, err := s.cacheRepo.GetMasterData()
	if err == nil && cachedData != nil {
		s.appLogger.Success("Cache Hit! Found master data in Redis.")
		return cachedData, nil
	}

	s.appLogger.Warn("Cache Miss for master data. Fetching from database...")

	// --- ขั้นตอนที่ 2: ถ้าไม่มีใน Cache... ก็ไปถาม Database ---
	var g errgroup.Group

	var elements []domain.Element
	var masteries []domain.Mastery
	var recipes []domain.Recipe
	var spells []domain.Spell
	var gameConfigs []domain.GameConfig

	// สั่งให้ Goroutine ที่ 1 ไปดึงข้อมูล Elements
	g.Go(func() error {
		var err error
		elements, err = s.gameDataRepo.FindAllElements()
		return err
	})

	// สั่งให้ Goroutine ที่ 2 ไปดึงข้อมูล Masteries
	g.Go(func() error {
		var err error
		masteries, err = s.gameDataRepo.FindAllMasteries()
		return err
	})

	// สั่งให้ Goroutine ที่ 3 ไปดึงข้อมูล Recipes
	g.Go(func() error {
		var err error
		recipes, err = s.gameDataRepo.FindAllRecipes()
		return err
	})

	// สั่งให้ Goroutine ที่ 4 ไปดึงข้อมูล Spells
	g.Go(func() error {
		var err error
		spells, err = s.gameDataRepo.FindAllSpells()
		return err
	})

	g.Go(func() error {
		var err error
		// เรียกใช้ฟังก์ชันใหม่ที่เราสร้างไว้ใน Repository
		gameConfigs, err = s.gameDataRepo.FindAllGameConfigs()
		return err
	})

	// รอให้ Goroutine ทั้งหมดทำงานเสร็จ
	if err := g.Wait(); err != nil {
		s.appLogger.Error("Failed to fetch master data from database", err)
		return nil, err
	}
	s.appLogger.Info("Successfully fetched all master data from database.")

	// --- ขั้นตอนที่ 3: ประกอบร่างข้อมูล ---
	response := &MasterDataResponse{
		Elements:    elements,
		Masteries:   masteries,
		Recipes:     recipes,
		Spells:      spells,
		GameConfigs: gameConfigs,
	}

	// --- ขั้นตอนที่ 4: ✨ ก่อนจะส่งกลับ... เอาไปเก็บใน Cache ก่อน! ✨ ---
	err = s.cacheRepo.SetMasterData(response, time.Hour*24) // Cache ไว้ 24 ชั่วโมง
	if err != nil {
		// แค่ log error ไว้ก็พอ ไม่ต้องทำให้ request ทั้งหมดล้มเหลว
		s.appLogger.Warn("Failed to set master data cache", "error", err.Error())
	} else {
		s.appLogger.Info("Master data has been saved to cache.")
	}

	return response, nil
}
