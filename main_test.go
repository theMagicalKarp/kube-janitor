package main

import (
	"sort"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNoJobsProvided(t *testing.T) {
	foundJobs := FindExpiredJobs([]v1.Job{}, 60, "foo.io")

	if len(foundJobs) != 0 {
		t.Errorf("Expected 0 jobs but got %d jobs", len(foundJobs))
		return
	}
}

func TestNilCompletionTime(t *testing.T) {
	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: nil,
			},
		},
	}

	foundJobs := FindExpiredJobs(jobs, 60, "foo.io")

	if len(foundJobs) != 0 {
		t.Errorf("Expected 0 jobs but got %d jobs", len(foundJobs))
		return
	}
}

func TestNoExpiredJobs(t *testing.T) {
	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now()},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * 30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour)},
			},
		},
	}

	foundJobs := FindExpiredJobs(jobs, 120, "foo.io")

	if len(foundJobs) != 0 {
		t.Errorf("Expected 0 jobs but got %d jobs", len(foundJobs))
		return
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
				CompletionTime: &metav1.Time{time.Now()},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -90)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "d",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -24)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "e",
				Namespace: "default",
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
	}

	foundJobs := FindExpiredJobs(jobs, 60, "foo.io")
	foundNames := []string{}
	for _, job := range foundJobs {
		foundNames = append(foundNames, job.ObjectMeta.Name)
	}

	sort.Strings(foundNames)

	if strings.Join(foundNames, "|") != "c|d|e" {
		t.Errorf(
			"Expected (c|d|e) jobs but got (%s)",
			strings.Join(foundNames, "|"),
		)
		return
	}
}

func TestIgnoreJobs(t *testing.T) {
	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "a",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/ignore": "true",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/ignore": "false",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "d",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/ignore": "TRUE",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "e",
				Namespace: "default",
				Annotations: map[string]string{
					"bar.io/ignore": "TRUE",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "f",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/ignore": "1",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "g",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/ignore": "0",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "h",
				Namespace: "default",
				Annotations: map[string]string{
					"foo.io/ignore": "this-is-not-a-valid-bool",
				},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{time.Now().Add(time.Hour * -168)},
			},
		},
	}

	foundJobs := FindExpiredJobs(jobs, 60, "foo.io")
	foundNames := []string{}
	for _, job := range foundJobs {
		foundNames = append(foundNames, job.ObjectMeta.Name)
	}

	sort.Strings(foundNames)

	if strings.Join(foundNames, "|") != "a|c|e|g|h" {
		t.Errorf(
			"Expected (a|c|e|g|h) jobs but got (%s)",
			strings.Join(foundNames, "|"),
		)
		return
	}
}

func TestCustomExpirations(t *testing.T) {
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -30)},
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
				CompletionTime: &metav1.Time{time.Now().Add(time.Minute * -90)},
			},
		},
	}

	foundJobs := FindExpiredJobs(jobs, 60, "foo.io")
	foundNames := []string{}
	for _, job := range foundJobs {
		foundNames = append(foundNames, job.ObjectMeta.Name)
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
