package main

import (
	"menedger_paroley/internal/app"
	"menedger_paroley/internal/auth"
	"menedger_paroley/internal/config"
	"menedger_paroley/output"

	"github.com/fatih/color"
)

func main() {
	color.Cyan("🔒 Менеджер паролей")

	db := config.ChooseStorage()
	password := app.PromptPassword("Введите мастер-пароль: ")

	vault := app.LoadVault(db, password) // ← передаём пароль

	if len(vault.Data.Accounts) == 0 {
		err := auth.SetMasterPassword(password)
		if err != nil {
			output.PrintError("Ошибка сохранения: " + err.Error())
			return
		}
		color.Green("Мастер-пароль установлен")
	} else {
		if !auth.Verify(password) {
			output.PrintError("❌ Неверный пароль")
			return
		}
	}

	app.RunCLI(vault, password) // ← передаём пароль
}
