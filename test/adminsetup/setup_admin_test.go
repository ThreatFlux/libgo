package adminsetup

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/wroersma/libgo/internal/auth/user"
	"github.com/wroersma/libgo/internal/config"
	"github.com/wroersma/libgo/internal/database"
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

	// Initialize database connection
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Use the same admin details as in test-config.yaml
	adminUsername := "admin"
	adminPassword := "admin"
	adminEmail := "admin@example.com"

	// Use the database directly for more control
	// First, let's check if we already have a user with the admin ID we want
	var count int64
	if err := db.Table("users").Where("id = ?", "11111111-2222-3333-4444-555555555555").Count(&count).Error; err != nil {
		t.Fatalf("Failed to check for existing user: %v", err)
	}

	if count > 0 {
		t.Logf("Admin user already exists with correct ID")
		fmt.Println("✅ Admin user verified! ID: 11111111-2222-3333-4444-555555555555")
		return
	}

	// Direct SQL approach to handle the unique constraint properly
	// First delete any existing admin user to avoid unique constraint errors
	db.Exec("DELETE FROM users WHERE username = ?", adminUsername)

	// Now create our admin user with the fixed ID
	hashedPassword, err := user.HashPassword(adminPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user record directly with SQL
	t.Logf("Creating admin user with ID: 11111111-2222-3333-4444-555555555555")
	query := `INSERT INTO users (id, username, password, email, roles, active, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	err = db.Exec(query,
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
