package screens

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/system"
	tea "github.com/charmbracelet/bubbletea"
)

// stubDetectOK returns a mock detect function that always succeeds.
func stubDetectOK(result *system.DetectionResult) DetectFunc {
	return func(_ context.Context) (*system.DetectionResult, error) {
		return result, nil
	}
}

// stubDetectErr returns a mock detect function that always fails.
func stubDetectErr(err error) DetectFunc {
	return func(_ context.Context) (*system.DetectionResult, error) {
		return nil, err
	}
}

func TestDetection_InitialState_IsLoading(t *testing.T) {
	m := NewDetectionModel(stubDetectOK(&system.DetectionResult{}))
	view := m.View()

	if !strings.Contains(view, "Detecting") {
		t.Error("loading view should contain 'Detecting'")
	}
	if strings.Contains(view, "OS:") {
		t.Error("loading view should not contain OS information yet")
	}
}

func TestDetection_SuccessfulResult_TransitionsToReady(t *testing.T) {
	m := NewDetectionModel(stubDetectOK(&system.DetectionResult{}))

	result := &system.DetectionResult{
		System: system.SystemInfo{
			OS:        "linux",
			Arch:      "amd64",
			Shell:     "bash",
			Supported: true,
		},
	}

	newModel, _ := m.Update(detectionResultMsg{result: result, err: nil})
	m = newModel.(DetectionModel)

	if !m.ready {
		t.Fatal("model should be in ready state after successful detection")
	}
	if m.err != nil {
		t.Fatalf("model.err should be nil, got %v", m.err)
	}

	view := m.View()
	if !strings.Contains(view, "linux") {
		t.Error("ready view should contain OS string 'linux'")
	}
}

func TestDetection_ErrorResult_TransitionsToErrorState(t *testing.T) {
	m := NewDetectionModel(stubDetectErr(errors.New("timeout")))

	newModel, _ := m.Update(detectionResultMsg{result: nil, err: errors.New("timeout")})
	m = newModel.(DetectionModel)

	if !m.ready {
		t.Fatal("model should be ready (error state) after failed detection")
	}
	if m.err == nil {
		t.Fatal("model.err should not be nil in error state")
	}

	view := m.View()
	if !strings.Contains(view, "timeout") {
		t.Error("error view should contain the error message 'timeout'")
	}
	if !strings.Contains(view, "continue") {
		t.Error("error view should contain a continue prompt")
	}
}

func TestDetection_EnterInReadyState_EmitsConfirmedMsg(t *testing.T) {
	m := NewDetectionModel(stubDetectOK(&system.DetectionResult{}))

	result := &system.DetectionResult{System: system.SystemInfo{OS: "darwin"}}
	newModel, _ := m.Update(detectionResultMsg{result: result, err: nil})
	m = newModel.(DetectionModel)

	// Press enter
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = newModel

	if cmd == nil {
		t.Fatal("expected a cmd after enter in ready state")
	}

	msg := cmd()
	confirmed, ok := msg.(DetectionConfirmedMsg)
	if !ok {
		t.Fatalf("expected DetectionConfirmedMsg, got %T", msg)
	}
	if confirmed.Result != result {
		t.Error("DetectionConfirmedMsg.Result should match the stored result")
	}
}

func TestDetection_SpaceInReadyState_EmitsConfirmedMsg(t *testing.T) {
	m := NewDetectionModel(stubDetectOK(&system.DetectionResult{}))

	result := &system.DetectionResult{System: system.SystemInfo{OS: "windows"}}
	newModel, _ := m.Update(detectionResultMsg{result: result, err: nil})
	m = newModel.(DetectionModel)

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	_ = newModel

	if cmd == nil {
		t.Fatal("expected a cmd after space in ready state")
	}
	msg := cmd()
	if _, ok := msg.(DetectionConfirmedMsg); !ok {
		t.Fatalf("expected DetectionConfirmedMsg, got %T", msg)
	}
}

func TestDetection_EnterInErrorState_EmitsConfirmedMsg(t *testing.T) {
	m := NewDetectionModel(stubDetectErr(errors.New("timeout")))

	newModel, _ := m.Update(detectionResultMsg{result: nil, err: errors.New("timeout")})
	m = newModel.(DetectionModel)

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = newModel

	if cmd == nil {
		t.Fatal("expected a cmd after enter in error state")
	}
	msg := cmd()
	if _, ok := msg.(DetectionConfirmedMsg); !ok {
		t.Fatalf("expected DetectionConfirmedMsg, got %T", msg)
	}
}

func TestDetection_EnterBeforeReady_IsNoop(t *testing.T) {
	m := NewDetectionModel(stubDetectOK(&system.DetectionResult{}))

	// ready is false; enter should produce no cmd
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(DetectionModel)

	if cmd != nil {
		t.Error("enter before ready should produce no cmd")
	}
	if m.ready {
		t.Error("model should still be in loading state")
	}
}

func TestDetection_QKey_EmitsQuit(t *testing.T) {
	m := NewDetectionModel(stubDetectOK(&system.DetectionResult{}))

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q key should produce a quit cmd")
	}
	// tea.Quit() returns tea.QuitMsg when called
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg from q key, got %T", msg)
	}
}

// TestDetection_Integration_RealDetect is an integration test that calls the real
// system.Detect function. It is skipped in -short mode.
func TestDetection_Integration_RealDetect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	fn := func(ctx context.Context) (*system.DetectionResult, error) {
		r, err := system.Detect(ctx)
		return &r, err
	}
	m := NewDetectionModel(fn)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() should return a non-nil cmd")
	}

	msg := cmd()
	rm, ok := msg.(detectionResultMsg)
	if !ok {
		t.Fatalf("expected detectionResultMsg, got %T", msg)
	}
	if rm.err != nil {
		t.Logf("detection error (acceptable in CI): %v", rm.err)
	}
}
