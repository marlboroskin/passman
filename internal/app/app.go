package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"menedger_paroley/account"
	"menedger_paroley/crypto"
	"menedger_paroley/files"
	"menedger_paroley/output"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
)

var (
	clearTimer *time.Timer
	Mu         sync.Mutex
)

func RunCLI(vault *account.VaultWithDb, password string) {
	for {
		showMenu()
		choice := prompt("Выберите: ")

		switch choice {
		case "1":
			createAccount(vault, password)
		case "2":
			findAccount(vault)
		case "3":
			deleteAccount(vault, password)
		case "4":
			color.Green("Выход...")
			return
		case "5":
			generatePassword()
		case "6":
			copyPassword(vault)
		case "7":
			backupVault(vault)
		case "8":
			restoreFromBackup(vault)
		default:
			output.PrintError("Неверный выбор")
		}
	}
}

func showMenu() {
	color.Cyan("\n__Менеджер паролей__")
	color.White("1. Создать аккаунт")
	color.White("2. Найти аккаунт")
	color.White("3. Удалить аккаунт")
	color.White("4. Выход")
	color.White("5. Сгенерировать пароль")
	color.White("6. Скопировать пароль")
	color.White("7. Создать резервную копию")
	color.White("8. Восстановить из бэкапа")
}

func prompt(prompt string) string {
	print(prompt)
	scanner := bufio.NewReader(os.Stdin)
	text, _ := scanner.ReadString('\n')
	return strings.TrimSpace(text)
}

func PromptPassword(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewReader(os.Stdin)
	password, _ := scanner.ReadString('\n')
	return strings.TrimSpace(password)
}

func createAccount(vault *account.VaultWithDb, password string) {
	name := prompt("Имя: ")
	login := prompt("Логин: ")
	pass := prompt("Пароль (Enter — сгенерировать): ")
	url := prompt("URL: ")

	if pass == "" {
		pass = generateRandomPassword(12)
	}

	acc, err := account.NewAccount(name, login, pass, url)
	if err != nil {
		output.PrintError(err)
		return
	}

	vault.AddAccount(*acc)
	err = SaveEncrypted(vault, password)
	if err != nil {
		output.PrintError("Ошибка сохранения")
	}
	color.Green("Аккаунт добавлен")
}

func findAccount(vault *account.VaultWithDb) {
	query := prompt("Поиск: ")
	accounts := vault.FindAccount(query) // ← не Data.FindAccount
	if len(accounts) == 0 {
		output.PrintError("Не найдено")
		return
	}
	for _, acc := range accounts {
		acc.Output()
	}
}

func deleteAccount(vault *account.VaultWithDb, password string) {
	url := prompt("Частичный URL для удаления: ")
	if vault.DeleteAccountByURL(url) {
		err := SaveEncrypted(vault, password)
		if err != nil {
			output.PrintError("Ошибка сохранения")
		}
		color.Green("Удалено")
	} else {
		output.PrintError("Не найдено")
	}
}

func generatePassword() {
	n := 12
	length := prompt("Длина (8–128): ")
	if _, err := fmt.Sscanf(length, "%d", &n); err != nil || n < 8 || n > 128 {
		n = 12
	}
	color.Green("Сгенерировано: %s", generateRandomPassword(n))
}

func generateRandomPassword(n int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"
	runes := []rune(chars)
	res := make([]rune, n)
	for i := range res {
		res[i] = runes[rand.IntN(len(runes))]
	}
	return string(res)
}

func copyPassword(vault *account.VaultWithDb) {
	query := prompt("Поиск: ")
	accounts := vault.FindAccount(query)
	if len(accounts) == 0 {
		output.PrintError("Не найдено")
		return
	}

	var acc account.Account
	if len(accounts) == 1 {
		acc = accounts[0]
	} else {
		for i, a := range accounts {
			color.White("%d. %s (%s)", i+1, a.Name, a.Login)
		}
		idx := prompt("Выберите номер: ")
		var n int
		fmt.Sscanf(idx, "%d", &n)
		if n < 1 || n > len(accounts) {
			output.PrintError("Неверный номер")
			return
		}
		acc = accounts[n-1]
	}

	clipboard.WriteAll(acc.Password)
	color.Green("Пароль скопирован")

	if clearTimer != nil {
		clearTimer.Stop()
	}
	clearTimer = time.AfterFunc(10*time.Second, func() {
		clipboard.WriteAll("")
		color.Yellow("Буфер обмена очищен")
	})
}

func backupVault(vault *account.VaultWithDb) {
	data, err := json.MarshalIndent(&vault.Data, "", "  ")
	if err != nil {
		output.PrintError("Ошибка экспорта")
		return
	}

	password := PromptPassword("Пароль для бэкапа: ")
	if !isStrongPassword(password) {
		color.Red("Слабый пароль. Используйте 8+ символов, цифры и спецсимволы.")
		return
	}

	encrypted, err := crypto.Encrypt(data, []byte(password))
	if err != nil {
		output.PrintError("Ошибка шифрования")
		return
	}

	os.Mkdir("backup", 0700)
	t := time.Now().Format("2006-01-02_15-04-05")
	fn := "backup/vault_" + t + ".enc"
	os.WriteFile(fn, encrypted, 0600)
	color.Green("Резервная копия: %s", fn)
}

func restoreFromBackup(vault *account.VaultWithDb) {
	fn := prompt("Путь к бэкапу: ")
	password := PromptPassword("Мастер-пароль: ")

	data, err := files.NewJsonDb(fn).ReadFile()
	if err != nil {
		output.PrintError("Файл не найден")
		return
	}

	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		color.Red("Неверный пароль")
		return
	}

	var backup account.Vault
	json.Unmarshal(decrypted, &backup)

	vault.Lock()
	vault.Data.Accounts = backup.Accounts
	vault.Data.UpdatedAt = time.Now()
	vault.Data.Verification = "VERIFIED"
	vault.Unlock()

	vault.Save()
	color.Green("Восстановлено!")
}

func isStrongPassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	var upper, lower, digit, special bool
	specials := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, c := range p {
		switch {
		case c >= 'A' && c <= 'Z':
			upper = true
		case c >= 'a' && c <= 'z':
			lower = true
		case c >= '0' && c <= '9':
			digit = true
		case strings.ContainsRune(specials, c):
			special = true
		}
	}
	return upper && lower && digit && special
}

func LoadVault(db account.Db, password string) *account.VaultWithDb {
	data, err := db.Read()
	if err != nil {
		color.Cyan("Файл не найден. Создаём новый сейф.")
		return &account.VaultWithDb{
			Data: account.Vault{
				Accounts:  []account.Account{},
				UpdatedAt: time.Now(),
			},
			Db: db,
		}
	}

	// Сначала попробуем расшифровать
	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		// Если не получилось — может, файл не шифровался?
		var vault account.Vault
		if json.Unmarshal(data, &vault) == nil {
			color.Yellow("Загружено без шифрования")
			return &account.VaultWithDb{Data: vault, Db: db}
		}
		color.Red("Неверный пароль или повреждённый файл")
		os.Exit(1)
	}

	var vault account.Vault
	if json.Unmarshal(decrypted, &vault) != nil {
		color.Red("Ошибка анализа данных")
		os.Exit(1)
	}

	return &account.VaultWithDb{Data: vault, Db: db}
}

func SaveEncrypted(vault *account.VaultWithDb, password string) error {
	data, err := vault.ToBytes()
	if err != nil {
		return err
	}

	encrypted, err := crypto.Encrypt(data, []byte(password))
	if err != nil {
		return err
	}

	return vault.Db.Write(encrypted)
}
