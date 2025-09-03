package chaos

import (
	"context"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/isurusiri/tipsy/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Helper function to create test pods
func createTestPods(count int) []corev1.Pod {
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
				Phase: corev1.PodRunning,
			},
		}
	}
	return pods
}

func TestKillPods_Success(t *testing.T) {
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
		name           string
		podCount       int
		killCount      int
		dryRun         bool
		expectedRemaining int
		description    string
	}{
		{
			name:           "kill one pod from three",
			podCount:       3,
			killCount:      1,
			dryRun:         false,
			expectedRemaining: 2,
			description:    "Should delete exactly one pod",
		},
		{
			name:           "kill two pods from three",
			podCount:       3,
			killCount:      2,
			dryRun:         false,
			expectedRemaining: 1,
			description:    "Should delete exactly two pods",
		},
		{
			name:           "kill more pods than available",
			podCount:       2,
			killCount:      5,
			dryRun:         false,
			expectedRemaining: 0,
			description:    "Should delete all available pods when count exceeds available",
		},
		{
			name:           "dry run mode",
			podCount:       3,
			killCount:      2,
			dryRun:         true,
			expectedRemaining: 3,
			description:    "Should not delete any pods in dry run mode",
		},
		{
			name:           "zero count",
			podCount:       3,
			killCount:      0,
			dryRun:         false,
			expectedRemaining: 3,
			description:    "Should not delete any pods with zero count",
		},
		{
			name:           "negative count",
			podCount:       3,
			killCount:      -1,
			dryRun:         false,
			expectedRemaining: 3,
			description:    "Should not delete any pods with negative count",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods
			pods := createTestPods(tc.podCount)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			// Execute KillPods
			err := KillPods(fakeClient, "default", "app=nginx", tc.killCount, tc.dryRun)

			// Check for errors
			if err != nil {
				t.Errorf("Unexpected error: %v - %s", err, tc.description)
			}

			// Verify remaining pods
			remainingPods, err := fakeClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Failed to list remaining pods: %v", err)
			}

			if len(remainingPods.Items) != tc.expectedRemaining {
				t.Errorf("Expected %d pods remaining, got %d - %s", tc.expectedRemaining, len(remainingPods.Items), tc.description)
			}
		})
	}
}

func TestKillPods_NoPodsFound(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create fake client with no pods
	fakeClient := fake.NewSimpleClientset()

	// Execute KillPods
	err := KillPods(fakeClient, "default", "app=nonexistent", 1, false)

	// Should not return an error, just log a warning
	if err != nil {
		t.Errorf("Unexpected error when no pods found: %v", err)
	}

	// Verify no pods exist
	remainingPods, err := fakeClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list remaining pods: %v", err)
	}

	if len(remainingPods.Items) != 0 {
		t.Errorf("Expected no pods remaining, got %d", len(remainingPods.Items))
	}
}

func TestKillPods_RandomSelection(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create multiple pods
	pods := createTestPods(5)
	
	// Run multiple times to test randomness
	selectedPods := make(map[string]int)
	iterations := 50

	for i := 0; i < iterations; i++ {
		// Create fresh fake client for each iteration
		fakeClient := fake.NewSimpleClientset()
		
		// Add pods to the fake client
		for _, pod := range pods {
			_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Failed to create pod in fake client: %v", err)
			}
		}

		// Execute KillPods
		err := KillPods(fakeClient, "default", "app=nginx", 2, false)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Check which pods remain
		remainingPods, err := fakeClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list remaining pods: %v", err)
		}

		// Count which pods were deleted (not in remaining list)
		remainingNames := make(map[string]bool)
		for _, pod := range remainingPods.Items {
			remainingNames[pod.Name] = true
		}

		for _, pod := range pods {
			if !remainingNames[pod.Name] {
				selectedPods[pod.Name]++
			}
		}
	}

	// Check that all pods have a chance to be selected
	// With 5 pods and selecting 2, each pod should be selected roughly 40% of the time
	// We'll use a generous range to account for randomness
	expectedMin := iterations * 2 / 5 / 2 // Roughly 20% of iterations
	expectedMax := iterations * 2 / 5 * 2 // Roughly 80% of iterations

	for podName, count := range selectedPods {
		if count < expectedMin || count > expectedMax {
			t.Errorf("Pod %s selected %d times, expected between %d and %d", podName, count, expectedMin, expectedMax)
		}
	}
}

func TestKillPods_EdgeCases(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	testCases := []struct {
		name        string
		podCount    int
		killCount   int
		expectError bool
		description string
	}{
		{
			name:        "single pod",
			podCount:    1,
			killCount:   1,
			expectError: false,
			description: "Should handle single pod correctly",
		},
		{
			name:        "count equals available pods",
			podCount:    3,
			killCount:   3,
			expectError: false,
			description: "Should delete all available pods when count matches",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pods
			pods := createTestPods(tc.podCount)
			
			// Create fake client
			fakeClient := fake.NewSimpleClientset()
			
			// Add pods to the fake client
			for _, pod := range pods {
				_, err := fakeClient.CoreV1().Pods("default").Create(context.TODO(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod in fake client: %v", err)
				}
			}

			err := KillPods(fakeClient, "default", "app=nginx", tc.killCount, false)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestKillPods_DifferentNamespaces(t *testing.T) {
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

	// Execute KillPods on default namespace only
	err = KillPods(fakeClient, "default", "app=nginx", 1, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify only default namespace pod was deleted
	defaultPods, err := fakeClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list default namespace pods: %v", err)
	}

	productionPods, err := fakeClient.CoreV1().Pods("production").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list production namespace pods: %v", err)
	}

	if len(defaultPods.Items) != 0 {
		t.Errorf("Expected 0 pods in default namespace, got %d", len(defaultPods.Items))
	}

	if len(productionPods.Items) != 1 {
		t.Errorf("Expected 1 pod in production namespace, got %d", len(productionPods.Items))
	}
}

// Benchmark tests
func BenchmarkKillPods(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPods(100)

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
		
		KillPods(fakeClient, "default", "app=nginx", 10, false)
	}
}

func BenchmarkKillPods_DryRun(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Create test pods
	pods := createTestPods(100)

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
		
		KillPods(fakeClient, "default", "app=nginx", 10, true)
	}
}