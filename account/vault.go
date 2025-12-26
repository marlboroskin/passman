package account

import (
	"encoding/json"
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
	Mu           sync.RWMutex
}

type VaultWithDb struct {
	Vault
	Db Db
}

func NewVault(db Db) *VaultWithDb {
	file, err := db.Read()
	if err != nil {
		return &VaultWithDb{
			Vault: Vault{
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
			Vault: Vault{
				Accounts:  []Account{},
				UpdatedAt: time.Now(),
			},
			Db: db,
		}
	}

	return &VaultWithDb{
		Vault: vault,
		Db:    db,
	}
}

func (vault *Vault) DeleteAccountByURL(url string) bool {
	var accounts []Account
	isDeleted := false
	for _, account := range vault.Accounts {
		isMatched := strings.Contains(strings.ToLower(account.URL), strings.ToLower(url))
		if !isMatched {
			accounts = append(accounts, account)
			continue
		}
		isDeleted = true
	}
	vault.Accounts = accounts
	vault.UpdatedAt = time.Now()
	return isDeleted
}

func (vault *Vault) AddAccount(acc Account) {
	vault.Mu.Lock()
	defer vault.Mu.Unlock()
	vault.Accounts = append(vault.Accounts, acc)
	vault.UpdatedAt = time.Now()
}

func (vault *Vault) ToBytes() ([]byte, error) {
	data, err := json.MarshalIndent(vault, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (vault *Vault) FindAccount(query string) []Account {
	vault.Mu.RLock()
	defer vault.Mu.RUnlock()
	var accounts []Account
	q := strings.ToLower(query)

	for _, acc := range vault.Accounts {
		if strings.Contains(strings.ToLower(acc.Name), q) ||
			strings.Contains(strings.ToLower(acc.Login), q) ||
			strings.Contains(strings.ToLower(acc.URL), q) {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

func (vault *VaultWithDb) Save() error {
	data, err := vault.ToBytes()
	if err != nil {
		output.PrintError(err)
		return err
	}
	err = vault.Db.Write(data)
	if err != nil {
		output.PrintError(err)
		return err
	}
	return nil
}

func (v *Vault) Lock() {
	v.Mu.Lock()
}

func (v *Vault) Unlock() {
	v.Mu.Unlock()
}

func (v *Vault) RLock() {
	v.Mu.RLock()
}

func (v *Vault) RUnlock() {
	v.Mu.RUnlock()
}
