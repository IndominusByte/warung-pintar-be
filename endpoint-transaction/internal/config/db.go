package config

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func DBConnect(cfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.Open(cfg.Database.Driver, cfg.Database.FollowerDsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// setting db
	maxIdleConns, err := strconv.Atoi(cfg.Database.MaxIdleConns)
	if err != nil {
		return nil, err
	}
	maxOpenConns, err := strconv.Atoi(cfg.Database.MaxOpenConns)
	if err != nil {
		return nil, err
	}
	connMaxIdletime, err := time.ParseDuration(cfg.Database.ConnMaxIdletime)
	if err != nil {
		return nil, err
	}
	connMaxLifetime, err := time.ParseDuration(cfg.Database.ConnMaxLifetime)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxIdleTime(connMaxIdletime)
	db.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

func RedisConnect(cfg *Config) (*redis.Pool, error) {
	maxIdleConn, err := strconv.Atoi(cfg.Redis.MaxIdleConn)
	if err != nil {
		return nil, err
	}
	maxActiveConn, err := strconv.Atoi(cfg.Redis.MaxActiveConn)
	if err != nil {
		return nil, err
	}
	timeout, err := time.ParseDuration(cfg.Redis.Timeout)
	if err != nil {
		return nil, err
	}

	pool := &redis.Pool{
		MaxIdle:     maxIdleConn,
		MaxActive:   maxActiveConn,
		IdleTimeout: timeout,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.Redis.Address)
		},
	}

	// check if redis connected
	conn := pool.Get()
	defer conn.Close()

	_, err = conn.Do("PING")
	if err != nil {
		return nil, err
	}

	return pool, nil
}
