package main

import (
	"fmt"
	"strconv"
	"time"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
)

type JobValidator func(v1.Job) (bool, error)

func PendingJobs(maxAge float64, kubeClient kubernetes.Interface) JobValidator {
	return func(job v1.Job) (bool, error) {
		podList, err := kubeClient.CoreV1().Pods("").List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", job.ObjectMeta.Name),
		})

		if err != nil {
			return false, err
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == "Pending" {
				if pod.ObjectMeta.CreationTimestamp.Time.IsZero() {
					continue
				}
				age := time.Since(pod.ObjectMeta.CreationTimestamp.Time)
				if age.Minutes() >= maxAge {
					return true, nil
				}
			}
		}

		return false, nil
	}
}

func ExpiredJobs(maxAge float64, annotation string) JobValidator {
	expirationAnnotationName := fmt.Sprintf("%s/expiration", annotation)

	return func(job v1.Job) (bool, error) {
		var age time.Duration
		if job.Status.CompletionTime == nil {
			for _, condition := range job.Status.Conditions {
				if condition.Reason == "BackoffLimitExceeded" && condition.Status == "True" {
					age = time.Since(condition.LastProbeTime.Time)
					break
				}
			}
		} else {
			age = time.Since(job.Status.CompletionTime.Time)
		}

		if age == 0 {
			return false, nil
		}

		expirationAnnotation := job.ObjectMeta.Annotations[expirationAnnotationName]

		maxAgeOverride, err := strconv.ParseFloat(expirationAnnotation, 64)
		if err == nil {
			log.Debugf(
				"Expiration override for (%s:%s) with annotation (%s) of (%s)",
				job.ObjectMeta.Namespace,
				job.ObjectMeta.Name,
				expirationAnnotationName,
				expirationAnnotation,
			)
			if age.Minutes() >= maxAgeOverride {
				return true, nil
			}
			return false, nil
		}

		return age.Minutes() >= maxAge, nil
	}
}
