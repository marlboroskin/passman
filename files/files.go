package files

import (
	"fmt"
	"menedger_paroley/output"
	"os"
)

type JsonDb struct {
	name string
}

func NewJsonDb(name string) *JsonDb {
	return &JsonDb{name: name}
}

func (db *JsonDb) WriteFile(content []byte) error {
	file, err := os.OpenFile(db.name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		output.PrintError(err)
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		output.PrintError(err)
		return err
	}
	fmt.Println("Запись успешна")
	os.Stdout.Sync()
	return nil
}

func (db *JsonDb) ReadFile() ([]byte, error) {
	data, err := os.ReadFile(db.name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (db *JsonDb) Read() ([]byte, error) {
	return os.ReadFile(db.name)
}

func (db *JsonDb) Write(data []byte) error {
	file, err := os.OpenFile(db.name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		output.PrintError(err)
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		output.PrintError(err)
		return err
	}
	fmt.Println("Запись успешна")
	os.Stdout.Sync()
	return nil
}

