package model

import (
	"sync"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/doug-martin/goqu/v9"
)

type ServerConfig struct {
	Id                  uint64 `db:"id" goqu:"skipinsert"`
	MaxMessageSize      int    `db:"max_message_size"`
	UserTokenExpiration int    `db:"user_token_expiration"` // in minutes
	MaxFileSize         int64  `db:"max_file_size"`         // in megabytes (MB)
	NoticeMessage       string `db:"notice_message"`
	Lockdown            bool   `db:"lockdown"`
}

var (
	cachedConfig     *ServerConfig
	cachedConfigLock sync.RWMutex
)

func GetServerConfig() ServerConfig {
	cachedConfigLock.RLock()
	if cachedConfig != nil {
		cfg := *cachedConfig
		cachedConfigLock.RUnlock()
		return cfg
	}
	cachedConfigLock.RUnlock()

	return ReloadServerConfig()
}

func ReloadServerConfig() ServerConfig {
	cachedConfigLock.Lock()
	defer cachedConfigLock.Unlock()

	cfg := ServerConfig{
		Id:                  1,
		MaxMessageSize:      8000,
		UserTokenExpiration: 180,
		MaxFileSize:         10,
		NoticeMessage:       "",
		Lockdown:            false,
	}

	if Sqlx != nil {
		var dbCfg ServerConfig
		err := First(TableServerConfig, goqu.Ex{"id": 1}, &dbCfg)
		if err != nil {
			_ = Insert(TableServerConfig, &cfg)
		} else {
			if dbCfg.MaxFileSize > 100000 {
				dbCfg.MaxFileSize = dbCfg.MaxFileSize / (1024 * 1024)
			}
			cfg = dbCfg
		}
	}

	environment.ChatMessageMaxLength = cfg.MaxMessageSize
	environment.UserAuthTokenExpiration = time.Duration(cfg.UserTokenExpiration) * time.Minute

	cachedConfig = &cfg
	return cfg
}

func (sc *ServerConfig) Update() error {
	cachedConfigLock.Lock()
	defer cachedConfigLock.Unlock()

	sc.Id = 1
	err := Update(TableServerConfig, sc, goqu.Ex{"id": 1})
	if err == nil {
		environment.ChatMessageMaxLength = sc.MaxMessageSize
		environment.UserAuthTokenExpiration = time.Duration(sc.UserTokenExpiration) * time.Minute
		cachedConfig = sc
	}
	return err
}

func (sc ServerConfig) GetUserTokenExpirationDuration() time.Duration {
	if sc.UserTokenExpiration <= 0 {
		return 180 * time.Minute
	}
	return time.Duration(sc.UserTokenExpiration) * time.Minute
}

func (sc ServerConfig) MaxFileSizeBytes() int64 {
	if sc.MaxFileSize <= 0 {
		return 0
	}
	return sc.MaxFileSize * 1024 * 1024
}
