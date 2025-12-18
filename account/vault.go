package account

import (
	"encoding/json"
	"strings"
	"time"
)

type Vault struct {
	Accounts  []Account `json:"accounts"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewVault() *Vault {
	return &Vault{
		Accounts:  []Account{},
		UpdatedAt: time.Now(),
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

func (vault *Vault) FindAccountByURL(urlString string) []Account {
	var accounts []Account
	for _, account := range vault.Accounts {
		isMatched := strings.Contains(account.URL, urlString)
		if isMatched {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func (vault *Vault) AddAccount(acc Account) {
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
