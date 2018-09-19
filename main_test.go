package main

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func _validatorTrue(v1.Job) (bool, error) {
	return true, nil
}

func _validatorFalse(v1.Job) (bool, error) {
	return false, nil
}

func _validatorError(v1.Job) (bool, error) {
	// Return true so we can assert it is not removed if there is an error
	return true, errors.New("Error!")
}

func TestFindExpiredJobsNoJobsProvided(t *testing.T) {
	validatorsList := []JobValidator{}

	foundJobs := FindExpiredJobs([]v1.Job{}, "foo.io", validatorsList)

	if len(foundJobs) != 0 {
		t.Errorf("Expected 0 jobs but got %d jobs", len(foundJobs))
		return
	}
}

func TestFindExpiredJobsError(t *testing.T) {
	validatorsList := []JobValidator{_validatorError}

	foundJobs := FindExpiredJobs([]v1.Job{}, "foo.io", validatorsList)

	if len(foundJobs) != 0 {
		t.Errorf("Expected 0 jobs but got %d jobs", len(foundJobs))
		return
	}
}

func TestFindExpiredJobsIgnoreJobs(t *testing.T) {
	validatorsList := []JobValidator{_validatorTrue}

	jobs := []v1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "a",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Status: v1.JobStatus{
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
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
				CompletionTime: &metav1.Time{Time: time.Now().Add(time.Hour * -168)},
			},
		},
	}

	foundJobs := FindExpiredJobs(jobs, "foo.io", validatorsList)
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
