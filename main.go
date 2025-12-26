package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"menedger_paroley/account"
	"menedger_paroley/cloud"
	"menedger_paroley/crypto"
	"menedger_paroley/files"
	"menedger_paroley/output"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
)

var (
	attemptCount   = 0
	blockUntilTime time.Time
	clearTimer     *time.Timer

	rememberedHash []byte
	rememberUntil  time.Time
	hashMutex      sync.RWMutex
)

var menu = map[string]func(*account.VaultWithDb){
	"1": createAccount,
	"2": findAccount,
	"3": deleteAccount,
	"4": func(v *account.VaultWithDb) { saveVault(v); os.Exit(0) },
	"5": func(v *account.VaultWithDb) {
		length := promptInt("Длина пароля: ")
		password := GeneratePassword(length)
		fmt.Println("Сгенерировано:", password)
	},
	"6": copyPassword,
	"7": backupVault,
	"8": restoreFromBackup,
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			fmt.Println("\n❌ Используйте пункт меню 4 для выхода.")
			os.Stdout.Sync()
		}
	}()

	defer func() {
		if clearTimer != nil {
			clearTimer.Stop()
			clipboard.WriteAll("")
			color.Yellow("Буфер обмена очищен при выходе.")
		}

		// Сбросим временный хэш
		hashMutex.Lock()
		rememberedHash = nil
		rememberUntil = time.Time{}
		hashMutex.Unlock()
	}()

	data, err := os.ReadFile("block.lock")
	if err == nil {
		if t, err := time.Parse(time.RFC3339, string(data)); err == nil {
			if time.Now().Before(t) {
				remaining := t.Sub(time.Now()).Minutes()
				color.Red("Приложение заблокировано. Повторите через %.0f минут.", math.Ceil(remaining))
				color.Red("Закройте программу и попробуйте позже.")
				os.Stdout.Sync()
				time.Sleep(5 * time.Second)
				return
			}
		}
		os.Remove("block.lock")
	}

	fmt.Println("__Менеджер паролей__")
	db := chooseStorage()
	vault := loadVault(db)

	// Выбор режима
	fmt.Println("Запуск в режиме CLI...")
	runCLI(vault)
}

func menuFunc(vault *account.VaultWithDb) {
	for {
		fmt.Println("\n__Менеджер паролей__")
		fmt.Println("1. Создать аккаунт")
		fmt.Println("2. Найти аккаунт")
		fmt.Println("3. Удалить аккаунт")
		fmt.Println("4. Выход")
		fmt.Println("5. Сгенерировать пароль")
		fmt.Println("6. Скопировать пароль в буфер обмена")
		fmt.Println("7. Создать резервную копию сейфа")
		fmt.Println("8. Восстановить сейф из резервной копии")

		choice := promptData("Выберите пункт меню: ")

		if action, ok := menu[choice]; ok {
			action(vault)
		} else {
			color.Red("Неверный выбор. Попробуйте снова.")
		}
	}
}

// CLI логика — вынесем в отдельную функцию
func runCLI(vault *account.VaultWithDb) {
	menuFunc(vault)
}

func getMenu() int {
	fmt.Println("Выберите вариант:")
	fmt.Println("1. Создать аккаунт")
	fmt.Println("2. Найти аккаунт")
	fmt.Println("3. Удалить аккаунт")
	fmt.Println("4. Выход")
	fmt.Println("5. Сгенерировать пароль")
	fmt.Println("6. Скопировать пароль в буфер обмена")
	fmt.Println("7. Создать резервную копию сейфа")
	fmt.Println("8. Восстановить сейф из резервной копии")

	input := promptData("Выберите пункт меню: ")
	var variant int
	fmt.Sscanf(input, "%d", &variant)

	if variant < 1 || variant > 8 {
		fmt.Println("Неверный выбор. Попробуйте снова.")
		return getMenu() // Рекурсивно, пока не введут правильно
	}
	return variant
}

func findAccount(vault *account.VaultWithDb) {
	if time.Now().Before(blockUntilTime) {
		remaining := blockUntilTime.Sub(time.Now()).Minutes()
		color.Red("Приложение заблокировано. Повторите через %.0f минут.", math.Ceil(remaining))
		return
	}

	password := promptPassword("Введите мастер-пароль для доступа: ")

	if len(vault.Accounts) > 0 {
		if !verifyMasterPassword(password) {
			attemptCount++
			color.Red("Неверный пароль. Осталось попыток: %d", 3-attemptCount)
			playErrorSound()
			os.Stdout.Sync()

			if attemptCount >= 3 {
				blockUntilTime = time.Now().Add(20 * time.Minute)
				err := os.WriteFile("block.lock", []byte(blockUntilTime.Format(time.RFC3339)), 0600)
				if err != nil {
					fmt.Println("Не удалось сохранить блокировку")
				}
				color.Red("Слишком много попыток. Приложение заблокировано на 20 минут.")
				os.Stdout.Sync()
			}
			return
		}
	}

	attemptCount = 0

	query := promptData("Введите запрос (имя, логин или URL): ")
	accounts := vault.FindAccount(query)
	if len(accounts) == 0 {
		fmt.Println("Аккаунты не найдены")
		return
	}
	fmt.Printf("Найдено: %d аккаунтов\n", len(accounts))
	fmt.Println("---")
	for _, acc := range accounts {
		acc.Output()
	}
}

func deleteAccount(vault *account.VaultWithDb) {
	url := promptData("Введите полный или частичный URL: ")
	isDeleted := vault.DeleteAccountByURL(url)
	if isDeleted {
		color.Green("Аккаунт(ы) удалён(ы)")
		saveVault(vault)
	} else {
		output.PrintError("Аккаунт не найден")
	}
}

func createAccount(vault *account.VaultWithDb) {
	name := promptData("Введите имя аккаунта (например, Гугл): ")
	login := promptData("Введите логин: ")
	password := promptData("Введите пароль: ")
	urlString := promptData("Введите URL: ")

	myAccount, err := account.NewAccount(name, login, password, urlString)
	if err != nil {
		output.PrintError("Неверный формат URL или Логина")
		return
	}

	// Добавляем аккаунт
	vault.AddAccount(*myAccount)

	// Сохраняем
	saveVault(vault)
}

func loadVault(db account.Db) *account.VaultWithDb {
	vaultWithDb := account.NewVault(db)

	data, err := db.Read()
	if err != nil {
		fmt.Println("Файл не найден. Создаём новый сейф.")
		return vaultWithDb
	}

	password := promptPassword("Введите мастер-пароль: ")
	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		color.Red("Неверный пароль или повреждённый файл")
		playErrorSound()
		os.Exit(1)
	}

	var loadedVault account.Vault
	err = json.Unmarshal(decrypted, &loadedVault)
	if err != nil {
		color.Red("Ошибка анализа данных")
		os.Exit(1)
	}

	vaultWithDb.Mu.Lock()
	vaultWithDb.Accounts = loadedVault.Accounts
	vaultWithDb.UpdatedAt = loadedVault.UpdatedAt
	vaultWithDb.Verification = "VERIFIED"
	vaultWithDb.Mu.Unlock()

	return vaultWithDb
}

func saveVault(vault *account.VaultWithDb) {
	vault.Verification = "VERIFIED"

	data, err := json.MarshalIndent(&vault.Vault, "", "  ")
	if err != nil {
		fmt.Println("Ошибка JSON")
		return
	}

	password := promptPassword("Подтвердите мастер-пароль для сохранения: ")
	if !isStrongPassword(password) {
		color.Red("Слабый пароль")
		return
	}

	// Шифруем данные
	encrypted, err := crypto.Encrypt(data, []byte(password))
	if err != nil {
		fmt.Println("Ошибка шифрования")
		return
	}

	// Записываем зашифрованные данные в хранилище
	err = vault.Db.Write(encrypted)
	if err != nil {
		color.Red("Ошибка сохранения: %v", err)
		return
	}

	// Шифруем и сохраняем токен
	token := []byte("MASTER_PASSWORD_VERIFIED")
	encryptedToken, err := crypto.Encrypt(token, []byte(password))
	if err != nil {
		fmt.Println("Ошибка шифрования токена")
		return
	}
	err = os.WriteFile("token.enc", encryptedToken, 0600)
	if err != nil {
		fmt.Println("Не удалось сохранить токен")
		return
	}

	color.Green("Данные сохранены")
}

func promptData(prompt string) string {
	fmt.Print(prompt)
	os.Stdout.Sync()
	reader := bufio.NewReader(os.Stdin)
	res, _ := reader.ReadString('\n')
	return strings.TrimSpace(res)
}

func GeneratePassword(n int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"
	runes := []rune(chars)
	res := make([]rune, n)
	for i := range res {
		res[i] = runes[rand.IntN(len(runes))]
	}
	return string(res)
}

func promptInt(prompt string) int {
	for {
		input := promptData(prompt)
		var value int
		n, err := fmt.Sscanf(input, "%d", &value)
		if n != 1 || err != nil {
			fmt.Println("Введите число.")
			continue
		}
		if value < 1 || value > 128 {
			fmt.Println("Длина от 1 до 128.")
			continue
		}
		return value
	}
}

func promptPassword(prompt string) string {
	for {
		fmt.Print(prompt)
		defer fmt.Println()
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err == nil {
			return strings.TrimSpace(string(password))
		}
		fmt.Println("\nОшибка ввода. Попробуйте ещё раз.")
	}
}

func copyPassword(vault *account.VaultWithDb) {
	query := promptData("Введите запрос (имя, логин или URL): ")
	accounts := vault.FindAccount(query)
	if len(accounts) == 0 {
		fmt.Println("Не найдено")
		return
	}

	// Останавливаем предыдущий таймер, если он был
	if clearTimer != nil {
		clearTimer.Stop()
	}

	var selectedAccount account.Account
	if len(accounts) == 1 {
		selectedAccount = accounts[0]
	} else {
		fmt.Println("Найдено несколько аккаунтов:")
		for i, acc := range accounts {
			fmt.Printf("%d. %s (%s)\n", i+1, acc.Name, acc.Login)
		}
		n := promptInt("Выберите номер: ")
		if n < 1 || n > len(accounts) {
			fmt.Println("Неверный номер")
			return
		}
		selectedAccount = accounts[n-1]
	}

	// Копируем пароль
	clipboard.WriteAll(selectedAccount.Password)
	color.Green("Пароль скопирован")

	// Запускаем таймер на очистку
	clearTimer = time.AfterFunc(10*time.Second, func() {
		clipboard.WriteAll("")
		color.Yellow("Буфер обмена очищен")
	})
}

func backupVault(vault *account.VaultWithDb) {
	data, err := vault.ToBytes()
	if err != nil {
		fmt.Println("Ошибка экспорта данных")
		return
	}

	password := promptPassword("Введите мастер-пароль для шифрования бэкапа: ")
	if !isStrongPassword(password) {
		color.Red("Пароль слишком слабый. Используйте 8+ символов, цифры и спецсимволы.")
		return
	}

	encrypted, err := crypto.Encrypt(data, []byte(password))
	if err != nil {
		fmt.Println("Ошибка шифрования бэкапа")
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup/vault_%s.enc", timestamp)

	os.Mkdir("backup", 0700)
	err = os.WriteFile(filename, encrypted, 0600)
	if err != nil {
		fmt.Println("Ошибка записи файла:", err)
		return
	}

	color.Green("Зашифрованная резервная копия создана: %s", filename)
}

func isStrongPassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	hasSpecial := false
	for _, c := range p {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, c):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func restoreFromBackup(vault *account.VaultWithDb) {
	filename := promptData("Введите путь к бэкапу (например, backup/vault_2025-04-05.enc): ")
	password := promptPassword("Введите мастер-пароль: ")

	data, err := files.NewJsonDb(filename).ReadFile()
	if err != nil {
		fmt.Println("Файл не найден:", err)
		return
	}

	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		color.Red("Неверный пароль или повреждённый файл")
		return
	}

	var backupVault account.Vault
	err = json.Unmarshal(decrypted, &backupVault)
	if err != nil {
		fmt.Println("Ошибка анализа данных:", err)
		return
	}

	vault.Lock()
	vault.Accounts = backupVault.Accounts
	vault.UpdatedAt = time.Now()
	vault.Verification = "VERIFIED"
	vault.Unlock()

	saveVault(vault)
	color.Green("Восстановление успешно!")
}

func verifyMasterPassword(password string) bool {
	// Проверяем, не запомнен ли пароль
	hashMutex.RLock()
	if time.Now().Before(rememberUntil) && constantTimeEqual(generateHash(password), rememberedHash) {
		hashMutex.RUnlock()
		return true
	}
	hashMutex.RUnlock()

	// Проверяем токен
	data, err := os.ReadFile("token.enc")
	if err != nil {
		return false
	}

	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		return false
	}

	if string(decrypted) == "MASTER_PASSWORD_VERIFIED" {
		// Запоминаем хэш на 10 минут
		hashMutex.Lock()
		rememberedHash = generateHash(password)
		rememberUntil = time.Now().Add(10 * time.Minute)
		hashMutex.Unlock()
		return true
	}

	return false
}

// generateHash создаёт хэш пароля (без salt — только для временного сравнения)
func generateHash(p string) []byte {
    h := sha256.Sum256([]byte(p))
    return h[:] 
}

// constantTimeEqual безопасно сравнивает два хэша
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func playErrorSound() {
	fmt.Print("\a")
}

func chooseStorage() account.Db {
	fmt.Println("Выберите хранилище:")
	fmt.Println("1. Локальный файл (data.enc)")
	fmt.Println("2. Облако (WebDAV / HTTP сервер)")

	var choice int
	fmt.Scanf("%d", &choice)

	switch choice {
	case 1:
		return files.NewJsonDb("data.enc")
	case 2:
		return configureCloud()
	default:
		fmt.Println("Неверный выбор, используем локальное хранилище.")
		return files.NewJsonDb("data.enc")
	}
}

func configureCloud() account.Db {
	url := promptData("Введите URL облака (например, https://example.com/vault.enc): ")
	user := promptData("Логин для облака: ")
	pass := promptPassword("Пароль для облака: ")
	return cloud.NewCloudDb(url, user, pass)
}
