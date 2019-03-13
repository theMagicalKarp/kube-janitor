package main

import (
	"sort"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

func TestPendingJobs(t *testing.T) {
	client := fakeclientset.NewSimpleClientset()
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: "default",
		},
	}
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Minute * -70)},
			Name:              "my-pod",
			Labels:            map[string]string{"job-name": "a"},
		},
		Status: core.PodStatus{
			Phase: "Pending",
		},
	}
	_, err := client.CoreV1().Pods("").Create(pod)
	if err != nil {
		t.Errorf("error injecting pod add: %v", err)
	}

	removeCheck := PendingJobs(60, client)
	remove, err := removeCheck(job)

	if remove != true {
		t.Errorf("Expected true remove check")
		return
	}
}

func TestPendingJobsNoneExpired(t *testing.T) {
	client := fakeclientset.NewSimpleClientset()
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: "default",
		},
	}
	pod1 := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			Name:              "my-pod",
			Labels:            map[string]string{"job-name": "a"},
		},
		Status: core.PodStatus{
			Phase: "Pending",
		},
	}
	_, err := client.CoreV1().Pods("").Create(pod1)
	if err != nil {
		t.Errorf("error injecting pod add: %v", err)
	}
	pod2 := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			Name:              "my-pod2",
			Labels:            map[string]string{"job-name": "a"},
		},
		Status: core.PodStatus{
			Phase: "Succeeded",
		},
	}
	_, err = client.CoreV1().Pods("").Create(pod2)
	if err != nil {
		t.Errorf("error injecting pod add: %v", err)
	}

	removeCheck := PendingJobs(60, client)
	remove, err := removeCheck(job)

	if remove != false {
		t.Errorf("Expected false remove check")
		return
	}
}

func TestPendingJobsNoCreationTime(t *testing.T) {
	client := fakeclientset.NewSimpleClientset()
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: "default",
		},
	}
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "my-pod",
			Labels: map[string]string{"job-name": "a"},
		},
		Status: core.PodStatus{
			Phase: "Pending",
		},
	}
	_, err := client.CoreV1().Pods("").Create(pod)
	if err != nil {
		t.Errorf("error injecting pod add: %v", err)
	}

	removeCheck := PendingJobs(60, client)
	remove, err := removeCheck(job)

	if remove != false {
		t.Errorf("Expected false remove check")
		return
	}
}

func TestExpiredJobsNilCompletionTime(t *testing.T) {
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: "default",
		},
		Status: v1.JobStatus{
			CompletionTime: nil,
		},
	}

	removeCheck := ExpiredJobs(60, "foo.io")
	remove, _ := removeCheck(job)

	if remove != false {
		t.Errorf("Expected false remove check")
		return
	}
}

func TestExpiredJobsNilCompletionTimeConditions(t *testing.T) {
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: "default",
		},
		Status: v1.JobStatus{
			CompletionTime: nil,
			Conditions: []v1.JobCondition{
				{
					Type:   "Failed",
					Status: "True",
					Reason: "DeadlineExceeded",
				},
			},
		},
	}

	removeCheck := ExpiredJobs(60, "foo.io")
	remove, _ := removeCheck(job)

	if remove != false {
		t.Errorf("Expected false remove check")
		return
	}
}

func TestExpiredJobsNoneExpired(t *testing.T) {
	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now()},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * 30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour)},
			},
		},
	}

	removeCheck := ExpiredJobs(120, "foo.io")
	remove := true
	for _, job := range jobs {
		remove = true
		remove, _ = removeCheck(job)

		if remove != false {
			t.Errorf("Expected false remove check")
			return
		}
	}
}

func TestExpiredJobs(t *testing.T) {
	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now()},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -90)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "d",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -24)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "e",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "f",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				Conditions: []v1.JobCondition{
					{
						Status:        "True",
						Type:          "Failed",
						Reason:        "BackoffLimitExceeded",
						LastProbeTime: metav1.Time{Time: time.Now().Add(time.Hour * -168)},
					},
				},
			},
		},
	}

	removeCheck := ExpiredJobs(60, "foo.io")
	foundNames := []string{}
	remove := false
	for _, job := range jobs {
		remove, _ = removeCheck(job)
		if remove == true {
			foundNames = append(foundNames, job.ObjectMeta.Name)
		}
	}

	sort.Strings(foundNames)

	if strings.Join(foundNames, "|") != "c|d|e|f" {
		t.Errorf(
			"Expected (c|d|e|f) jobs but got (%s)",
			strings.Join(foundNames, "|"),
		)
		return
	}
}

func TestExpiredJobsCustomExpirations(t *testing.T) {
	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "15",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "90",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "15.05",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "d",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "45",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "e",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "45.05",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "f",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "this-is-not-a-valid-float",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "g",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/expiration": "this-is-not-a-valid-float",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Minute * -90)},
			},
		},
	}

	removeCheck := ExpiredJobs(60, "foo.io")
	foundNames := []string{}
	remove := false
	for _, job := range jobs {
		remove, _ = removeCheck(job)
		if remove == true {
			foundNames = append(foundNames, job.ObjectMeta.Name)
		}
	}

	sort.Strings(foundNames)

	if strings.Join(foundNames, "|") != "a|c|g" {
		t.Errorf(
			"Expected (a|c|g) jobs but got (%s)",
			strings.Join(foundNames, "|"),
		)
		return
	}
}
