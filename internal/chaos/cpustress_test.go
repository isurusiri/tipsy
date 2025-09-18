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

// Helper function to create test pods for CPU stress testing
func createTestPodsForCPUStress(count int, phase corev1.PodPhase) []corev1.Pod {
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

func TestInjectCPUStress_Success(t *testing.T) {
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
		method      string
		duration    time.Duration
		description string
	}{
		{
			name:        "inject CPU stress to running pods with stress-ng",
			podCount:    3,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			method:      "stress-ng",
			duration:    30 * time.Second,
			description: "Should inject CPU stress to all running pods using stress-ng",
		},
		{
			name:        "inject CPU stress to running pods with yes",
			podCount:    2,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			method:      "yes",
			duration:    60 * time.Second,
			description: "Should inject CPU stress to all running pods using yes",
		},
		{
			name:        "dry run mode with stress-ng",
			podCount:    2,
			podPhase:    corev1.PodRunning,
			dryRun:      true,
			method:      "stress-ng",
			duration:    60 * time.Second,
			description: "Should simulate CPU stress injection without actually doing it",
		},
		{
			name:        "dry run mode with yes",
			podCount:    1,
			podPhase:    corev1.PodRunning,
			dryRun:      true,
			method:      "yes",
			duration:    10 * time.Second,
			description: "Should simulate CPU stress injection with yes method",
		},
		{
			name:        "single pod with stress-ng",
			podCount:    1,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			method:      "stress-ng",
			duration:    10 * time.Second,
			description: "Should handle single pod correctly with stress-ng",
		},
		{
			name:        "single pod with yes",
			podCount:    1,
			podPhase:    corev1.PodRunning,
			dryRun:      false,
			method:      "yes",
			duration:    15 * time.Second,
			description: "Should handle single pod correctly with yes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods
			pods := createTestPodsForCPUStress(tc.podCount, tc.podPhase)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			// Execute InjectCPUStress
			_, err := InjectCPUStress(fakeClient, "default", "app=nginx", tc.method, tc.duration, tc.dryRun)

			// Check for errors
			if err != nil {
				t.Errorf("Unexpected error: %v - %s", err, tc.description)
			}
		})
	}
}

func TestInjectCPUStress_NoPodsFound(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create fake client with no pods
	fakeClient := fake.NewSimpleClientset()

	// Execute InjectCPUStress
	_, err := InjectCPUStress(fakeClient, "default", "app=nonexistent", "stress-ng", 30*time.Second, false)

	// Should not return an error, just log a warning
	if err != nil {
		t.Errorf("Unexpected error when no pods found: %v", err)
	}
}

func TestInjectCPUStress_NonRunningPods(t *testing.T) {
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
			pods := createTestPodsForCPUStress(2, tc.podPhase)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			// Execute InjectCPUStress
			_, err := InjectCPUStress(fakeClient, "default", "app=nginx", "stress-ng", 30*time.Second, false)

			// Should not return an error, just skip non-running pods
			if err != nil {
				t.Errorf("Unexpected error with %s: %v", tc.name, err)
			}
		})
	}
}

func TestInjectCPUStress_MixedPodPhases(t *testing.T) {
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

	// Execute InjectCPUStress
	_, err := InjectCPUStress(fakeClient, "default", "app=nginx", "stress-ng", 30*time.Second, false)

	// Should not return an error, should process running pods and skip others
	if err != nil {
		t.Errorf("Unexpected error with mixed pod phases: %v", err)
	}
}

func TestInjectCPUStress_DifferentNamespaces(t *testing.T) {
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

	// Execute InjectCPUStress on default namespace only
	_, err = InjectCPUStress(fakeClient, "default", "app=nginx", "stress-ng", 30*time.Second, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify only default namespace pod was processed
	// (We can't easily verify ephemeral containers were added with fake client,
	// but we can verify no errors occurred)
}

func TestInjectCPUStress_EdgeCases(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	testCases := []struct {
		name        string
		method      string
		duration    time.Duration
		expectError bool
		description string
	}{
		{
			name:        "zero duration with stress-ng",
			method:      "stress-ng",
			duration:    0,
			expectError: false,
			description: "Should handle zero duration with stress-ng",
		},
		{
			name:        "zero duration with yes",
			method:      "yes",
			duration:    0,
			expectError: false,
			description: "Should handle zero duration with yes",
		},
		{
			name:        "very short duration with stress-ng",
			method:      "stress-ng",
			duration:    1 * time.Millisecond,
			expectError: false,
			description: "Should handle very short duration with stress-ng",
		},
		{
			name:        "very short duration with yes",
			method:      "yes",
			duration:    1 * time.Millisecond,
			expectError: false,
			description: "Should handle very short duration with yes",
		},
		{
			name:        "very long duration with stress-ng",
			method:      "stress-ng",
			duration:    24 * time.Hour,
			expectError: false,
			description: "Should handle very long duration with stress-ng",
		},
		{
			name:        "very long duration with yes",
			method:      "yes",
			duration:    24 * time.Hour,
			expectError: false,
			description: "Should handle very long duration with yes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods
			pods := createTestPodsForCPUStress(1, corev1.PodRunning)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			_, err := InjectCPUStress(fakeClient, "default", "app=nginx", tc.method, tc.duration, false)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestInjectCPUStress_MethodValidation(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	testCases := []struct {
		name        string
		method      string
		expectError bool
		description string
	}{
		{
			name:        "valid stress-ng method",
			method:      "stress-ng",
			expectError: false,
			description: "Should accept stress-ng method",
		},
		{
			name:        "valid yes method",
			method:      "yes",
			expectError: false,
			description: "Should accept yes method",
		},
		{
			name:        "invalid method",
			method:      "invalid",
			expectError: true,
			description: "Should reject invalid method",
		},
		{
			name:        "empty method",
			method:      "",
			expectError: true,
			description: "Should reject empty method",
		},
		{
			name:        "uppercase stress-ng",
			method:      "STRESS-NG",
			expectError: true,
			description: "Should reject uppercase stress-ng (case sensitive)",
		},
		{
			name:        "uppercase yes",
			method:      "YES",
			expectError: true,
			description: "Should reject uppercase yes (case sensitive)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fake client for testing
			fakeClient := fake.NewSimpleClientset()
			
			// Test the individual function in dry-run mode to avoid pod creation issues
			err := injectCPUStressToPod(fakeClient, "default", "test-pod", tc.method, 30*time.Second, true)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestInjectCPUStressToPod_StressNGMethod(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create pods
	pods := createTestPodsForCPUStress(1, corev1.PodRunning)
	
	// Create fake client
	fakeClient := fake.NewSimpleClientset()
	
	// Add pods to the fake client
	for _, pod := range pods {
		_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create pod in fake client: %v", err)
		}
	}

	// Test stress-ng method
	err := injectCPUStressToPod(fakeClient, "default", "test-pod-1", "stress-ng", 30*time.Second, false)
	if err != nil {
		t.Errorf("Unexpected error with stress-ng method: %v", err)
	}
}

func TestInjectCPUStressToPod_YesMethod(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create pods
	pods := createTestPodsForCPUStress(1, corev1.PodRunning)
	
	// Create fake client
	fakeClient := fake.NewSimpleClientset()
	
	// Add pods to the fake client
	for _, pod := range pods {
		_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create pod in fake client: %v", err)
		}
	}

	// Test yes method
	err := injectCPUStressToPod(fakeClient, "default", "test-pod-1", "yes", 30*time.Second, false)
	if err != nil {
		t.Errorf("Unexpected error with yes method: %v", err)
	}
}

func TestInjectCPUStressToPod_DryRun(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create pods
	pods := createTestPodsForCPUStress(1, corev1.PodRunning)
	
	// Create fake client
	fakeClient := fake.NewSimpleClientset()
	
	// Add pods to the fake client
	for _, pod := range pods {
		_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create pod in fake client: %v", err)
		}
	}

	// Test dry run mode
	err := injectCPUStressToPod(fakeClient, "default", "test-pod-1", "stress-ng", 30*time.Second, true)
	if err != nil {
		t.Errorf("Unexpected error in dry run mode: %v", err)
	}
}

func TestMarshalCPUStressEphemeralContainer(t *testing.T) {
	// Test that marshalCPUStressEphemeralContainer returns valid JSON
	container := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    "test-container",
			Image:   "test-image:latest",
			Command: []string{"stress-ng", "--cpu", "1", "--timeout", "30s"},
		},
	}

	jsonStr := marshalCPUStressEphemeralContainer(container)
	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Basic validation that it contains expected fields
	if !containsString(jsonStr, "test-container") {
		t.Error("Expected JSON to contain container name")
	}
	if !containsString(jsonStr, "test-image:latest") {
		t.Error("Expected JSON to contain container image")
	}
	if !containsString(jsonStr, "stress-ng") {
		t.Error("Expected JSON to contain command")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && containsString(s[1:], substr)
}

// Benchmark tests
func BenchmarkInjectCPUStress_StressNG(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPodsForCPUStress(100, corev1.PodRunning)

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
		
		_, _ = InjectCPUStress(fakeClient, "default", "app=nginx", "stress-ng", 30*time.Second, false)
	}
}

func BenchmarkInjectCPUStress_Yes(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPodsForCPUStress(100, corev1.PodRunning)

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
		
		_, _ = InjectCPUStress(fakeClient, "default", "app=nginx", "yes", 30*time.Second, false)
	}
}

func TestInjectCPUStress_DryRunComprehensive(t *testing.T) {
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
		namespace   string
		selector    string
		method      string
		duration    time.Duration
		description string
	}{
		{
			name:        "dry run with stress-ng method",
			namespace:   "default",
			selector:    "app=nginx",
			method:      "stress-ng",
			duration:    30 * time.Second,
			description: "Should simulate CPU stress injection with stress-ng without API calls",
		},
		{
			name:        "dry run with yes method",
			namespace:   "production",
			selector:    "tier=frontend",
			method:      "yes",
			duration:    60 * time.Second,
			description: "Should simulate CPU stress injection with yes method in different namespace",
		},
		{
			name:        "dry run with long duration",
			namespace:   "default",
			selector:    "app=nginx",
			method:      "stress-ng",
			duration:    10 * time.Minute,
			description: "Should handle long durations in dry run",
		},
		{
			name:        "dry run with complex selector",
			namespace:   "default",
			selector:    "app=nginx,environment=staging",
			method:      "yes",
			duration:    10 * time.Second,
			description: "Should handle complex selectors in dry run",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fake client (but we won't add any pods since dry-run should avoid API calls)
			fakeClient := fake.NewSimpleClientset()

			// Execute InjectCPUStress in dry-run mode
			affectedPods, err := InjectCPUStress(fakeClient, tc.namespace, tc.selector, tc.method, tc.duration, true)

			// Should not return an error
			if err != nil {
				t.Errorf("Unexpected error in dry-run mode: %v - %s", err, tc.description)
			}

			// Should return empty slice in dry-run mode
			if len(affectedPods) != 0 {
				t.Errorf("Expected empty slice in dry-run mode, got %d pods - %s", len(affectedPods), tc.description)
			}

			// Verify no pods were actually created or modified
			pods, err := fakeClient.CoreV1().Pods(tc.namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Failed to list pods: %v", err)
			}

			// Should have no pods since we never created any
			if len(pods.Items) != 0 {
				t.Errorf("Expected no pods in fake client, got %d - %s", len(pods.Items), tc.description)
			}
		})
	}
}

func BenchmarkInjectCPUStress_DryRun(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPodsForCPUStress(100, corev1.PodRunning)

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
		
		_, _ = InjectCPUStress(fakeClient, "default", "app=nginx", "stress-ng", 30*time.Second, true)
	}
}
