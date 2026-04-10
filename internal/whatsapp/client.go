package whatsapp

import (
	"context"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func GetPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".wapclip")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.db"), nil
}

func InitClient() (*whatsmeow.Client, error) {
	path, err := GetPath()
	if err != nil {
		return nil, err
	}

	dbLog := waLog.Stdout("Database", "ERROR", true)
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	container, err := sqlstore.New(context.Background(), "sqlite", dsn, dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return client, nil
}
