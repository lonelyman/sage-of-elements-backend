package appconfig

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config คือ struct หลักที่เก็บทุกอย่าง
type Config struct {
	App      AppConfig    `mapstructure:"app"`
	Server   ServerConfig `mapstructure:"server"`
	Postgres PostgresDbs  `mapstructure:"postgres"`
	Auth     AuthConfig   `mapstructure:"auth"`
	Redis    RedisConfig  `mapstructure:"redis"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type ServerConfig struct {
	Mode     string `mapstructure:"mode"`
	AppPort  string `mapstructure:"appport"`  // ✨ ชัดเจน! นี่คือพอร์ตของ App ข้างใน
	HostPort string `mapstructure:"hostport"` // ✨ ชัดเจน! นี่คือพอร์ตบน Host ข้างนอก
}

type PostgresDbs struct {
	Primary PostgresConfig `mapstructure:"primary"`
	Logs    PostgresConfig `mapstructure:"logs"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type AuthConfig struct {
	JWTAccessSecret  string `mapstructure:"jwtAccessSecret"`
	JWTRefreshSecret string `mapstructure:"jwtRefreshSecret"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// BuildDSN สร้าง DSN string
func (p PostgresConfig) BuildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode)
}

// LoadConfig โหลด Config จากไฟล์และ Env Var
func LoadConfig() (*Config, error) {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	// ⭐️ เวทมนตร์อยู่ที่นี่! ⭐️
	// บอกให้ Viper อ่าน Env Var มาทับค่าในไฟล์ .yml ได้โดยอัตโนมัติ
	// และให้มันเข้าใจรูปแบบ POSTGRES_PRIMARY_HOST
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// อ่านไฟล์ config.yml (เป็นค่าเริ่มต้น)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Info: No config file found, using environment variables only.")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}
