package ovs

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/threatflux/libgo/pkg/utils/exec"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// SimpleCommandExecutor for basic testing
type SimpleCommandExecutor struct {
	commands []string
	outputs  map[string][]byte
	errors   map[string]error
}

func NewSimpleCommandExecutor() *SimpleCommandExecutor {
	return &SimpleCommandExecutor{
		commands: []string{},
		outputs:  make(map[string][]byte),
		errors:   make(map[string]error),
	}
}

func (e *SimpleCommandExecutor) Execute(cmd string, args ...string) ([]byte, error) {
	return e.ExecuteContext(context.Background(), cmd, args...)
}

func (e *SimpleCommandExecutor) ExecuteContext(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	fullCmd := cmd
	for _, arg := range args {
		fullCmd += " " + arg
	}

	e.commands = append(e.commands, fullCmd)

	if err, exists := e.errors[fullCmd]; exists {
		return nil, err
	}

	if output, exists := e.outputs[fullCmd]; exists {
		return output, nil
	}

	return []byte{}, nil
}

func (e *SimpleCommandExecutor) SetOutput(cmd string, output []byte) {
	e.outputs[cmd] = output
}

func (e *SimpleCommandExecutor) SetError(cmd string, err error) {
	e.errors[cmd] = err
}

func (e *SimpleCommandExecutor) GetCommands() []string {
	return e.commands
}

func TestOVSManager_CreateAndDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	executor := NewSimpleCommandExecutor()
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Allow any logging calls
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	manager := NewOVSManager(executor, mockLogger)

	// Test bridge creation
	executor.SetError("ovs-vsctl br-exists test-bridge", fmt.Errorf("exit status 2")) // Bridge doesn't exist

	err := manager.CreateBridge(context.Background(), "test-bridge")
	assert.NoError(t, err)

	commands := executor.GetCommands()
	assert.Contains(t, commands, "ovs-vsctl br-exists test-bridge")
	assert.Contains(t, commands, "ovs-vsctl add-br test-bridge")
}

func TestOVSManager_ListBridges_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	executor := NewSimpleCommandExecutor()
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Allow any logging calls
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	manager := NewOVSManager(executor, mockLogger)

	// Mock empty bridge list
	executor.SetOutput("ovs-vsctl list-br", []byte("\n"))

	bridges, err := manager.ListBridges(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, bridges)
}

func TestOVSManager_InterfaceImplementation(t *testing.T) {
	// Test that OVSManager implements the Manager interface
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	executor := &exec.DefaultCommandExecutor{}
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	var manager Manager = NewOVSManager(executor, mockLogger)
	assert.NotNil(t, manager)
}
