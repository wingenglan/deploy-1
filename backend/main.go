package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	statusTodo       = "todo"
	statusInProgress = "in_progress"
	statusDone       = "done"
)

// Task is the task-board record persisted in SQLite and exposed by the API.
type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Owner     string `json:"owner"`
	Status    string `json:"status"`
	DueDate   string `json:"dueDate"`
	CreatedAt string `json:"createdAt"`
}

// taskInput defines the fields users can submit when creating or updating a task.
type taskInput struct {
	Title   string `json:"title"`
	Owner   string `json:"owner"`
	Status  string `json:"status"`
	DueDate string `json:"dueDate"`
}

// server groups the database dependency used by HTTP handlers.
type server struct {
	db *sql.DB
}

// main opens the database, creates the schema, and starts the HTTP API.
func main() {
	databasePath := environment("DATABASE_PATH", "flowboard.db")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil && filepath.Dir(databasePath) != "." {
		log.Fatalf("create database directory: %v", err)
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := initializeDatabase(db); err != nil {
		log.Fatalf("initialize database: %v", err)
	}

	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/tasks", s.listTasks)
	mux.HandleFunc("POST /api/tasks", s.createTask)
	mux.HandleFunc("PUT /api/tasks/{id}", s.updateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.deleteTask)

	port := environment("PORT", "8081")
	log.Printf("Flowboard API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(mux)))
}

// environment returns the configured value or a documented local default.
func environment(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

// initializeDatabase creates the minimal task table and adds starter records once.
func initializeDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        owner TEXT NOT NULL,
        status TEXT NOT NULL,
        due_date TEXT NOT NULL,
        created_at TEXT NOT NULL
    )`)
	if err != nil {
		return err
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count); err != nil || count > 0 {
		return err
	}
	_, err = db.Exec(`INSERT INTO tasks (title, owner, status, due_date, created_at) VALUES
        ('梳理本周发布清单', '林可', 'in_progress', '2026-08-08', ?),
        ('确认客户验收反馈', '陈舟', 'todo', '2026-08-10', ?),
        ('更新运行手册', '许宁', 'done', '2026-08-05', ?)`, now(), now(), now())
	return err
}

// health confirms that the process and its database connection are available.
func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	if err := s.db.Ping(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "数据库暂不可用")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// listTasks returns all tasks ordered by work status and creation time.
func (s *server) listTasks(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.db.Query(`SELECT id, title, owner, status, due_date, created_at FROM tasks
        ORDER BY CASE status WHEN 'in_progress' THEN 0 WHEN 'todo' THEN 1 ELSE 2 END, id DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取任务失败")
		return
	}
	defer rows.Close()
	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Owner, &task.Status, &task.DueDate, &task.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "读取任务失败")
			return
		}
		tasks = append(tasks, task)
	}
	writeJSON(w, http.StatusOK, tasks)
}

// createTask validates and persists a new task record.
func (s *server) createTask(w http.ResponseWriter, r *http.Request) {
	input, err := decodeInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.db.Exec("INSERT INTO tasks (title, owner, status, due_date, created_at) VALUES (?, ?, ?, ?, ?)", input.Title, input.Owner, input.Status, input.DueDate, now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建任务失败")
		return
	}
	id, _ := result.LastInsertId()
	writeJSON(w, http.StatusCreated, Task{ID: id, Title: input.Title, Owner: input.Owner, Status: input.Status, DueDate: input.DueDate, CreatedAt: now()})
}

// updateTask replaces the editable fields of an existing task.
func (s *server) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效任务编号")
		return
	}
	input, err := decodeInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.db.Exec("UPDATE tasks SET title = ?, owner = ?, status = ?, due_date = ? WHERE id = ?", input.Title, input.Owner, input.Status, input.DueDate, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新任务失败")
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(w, http.StatusNotFound, "任务不存在")
		return
	}
	writeJSON(w, http.StatusOK, Task{ID: id, Title: input.Title, Owner: input.Owner, Status: input.Status, DueDate: input.DueDate})
}

// deleteTask removes a task by its numeric identifier.
func (s *server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效任务编号")
		return
	}
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除任务失败")
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(w, http.StatusNotFound, "任务不存在")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeInput decodes a request body and enforces the task contract.
func decodeInput(r *http.Request) (taskInput, error) {
	var input taskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return input, errors.New("请求内容必须是 JSON")
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Owner = strings.TrimSpace(input.Owner)
	if input.Title == "" || input.Owner == "" || input.DueDate == "" {
		return input, errors.New("任务名称、负责人和截止日期不能为空")
	}
	if input.Status != statusTodo && input.Status != statusInProgress && input.Status != statusDone {
		return input, errors.New("任务状态无效")
	}
	if _, err := time.Parse("2006-01-02", input.DueDate); err != nil {
		return input, errors.New("截止日期格式应为 YYYY-MM-DD")
	}
	return input, nil
}

// pathID reads the route id parameter as a positive integer.
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

// now returns the API timestamp format stored in the database.
func now() string { return time.Now().UTC().Format(time.RFC3339) }

// writeJSON writes a JSON response with the supplied status code.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// writeError writes the standard error payload returned by all API failures.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// cors permits the Vite development server while keeping production same-origin.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
