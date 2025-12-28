package config

import (
	"bufio"
	"fmt"
	"menedger_paroley/account"
	"menedger_paroley/cloud"
	"menedger_paroley/files"
	"os"
	"strings"

	"github.com/fatih/color"
)

func ChooseStorage() account.Db {
	color.Cyan("1. Локальное хранилище (data.enc)")
	color.Cyan("2. Облако (WebDAV)")
	choice := promptInt("👉 Введите номер хранилища (1 или 2): ")

	switch choice {
	case 1:
		color.Green("Выбрано: локальное хранилище")
		return files.NewJsonDb("data.enc")
	case 2:
		color.Green("Выбрано: облако")
		return configureCloud()
	default:
		color.Red("Неверный выбор. Используем локальное хранилище.")
		return files.NewJsonDb("data.enc")
	}
}

func configureCloud() account.Db {
	url := prompt("URL: ")
	user := prompt("Логин: ")
	pass := PromptPassword("Пароль: ")
	return cloud.NewCloudDb(url, user, pass)
}

func prompt(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewReader(os.Stdin)
	text, _ := scanner.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptInt(msg string) int {
	for {
		input := prompt(msg)
		var n int
		_, err := fmt.Sscanf(input, "%d", &n)
		if err == nil && (n == 1 || n == 2) {
			return n
		}
		color.Red("Введите 1 или 2")
	}
}

func PromptPassword(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewReader(os.Stdin)
	password, _ := scanner.ReadString('\n')
	return strings.TrimSpace(password)
}
