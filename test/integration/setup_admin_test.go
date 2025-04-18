package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/wroersma/libgo/internal/auth/user"
	"github.com/wroersma/libgo/internal/config"
	usermodels "github.com/wroersma/libgo/internal/models/user"
	"github.com/wroersma/libgo/pkg/logger"
)

// TestSetupAdminUser creates an admin user for testing
// Run with: go test -v ./test/integration -run TestSetupAdminUser
func TestSetupAdminUser(t *testing.T) {
	// Create logger with direct config
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "console",
	}
	log, err := logger.NewZapLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create user service
	userService := user.NewUserService(log)

	// Create admin user with specific ID for testing
	adminUser := usermodels.User{
		ID:       "11111111-2222-3333-4444-555555555555",
		Username: "admin",
		Password: "", // Will be hashed
		Email:    "admin@example.com",
		Roles:    []string{"admin"},
		Active:   true,
	}

	// Hash password
	hashedPassword, err := user.HashPassword("admin")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	adminUser.SetPassword(hashedPassword)

	// Load user into service
	err = userService.LoadUser(&adminUser)
	if err != nil {
		// User might already exist, which is fine
		t.Logf("Note: %v (this is okay if user already exists)", err)
	} else {
		t.Logf("Successfully added admin user to database with ID %s", adminUser.ID)
	}

	// Verify user exists
	ctx := context.Background()
	retrievedUser, err := userService.GetByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("Failed to retrieve admin user after setup: %v", err)
	}

	// Verify user is correctly configured
	if retrievedUser.ID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("Admin user has incorrect ID: %s", retrievedUser.ID)
	}
	if !retrievedUser.Active {
		t.Fatalf("Admin user is not active")
	}

	fmt.Println("✅ Admin user setup complete! ID:", retrievedUser.ID)
}
