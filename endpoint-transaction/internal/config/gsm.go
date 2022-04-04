package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/creent-production/cdk-go/encryption"
)

type gsmData struct {
	EncryptionKey  string `json:"encryption_key"`
	SecretKey      string `json:"secret_key"`
	PgTalkUser     string `json:"pg_talk_user"`
	PgTalkPassword string `json:"pg_talk_password"`
}

func (cfg *Config) loadFromGsm() error {
	filename := fmt.Sprintf("/app/conf/data.%s.json", os.Getenv("BACKEND_STAGE"))
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var data gsmData
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return err
	}

	// decode data
	cdn := &encryption.Credentials{Key: []byte(data.EncryptionKey)}

	secretkey, secretkeyerr := cdn.Decrypt(data.SecretKey)
	if secretkeyerr != nil {
		return err
	}

	pgtalkuser, pgtalkusererr := cdn.Decrypt(data.PgTalkUser)
	if pgtalkusererr != nil {
		return err
	}

	pgtalkpassword, pgtalkpassworderr := cdn.Decrypt(data.PgTalkPassword)
	if pgtalkpassworderr != nil {
		return pgtalkpassworderr
	}

	cfg.JWT.SecretKey = string(secretkey)
	cfg.Database.MasterDsn = fmt.Sprintf(cfg.Database.MasterDsnNoCred, pgtalkuser, pgtalkpassword)
	cfg.Database.FollowerDsn = fmt.Sprintf(cfg.Database.FollowerDsnNoCred, pgtalkuser, pgtalkpassword)

	return nil
}
