package chaos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/isurusiri/tipsy/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Helper function to create test pods for latency testing
func createTestPodsForLatency(count int, phase corev1.PodPhase) []corev1.Pod {
	pods := make([]corev1.Pod, count)
	for i := 0; i < count; i++ {
		pods[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("test-pod-%d", i+1),
				Namespace: "default",
				Labels: map[string]string{
					"app": "nginx",
				},
			},
			Status: corev1.PodStatus{
				Phase: phase,
			},
		}
	}
	return pods
}

func TestInjectLatency_Success(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Reset global config
	originalConfig := config.GlobalConfig
	defer func() {
		config.GlobalConfig = originalConfig
	}()

	testCases := []struct {
		name        string
		podCount    int
		podPhase    corev1.PodPhase
		dryRun      bool
		delay       string
		duration    time.Duration
		description string
	}{
		{
			name:        "inject latency to running pods",
			podCount:    3,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			delay:       "200ms",
			duration:    30 * time.Second,
			description: "Should inject latency to all running pods",
		},
		{
			name:        "dry run mode",
			podCount:    2,
			podPhase:    corev1.PodRunning,
			dryRun:      true,
			delay:       "500ms",
			duration:    60 * time.Second,
			description: "Should simulate latency injection without actually doing it",
		},
		{
			name:        "single pod",
			podCount:    1,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			delay:       "100ms",
			duration:    10 * time.Second,
			description: "Should handle single pod correctly",
		},
		{
			name:        "custom delay",
			podCount:    2,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			delay:       "1s",
			duration:    45 * time.Second,
			description: "Should handle custom delay values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods
			pods := createTestPodsForLatency(tc.podCount, tc.podPhase)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			// Execute InjectLatency
			err := InjectLatency(fakeClient, "default", "app=nginx", tc.delay, tc.duration, tc.dryRun)

			// Check for errors
			if err != nil {
				t.Errorf("Unexpected error: %v - %s", err, tc.description)
			}
		})
	}
}

func TestInjectLatency_NoPodsFound(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create fake client with no pods
	fakeClient := fake.NewSimpleClientset()

	// Execute InjectLatency
	err := InjectLatency(fakeClient, "default", "app=nonexistent", "200ms", 30*time.Second, false)

	// Should not return an error, just log a warning
	if err != nil {
		t.Errorf("Unexpected error when no pods found: %v", err)
	}
}

func TestInjectLatency_NonRunningPods(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	testCases := []struct {
		name     string
		podPhase corev1.PodPhase
	}{
		{"pending pods", corev1.PodPending},
		{"succeeded pods", corev1.PodSucceeded},
		{"failed pods", corev1.PodFailed},
		{"unknown pods", corev1.PodUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods with non-running phase
			pods := createTestPodsForLatency(2, tc.podPhase)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			// Execute InjectLatency
			err := InjectLatency(fakeClient, "default", "app=nginx", "200ms", 30*time.Second, false)

			// Should not return an error, just skip non-running pods
			if err != nil {
				t.Errorf("Unexpected error with %s: %v", tc.name, err)
			}
		})
	}
}

func TestInjectLatency_MixedPodPhases(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Create pods with different phases
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "running-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "nginx"},
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pending-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "nginx"},
			},
			Status: corev1.PodStatus{Phase: corev1.PodPending},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "failed-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "nginx"},
			},
			Status: corev1.PodStatus{Phase: corev1.PodFailed},
		},
	}

	// Add pods to the fake client
	for _, pod := range pods {
		_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create pod in fake client: %v", err)
		}
	}

	// Execute InjectLatency
	err := InjectLatency(fakeClient, "default", "app=nginx", "200ms", 30*time.Second, false)

	// Should not return an error, should process running pods and skip others
	if err != nil {
		t.Errorf("Unexpected error with mixed pod phases: %v", err)
	}
}

func TestInjectLatency_DifferentNamespaces(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Create pods in different namespaces
	pod1 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
			Labels:    map[string]string{"app": "nginx"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}

	pod2 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "production",
			Labels:    map[string]string{"app": "nginx"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}

	// Add pods to different namespaces
	_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pod in default namespace: %v", err)
	}

	_, err = fakeClient.CoreV1().Pods("production").Create(context.TODO(), &pod2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pod in production namespace: %v", err)
	}

	// Execute InjectLatency on default namespace only
	err = InjectLatency(fakeClient, "default", "app=nginx", "200ms", 30*time.Second, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify only default namespace pod was processed
	// (We can't easily verify ephemeral containers were added with fake client,
	// but we can verify no errors occurred)
}

func TestInjectLatency_EdgeCases(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	testCases := []struct {
		name        string
		delay       string
		duration    time.Duration
		expectError bool
		description string
	}{
		{
			name:        "zero duration",
			delay:       "200ms",
			duration:    0,
			expectError: false,
			description: "Should handle zero duration",
		},
		{
			name:        "very short duration",
			delay:       "100ms",
			duration:    1 * time.Millisecond,
			expectError: false,
			description: "Should handle very short duration",
		},
		{
			name:        "very long duration",
			delay:       "500ms",
			duration:    24 * time.Hour,
			expectError: false,
			description: "Should handle very long duration",
		},
		{
			name:        "empty delay",
			delay:       "",
			duration:    30 * time.Second,
			expectError: false,
			description: "Should handle empty delay string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods
			pods := createTestPodsForLatency(1, corev1.PodRunning)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			err := InjectLatency(fakeClient, "default", "app=nginx", tc.delay, tc.duration, false)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestGetMainProcessName(t *testing.T) {
	// Test that getMainProcessName returns a non-empty string
	processName := getMainProcessName()
	if processName == "" {
		t.Error("Expected non-empty process name")
	}
}

func TestMarshalEphemeralContainer(t *testing.T) {
	// Test that marshalEphemeralContainer returns valid JSON
	container := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  "test-container",
			Image: "test-image:latest",
			Command: []string{"sh", "-c", "echo test"},
		},
	}

	jsonStr := marshalEphemeralContainer(container)
	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Basic validation that it contains expected fields
	if !contains(jsonStr, "test-container") {
		t.Error("Expected JSON to contain container name")
	}
	if !contains(jsonStr, "test-image:latest") {
		t.Error("Expected JSON to contain container image")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && contains(s[1:], substr)
}

// Benchmark tests
func BenchmarkInjectLatency(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPodsForLatency(100, corev1.PodRunning)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh fake client for each iteration
		fakeClient := fake.NewSimpleClientset()
		
		// Add pods to the fake client
		for _, pod := range pods {
			_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
			if err != nil {
				b.Fatalf("Failed to create pod in fake client: %v", err)
			}
		}
		
		InjectLatency(fakeClient, "default", "app=nginx", "200ms", 30*time.Second, false)
	}
}

func BenchmarkInjectLatency_DryRun(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPodsForLatency(100, corev1.PodRunning)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh fake client for each iteration
		fakeClient := fake.NewSimpleClientset()
		
		// Add pods to the fake client
		for _, pod := range pods {
			_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
			if err != nil {
				b.Fatalf("Failed to create pod in fake client: %v", err)
			}
		}
		
		InjectLatency(fakeClient, "default", "app=nginx", "200ms", 30*time.Second, true)
	}
}
