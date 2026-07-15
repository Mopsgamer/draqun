package internal

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

type TestServer struct {
	App      *fiber.App
	Client   *client.Client
	Listener net.Listener
}

func chdirToRoot() {
	for range 5 {
		if _, err := os.Stat("go.mod"); err == nil {
			return
		}
		_ = os.Chdir("..")
	}
}

func testSetupApp(t testing.TB) *fiber.App {
	chdirToRoot()

	// Create a unique temporary directory for this test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_app_data.db")

	// Set required environment variables
	os.Setenv("JWT_KEY", "test_jwt_key_that_is_long_enough_12345")
	os.Setenv("DB_PATH", dbPath)

	// Close existing database if any to prevent locking issues
	if model.Sqlx != nil {
		_ = model.Sqlx.Close()
		model.Sqlx = nil
	}

	// Load environment variables
	err := environment.LoadEnv(nil)
	if err != nil {
		t.Fatalf("failed to load env: %v", err)
	}

	// Load database
	err = model.LoadDB()
	if err != nil {
		t.Fatalf("failed to load db: %v", err)
	}

	// Initialize tables by running SQL scripts in the correct order
	sqlFileList := []string{
		"./scripts/queries/create_users.sql",
		"./scripts/queries/create_groups.sql",
		"./scripts/queries/create_group_members.sql",
		"./scripts/queries/create_group_roles.sql",
		"./scripts/queries/create_group_role_assignees.sql",
		"./scripts/queries/create_group_messages.sql",
		"./scripts/queries/create_group_action_memberships.sql",
		"./scripts/queries/create_group_action_kicks.sql",
		"./scripts/queries/create_group_action_bans.sql",
	}

	for _, file := range sqlFileList {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read sql file %s: %v", file, err)
		}
		_, err = model.Sqlx.Exec(string(content))
		if err != nil {
			t.Fatalf("failed to execute sql script %s: %v", file, err)
		}
	}

	// Instantiate the Fiber application
	app, err := NewApp(os.DirFS("."), false)
	if err != nil {
		t.Fatalf("failed to create fiber app: %v", err)
	}

	return app
}

func startTestServer(t testing.TB, app *fiber.App) *TestServer {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on TCP: %v", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	go func() {
		_ = app.Listener(ln)
	}()

	t.Cleanup(func() {
		_ = ln.Close()
		_ = app.Shutdown()
	})

	cli := client.New()
	cli.SetBaseURL(baseURL)

	return &TestServer{
		App:      app,
		Client:   cli,
		Listener: ln,
	}
}
