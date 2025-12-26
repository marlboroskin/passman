package cloud

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CloudDb struct {
	URL      string
	Username string
	Password string
}

func NewCloudDb(url, username, password string) *CloudDb {
	return &CloudDb{
		URL:      url,
		Username: username,
		Password: password,
	}
}

func (db *CloudDb) Read() ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", db.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.SetBasicAuth(db.Username, db.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер вернул ошибку: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела: %w", err)
	}

	return data, nil
}

func (db *CloudDb) Write(data []byte) error {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("PUT", db.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.SetBasicAuth(db.Username, db.Password)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("сервер вернул: %d", resp.StatusCode)
	}

	fmt.Println("Данные успешно сохранены в облаке")
	return nil
}
