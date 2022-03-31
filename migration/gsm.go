package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/creent-production/cdk-go/encryption"
)

type gsmData struct {
	Driver          string `json:"driver"`
	EncryptionKey   string `json:"encryption_key"`
	PgTalkUser      string `json:"pg_talk_user"`
	PgTalkPassword  string `json:"pg_talk_password"`
	MasterDsn       string
	MasterDsnNoCred string `json:"master_dsn_no_cred"`
	FileMigration   string `json:"file_migration"`
}

func (gsm *gsmData) loadFromGsm() error {
	filename := fmt.Sprintf("./conf/data.%s.json", os.Getenv("BACKEND_STAGE"))
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(gsm)
	if err != nil {
		return err
	}

	// decode data
	cdn := &encryption.Credentials{Key: []byte(gsm.EncryptionKey)}
	pgtalkuser, pgtalkusererr := cdn.Decrypt(gsm.PgTalkUser)
	if pgtalkusererr != nil {
		return err
	}

	pgtalkpassword, pgtalkpassworderr := cdn.Decrypt(gsm.PgTalkPassword)
	if pgtalkpassworderr != nil {
		return pgtalkpassworderr
	}

	gsm.MasterDsn = fmt.Sprintf(gsm.MasterDsnNoCred, pgtalkuser, pgtalkpassword)

	return nil
}
