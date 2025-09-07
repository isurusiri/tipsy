package chaos

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildEndpointSubsetsFromPods(t *testing.T) {
	// Test data
	servicePorts := []corev1.ServicePort{
		{
			Name:     "http",
			Port:     80,
			Protocol: corev1.ProtocolTCP,
		},
		{
			Name:     "https",
			Port:     443,
			Protocol: corev1.ProtocolTCP,
		},
	}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "default",
				UID:       "pod1-uid",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				PodIP: "10.0.0.1",
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "default",
				UID:       "pod2-uid",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				PodIP: "10.0.0.2",
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod3",
				Namespace: "default",
				UID:       "pod3-uid",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				PodIP: "10.0.0.3",
			},
		},
	}

	// Test the function
	subsets, err := buildEndpointSubsetsFromPods(pods, servicePorts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify we have 2 subsets (ready and not ready)
	if len(subsets) != 2 {
		t.Fatalf("Expected 2 subsets, got %d", len(subsets))
	}

	// Find ready and not ready subsets
	var readySubset, notReadySubset *corev1.EndpointSubset
	for i := range subsets {
		if len(subsets[i].Addresses) > 0 {
			readySubset = &subsets[i]
		}
		if len(subsets[i].NotReadyAddresses) > 0 {
			notReadySubset = &subsets[i]
		}
	}

	// Verify ready subset
	if readySubset == nil {
		t.Fatal("Expected ready subset not found")
	}
	if len(readySubset.Addresses) != 1 {
		t.Fatalf("Expected 1 ready address, got %d", len(readySubset.Addresses))
	}
	if readySubset.Addresses[0].IP != "10.0.0.1" {
		t.Fatalf("Expected IP 10.0.0.1, got %s", readySubset.Addresses[0].IP)
	}
	if len(readySubset.Ports) != 2 {
		t.Fatalf("Expected 2 ports, got %d", len(readySubset.Ports))
	}

	// Verify not ready subset
	if notReadySubset == nil {
		t.Fatal("Expected not ready subset not found")
	}
	if len(notReadySubset.NotReadyAddresses) != 1 {
		t.Fatalf("Expected 1 not ready address, got %d", len(notReadySubset.NotReadyAddresses))
	}
	if notReadySubset.NotReadyAddresses[0].IP != "10.0.0.2" {
		t.Fatalf("Expected IP 10.0.0.2, got %s", notReadySubset.NotReadyAddresses[0].IP)
	}
}

func TestBuildEndpointSubsetsFromPodsEmptyServicePorts(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1",
				UID:  "pod1-uid",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				PodIP: "10.0.0.1",
			},
		},
	}

	// Test with empty service ports
	subsets, err := buildEndpointSubsetsFromPods(pods, []corev1.ServicePort{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return empty subsets
	if len(subsets) != 0 {
		t.Fatalf("Expected 0 subsets for empty service ports, got %d", len(subsets))
	}
}

func TestBuildEndpointSubsetsFromPodsNoRunningPods(t *testing.T) {
	servicePorts := []corev1.ServicePort{
		{
			Name:     "http",
			Port:     80,
			Protocol: corev1.ProtocolTCP,
		},
	}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1",
				UID:  "pod1-uid",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				PodIP: "10.0.0.1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2",
				UID:  "pod2-uid",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				PodIP: "", // No IP
			},
		},
	}

	// Test with no running pods
	subsets, err := buildEndpointSubsetsFromPods(pods, servicePorts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return empty subsets
	if len(subsets) != 0 {
		t.Fatalf("Expected 0 subsets for no running pods, got %d", len(subsets))
	}
}
