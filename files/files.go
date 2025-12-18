package files

import (
	"fmt"
	"os"
)

func WriteFile(content []byte, name string) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer file.Close() // ← Сразу после открытия

	_, err = file.Write(content)
	if err != nil {
		fmt.Println("Ошибка записи:", err)
		return
	}
	fmt.Println("Запись успешна")
}

func ReadFile(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}
