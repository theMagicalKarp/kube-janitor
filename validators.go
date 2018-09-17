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

func PendingJobs(maxAge float64, kubeClient kubernetes.Interface) func(v1.Job) bool {
	return func(job v1.Job) bool {
		podList, err := kubeClient.CoreV1().Pods("").List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", job.ObjectMeta.Name),
		})

		if err != nil {
			log.Fatal(err.Error())
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == "Pending" {
				if pod.ObjectMeta.CreationTimestamp.Time.IsZero() {
					continue
				}
				age := time.Since(pod.ObjectMeta.CreationTimestamp.Time)
				if age.Minutes() >= maxAge {
					return true
				}
			}
		}

		return false
	}
}

func ExpiredJobs(maxAge float64, annotation string) func(v1.Job) bool {
	expirationAnnotationName := fmt.Sprintf("%s/expiration", annotation)

	return func(job v1.Job) bool {
		if job.Status.CompletionTime == nil {
			return false
		}

		age := time.Since(job.Status.CompletionTime.Time)
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
				return true
			}
			return false
		}

		if age.Minutes() >= maxAge {
			return true
		}

		return false
	}
}
