package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"menedger_paroley/account"
	"menedger_paroley/crypto"
	"menedger_paroley/files"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
)

func main() {
	// 1. Создать аккаунт
	// 2. Найти аккаунт
	// 3. Удалить аккаунт
	// 4. Выход

	fmt.Println("__Менеджер паролей__")
	fmt.Println("Текущая папка:", getCurrentDir())
	vault := loadVault() // один раз

	for {
		switch getMenu() {
		case 1:
			createAccount(vault) // ← передаём тот же экземпляр
		case 2:
			findAccount(vault)
		case 3:
			deleteAccount(vault)
		case 4:
			saveVault(vault) // ← сохраняем перед выходом
			return
		case 5:
			length := promptInt("Длина пароля: ")
			password := GeneratePassword(length)
			fmt.Println("Сгенерировано:", password)
		case 6:
			copyPassword(vault)
		case 7:
			backupVault(vault)
		case 8:
			restoreFromBackup(vault)

		}
	}
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

func findAccount(vault *account.Vault) {
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

func deleteAccount(vault *account.Vault) {
	url := promptData("Введите URL для поиска: ")
	isDeleted := vault.DeleteAccountByURL(url)
	if isDeleted {
		color.Green("Аккаунт удалён")
		saveVault(vault) 
	} else {
		color.Red("Аккаунт не найден")
	}
}

func createAccount(vault *account.Vault) {
	name := promptData("Введите имя аккаунта (например, Гугл): ")
	login := promptData("Введите логин: ")
	password := promptData("Введите пароль: ")
	urlString := promptData("Введите URL: ")

	myAccount, err := account.NewAccount(name, login, password, urlString)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	// Добавляем аккаунт
	vault.AddAccount(*myAccount)

	// Сохраняем
	saveVault(vault)
}

func loadVault() *account.Vault {
	return loadVaultWithRetries(0)
}

func loadVaultWithRetries(attempts int) *account.Vault {
	// Увеличиваем задержку с каждой попыткой
	if attempts > 0 {
		seconds := time.Second * time.Duration(1<<attempts) // 2, 4, 8, 16... сек
		if seconds > 30*time.Second {
			seconds = 30 * time.Second // лимит
		}
		fmt.Printf("Ждите %s перед следующей попыткой...\n", seconds)
		time.Sleep(seconds)
	}

	password := promptPassword("Введите мастер-пароль: ")

	data, err := files.ReadFile("data.enc")
	if err != nil {
		fmt.Println("Файл не найден. Создаём новый.")
		return account.NewVault()
	}

	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		color.Red("Неверный пароль или ошибка расшифровки")
		if attempts >= 5 {
			color.Red("Слишком много попыток. Приложение завершено.")
			os.Exit(1)
		}
		return loadVaultWithRetries(attempts + 1)
	}

	var vault account.Vault
	err = json.Unmarshal(decrypted, &vault)
	if err != nil {
		fmt.Println("Ошибка анализа данных. Создаём новый сейф.")
		return account.NewVault()
	}

	return &vault
}

func saveVault(vault *account.Vault) {

	data, err := json.MarshalIndent(vault, "", "  ")
	if err != nil {
		fmt.Println("Не удалось преобразовать в JSON")
		return
	}

	password := promptPassword("Подтвердите мастер-пароль для сохранения: ")

	// Проверка силы пароля
	if !isStrongPassword(password) {
		color.Red("Пароль слишком слабый. Используйте минимум 8 символов, цифры и спецсимволы.")
		return
	}

	encrypted, err := crypto.Encrypt(data, []byte(password))
	if err != nil {
		fmt.Println("Ошибка шифрования", err)
		os.Stdout.Sync()
		return
	}

	files.WriteFile(encrypted, "data.enc")
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
		_, err := fmt.Sscanf(input, "%d", &value)
		if err != nil {
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
	fmt.Print(prompt)
	os.Stdout.Sync()
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password)
}
func copyPassword(vault *account.Vault) {
	query := promptData("Введите запрос (имя, логин или URL): ")
	accounts := vault.FindAccount(query)
	if len(accounts) == 0 {
		fmt.Println("Не найдено")
		return
	}
	if len(accounts) == 1 {
		clipboard.WriteAll(accounts[0].Password)
		color.Green("Пароль скопирован")
		return
	}
	fmt.Println("Найдено несколько аккаунтов:")
	for i, acc := range accounts {
		fmt.Printf("%d. %s (%s)\n", i+1, acc.Name, acc.Login)
	}
	n := promptInt("Выберите номер: ")
	if n < 1 || n > len(accounts) {
		fmt.Println("Неверный номер")
		return
	}
	clipboard.WriteAll(accounts[n-1].Password)
	color.Green("Пароль скопирован")
}

func backupVault(vault *account.Vault) {
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
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range p {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func restoreFromBackup(vault *account.Vault) {
	filename := promptData("Введите путь к бэкапу (например, backup/vault_2025-04-05.enc): ")
	password := promptPassword("Введите мастер-пароль: ")

	data, err := files.ReadFile(filename)
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

	// Перезаписываем текущий сейф
	*vault = backupVault
	vault.UpdatedAt = time.Now()
	saveVault(vault)
	color.Green("Восстановление успешно!")
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}
