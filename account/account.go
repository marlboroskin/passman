package account

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"time"
)

type Account struct {
	Name      string    `json:"name"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var LetterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (acc Account) Output() {
    fmt.Printf("Имя: %s\n", acc.Name)
    fmt.Printf("Логин: %s\n", acc.Login)
    fmt.Printf("Пароль: %s\n", maskPassword(acc.Password)) // ← маскировка
    fmt.Printf("URL: %s\n", acc.URL)
    fmt.Printf("Создан: %s\n", acc.CreatedAt.Format("02.01.2006"))
    fmt.Println("---")
}

func maskPassword(p string) string {
    if len(p) <= 4 {
        return "****"
    }
    return p[:2] + "****" + p[len(p)-2:]
}


func (acc *Account) GeneratePassword(n int) {
	res := make([]rune, n)
	for i := range res {
		res[i] = LetterRunes[rand.IntN(len(LetterRunes))]
	}
	acc.Password = string(res)
}

func NewAccount(name, login, password, urlString string) (*Account, error) {
	if login == "" {
		return nil, errors.New("некорректный логин")
	}
	_, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, errors.New("некорректный URL")
	}
	newAcc := &Account{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Login:     login,
		Password:  password,
		URL:       urlString,
	}
	if password == "" {
		newAcc.GeneratePassword(12)
	}
	return newAcc, nil
}


