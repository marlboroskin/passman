package account

import (
	"encoding/json"
	"menedger_paroley/crypto"
	"menedger_paroley/output"
	"strings"
	"sync"
	"time"
)

type Db interface {
	Read() ([]byte, error)
	Write([]byte) error
}

type Vault struct {
	Accounts     []Account `json:"accounts"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Verification string    `json:"verification"`
}

type VaultWithDb struct {
	Data Vault
	Db   Db
	sync.RWMutex
}

func NewVault(db Db) *VaultWithDb {
	file, err := db.Read()
	if err != nil {
		return &VaultWithDb{
			Data: Vault{
				Accounts:  []Account{},
				UpdatedAt: time.Now(),
			},
			Db: db,
		}
	}

	var vault Vault
	err = json.Unmarshal(file, &vault)
	if err != nil {
		output.PrintError("Ошибка чтения хранилища: неверный формат данных")
		return &VaultWithDb{
			Data: Vault{
				Accounts:  []Account{},
				UpdatedAt: time.Now(),
			},
			Db: db,
		}
	}

	return &VaultWithDb{
		Data: vault,
		Db:   db,
	}
}

func (v *VaultWithDb) DeleteAccountByURL(url string) bool {
	v.Lock()
	defer v.Unlock()
	var accounts []Account
	isDeleted := false
	for _, acc := range v.Data.Accounts {
		isMatched := strings.Contains(strings.ToLower(acc.URL), strings.ToLower(url))
		if !isMatched {
			accounts = append(accounts, acc)
			continue
		}
		isDeleted = true
	}
	v.Data.Accounts = accounts
	v.Data.UpdatedAt = time.Now()
	return isDeleted
}

func (v *VaultWithDb) AddAccount(acc Account) {
	v.Lock()
	defer v.Unlock()
	v.Data.Accounts = append(v.Data.Accounts, acc)
	v.Data.UpdatedAt = time.Now()
}

func (vault *VaultWithDb) ToBytes() ([]byte, error) {
	vault.RLock()
	defer vault.RUnlock()
	data, err := json.MarshalIndent(&vault.Data, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (v *VaultWithDb) FindAccount(query string) []Account {
	v.RLock()
	defer v.RUnlock()
	var accounts []Account
	q := strings.ToLower(query)

	for _, acc := range v.Data.Accounts {
		if strings.Contains(strings.ToLower(acc.Name), q) ||
			strings.Contains(strings.ToLower(acc.Login), q) ||
			strings.Contains(strings.ToLower(acc.URL), q) {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

func (vault *VaultWithDb) Save() error {
	vault.RLock()
	data, err := json.MarshalIndent(&vault.Data, "", "  ")
	vault.RUnlock()
	if err != nil {
		output.PrintError(err)
		return err
	}

	// 🔐 Шифруем данные
	password := "master" // ← НЕЛЬЗЯ ЖЁСТКО! Нужно получать из вне
	encrypted, err := crypto.Encrypt(data, []byte(password))
	if err != nil {
		output.PrintError("Ошибка шифрования")
		return err
	}

	return vault.Db.Write(encrypted)
}
