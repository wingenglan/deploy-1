package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

// newTestServer returns an isolated in-memory API server for handler tests.
func newTestServer(t *testing.T) *server {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := initializeDatabase(db); err != nil {
		t.Fatal(err)
	}
	return &server{db: db}
}

// TestCreateAndListTasks verifies the primary task creation and retrieval flow.
func TestCreateAndListTasks(t *testing.T) {
	s := newTestServer(t)
	body := bytes.NewBufferString(`{"title":"发布检查","owner":"李明","status":"todo","dueDate":"2026-08-12"}`)
	created := httptest.NewRecorder()
	s.createTask(created, httptest.NewRequest(http.MethodPost, "/api/tasks", body))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", created.Code, http.StatusCreated)
	}

	list := httptest.NewRecorder()
	s.listTasks(list, httptest.NewRequest(http.MethodGet, "/api/tasks", nil))
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", list.Code, http.StatusOK)
	}
	if !bytes.Contains(list.Body.Bytes(), []byte("发布检查")) {
		t.Fatalf("list response does not contain created task: %s", list.Body.String())
	}
}

// TestCreateTaskRejectsInvalidStatus verifies that invalid task state cannot enter storage.
func TestCreateTaskRejectsInvalidStatus(t *testing.T) {
	s := newTestServer(t)
	body := bytes.NewBufferString(`{"title":"发布检查","owner":"李明","status":"paused","dueDate":"2026-08-12"}`)
	response := httptest.NewRecorder()
	s.createTask(response, httptest.NewRequest(http.MethodPost, "/api/tasks", body))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
