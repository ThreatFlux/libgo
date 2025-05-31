package adminsetup

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/threatflux/libgo/internal/auth/user"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/database"
)

// TestSetupAdminUser creates an admin user for testing
// Run with: go test -v ./test/adminsetup -run TestSetupAdminUser
func TestSetupAdminUser(t *testing.T) {
	// No need for complex logger in this test, we'll use t.Log

	// Load test configuration to get database settings
	// Use relative path from test directory to project root
	wd, _ := os.Getwd()
	fmt.Println("Current working directory:", wd)

	configPath := "../../configs/test-config.yaml"
	fmt.Println("Using config path:", configPath)

	loader := config.NewYAMLLoader(configPath)
	cfg := &config.Config{}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}

	// Set the DSN directly for SQLite - the config is missing DSN processing
	cfg.Database.DSN = "../../libgo.db"
	t.Logf("Using database DSN: %s", cfg.Database.DSN)
	t.Logf("Database driver: %s", cfg.Database.Driver)

	// Initialize database connection
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a simple test table first to verify database connection works
	if err := db.Exec("CREATE TABLE IF NOT EXISTS test_table (id INTEGER)").Error; err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	t.Log("Test table created successfully")

	// Try creating the users table manually
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		roles TEXT NOT NULL DEFAULT '[]',
		active BOOLEAN DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`

	if err := db.Exec(createTableSQL).Error; err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	t.Log("Users table created successfully")

	// List all tables for debugging
	var allTables []string
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Pluck("name", &allTables).Error; err != nil {
		t.Fatalf("Failed to list tables: %v", err)
	}
	t.Logf("All tables in database: %v", allTables)

	// Verify table was created
	var tableCount int64
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount).Error; err != nil {
		t.Fatalf("Failed to verify table creation: %v", err)
	}
	t.Logf("Table count for 'users': %d", tableCount)
	if tableCount == 0 {
		t.Fatal("Users table was not created")
	}
	t.Log("Users table verified successfully")

	// Use the same admin details as in test-config.yaml
	adminUsername := "admin"
	adminPassword := "admin"
	adminEmail := "admin@example.com"

	// Use the database directly for more control
	// First, let's check if we already have a user with the admin ID we want
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM users WHERE id = ?", "11111111-2222-3333-4444-555555555555").Scan(&count).Error; err != nil {
		t.Fatalf("Failed to check for existing user: %v", err)
	}

	if count > 0 {
		t.Logf("Admin user already exists with correct ID")
		fmt.Println("✅ Admin user verified! ID: 11111111-2222-3333-4444-555555555555")
		return
	}

	// Direct approach to handle the unique constraint properly
	// First delete any existing admin user to avoid unique constraint errors
	db.Exec("DELETE FROM users WHERE username = ?", adminUsername)

	// Now create our admin user with the fixed ID
	hashedPassword, err := user.HashPassword(adminPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user record using raw SQL
	t.Logf("Creating admin user with ID: 11111111-2222-3333-4444-555555555555")
	insertSQL := `INSERT INTO users (id, username, password, email, roles, active, created_at, updated_at)
	              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	err = db.Exec(insertSQL,
		"11111111-2222-3333-4444-555555555555",
		adminUsername,
		hashedPassword,
		adminEmail,
		`["admin"]`,
		true,
		time.Now(),
		time.Now()).Error

	if err != nil {
		t.Fatalf("Failed to insert admin user: %v", err)
	}

	t.Logf("Successfully created admin user with fixed ID")
	fmt.Println("✅ Admin user setup complete! ID: 11111111-2222-3333-4444-555555555555")
}
