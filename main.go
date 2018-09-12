package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

func FindExpiredJobs(jobs []v1.Job, maxAge float64, annotation string) []v1.Job {
	expiredJobs := []v1.Job{}

	expirationAnnotationName := fmt.Sprintf("%s/expiration", annotation)
	ignoreAnnotationName := fmt.Sprintf("%s/ignore", annotation)

	log.Debugf("(%d) jobs found", len(jobs))

	for _, job := range jobs {
		ignoreAnnotaiton := job.ObjectMeta.Annotations[ignoreAnnotationName]
		ignore, err := strconv.ParseBool(ignoreAnnotaiton)
		if err == nil && ignore {
			log.Debugf(
				"Ignoring (%s:%s) with annotation (%s) of (%s)",
				job.ObjectMeta.Namespace,
				job.ObjectMeta.Name,
				ignoreAnnotationName,
				ignoreAnnotaiton,
			)
			continue
		}

		if job.Status.CompletionTime == nil {
			continue
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
				expiredJobs = append(expiredJobs, job)
			}
			continue
		}

		if age.Minutes() >= maxAge {
			expiredJobs = append(expiredJobs, job)
		}
	}

	log.Debugf("(%d) expired jobs found", len(expiredJobs))

	return expiredJobs
}

func main() {
	namespace := flag.String(
		"namespace",
		"",
		"Namespace to target when deleting jobs (by default all namespaces are targeted)",
	)
	expiration := flag.Float64(
		"expiration",
		60.0,
		"Expiration time on jobs (in minutes)",
	)
	dryrun := flag.Bool(
		"dryrun",
		false,
		"Logs what jobs will be deleted when fully ran",
	)
	annotation := flag.String(
		"annotation",
		"kube.janitor.io",
		"Annotation prefix to check when deleting jobs",
	)
	verbose := flag.Bool(
		"verbose",
		false,
		"Increase verbosity of logging",
	)
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	jobList, err := clientset.BatchV1().Jobs(*namespace).List(metav1.ListOptions{})

	if err != nil {
		log.Fatal(err.Error())
	}

	targetJobs := FindExpiredJobs(jobList.Items, *expiration, *annotation)

	deletePolicy := metav1.DeletePropagationForeground

	if *dryrun {
		log.Warnf("!!! DRY RUN (JOBS WON'T BE DISCARDED) !!!")
	}

	for _, job := range targetJobs {
		log.Infof("Deleting (%s:%s)", job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		if *dryrun {
			continue
		}

		jobClient := clientset.BatchV1().Jobs(job.ObjectMeta.Namespace)
		deletionOptions := &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		err = jobClient.Delete(job.ObjectMeta.Name, deletionOptions)

		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
