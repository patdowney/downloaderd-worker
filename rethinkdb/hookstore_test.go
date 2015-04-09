package rethinkdb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd-worker/download"
)

func createSession() (*r.Session, error) {

	host := os.Getenv("RETHINKDB_HOST")
	if host == "" {
		host = "localhost"
	}

	address := fmt.Sprintf("%s:28015", host)

	return r.Connect(r.ConnectOpts{Address: address})
}

func getRandomHexString(byteLength uint) string {
	data := make([]byte, byteLength)
	rand.Read(data)

	return hex.EncodeToString(data)
}

func TestRandomHexString(t *testing.T) {
	s1 := getRandomHexString(8)
	s2 := getRandomHexString(8)

	if s1 == s2 {
		t.Errorf("bytes match: %s=%s", s1, s2)
	}
}

func getTestTableName(prefix string) string {
	return fmt.Sprintf("TestTable_%s_%s", prefix, getRandomHexString(8))
}

func getTestDatabaseName(prefix string) string {
	return fmt.Sprintf("TestDatabase_%s_%s", prefix, getRandomHexString(8))
}

func createHookStore() (*HookStore, error) {
	s, err := createSession()
	if err != nil {
		return nil, err
	}

	tableName := getTestTableName("HookStore")
	databaseName := getTestDatabaseName("HookStore")

	hookStore, err := NewHookStoreWithSession(s, databaseName, tableName)
	if err != nil {
		return nil, err
	}

	return hookStore, nil
}

func deleteHookStore(store *HookStore) error {
	return r.DbDrop(store.DatabaseName).Exec(store.Session)
}

func TestHookKeyIndex(t *testing.T) {
	store, err := createHookStore()
	defer deleteHookStore(store)

	if err != nil {
		t.Fatal(err)
	}
	h := &download.Hook{DownloadID: "test-download-id", RequestID: "test-request-id"}
	err = store.Add(h)
	if err != nil {
		t.Error(err)
	}
	hooks, err := store.FindByHookKey("test-download-id", "test-request-id")
	if err != nil {
		t.Error(err)
	}

	if len(hooks) != 1 {
		t.Errorf("wrong number of hooks found: expected 1, got %v", len(hooks))
	}
}

func TestRequestIDIndex(t *testing.T) {
	store, err := createHookStore()
	defer deleteHookStore(store)

	if err != nil {
		t.Fatal(err)
	}
	h := &download.Hook{DownloadID: "test-download-id", RequestID: "test-request-id"}
	err = store.Add(h)
	if err != nil {
		t.Error(err)
	}
	hooks, err := store.FindByRequestID("test-request-id")
	if err != nil {
		t.Error(err)
	}

	if len(hooks) != 1 {
		t.Errorf("wrong number of hooks found: expected 1, got %v", len(hooks))
	}
}

func TestDownloadIDIndex(t *testing.T) {
	store, err := createHookStore()
	defer deleteHookStore(store)

	if err != nil {
		t.Fatal(err)

	}
	h := &download.Hook{DownloadID: "test-download-id", RequestID: "test-request-id"}
	err = store.Add(h)
	if err != nil {
		t.Error(err)
	}
	hooks, err := store.FindByDownloadID("test-download-id")
	if err != nil {
		t.Error(err)
	}

	if len(hooks) != 1 {
		t.Errorf("wrong number of hooks found: expected 1, got %v", len(hooks))
	}
}
