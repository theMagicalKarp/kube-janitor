package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

func FindExpiredJobs(jobList []v1.Job, annotation string, validatorList []func(v1.Job) bool) []v1.Job {
	expiredJobs := []v1.Job{}

	ignoreAnnotationName := fmt.Sprintf("%s/ignore", annotation)

	log.Debugf("(%d) jobs found", len(jobList))

	for _, job := range jobList {
		ignoreAnnotation := job.ObjectMeta.Annotations[ignoreAnnotationName]
		ignore, err := strconv.ParseBool(ignoreAnnotation)
		if err == nil && ignore {
			log.Debugf(
				"Ignoring (%s:%s) with annotation (%s) of (%s)",
				job.ObjectMeta.Namespace,
				job.ObjectMeta.Name,
				ignoreAnnotationName,
				ignoreAnnotation,
			)
			continue
		}

		for _, removeCheck := range validatorList {
			if removeCheck(job) == true {
				expiredJobs = append(expiredJobs, job)
				break
			}
		}
	}

	log.Debugf("(%d) jobs to remove", len(expiredJobs))

	return expiredJobs
}

func main() {
	annotation := flag.String(
		"annotation",
		"kube.janitor.io",
		"Annotation prefix to check when deleting jobs",
	)
	dryrun := flag.Bool(
		"dryrun",
		false,
		"Logs what jobs will be deleted when fully ran",
	)
	expiration := flag.Float64(
		"expiration",
		60.0,
		"Expiration time on jobs (in minutes)",
	)
	namespace := flag.String(
		"namespace",
		"",
		"Namespace to target when deleting jobs (by default all namespaces are targeted)",
	)
	pendingJobExpiration := flag.Float64(
		"pendingJobExpiration",
		-1.0,
		`Set the time (in minutes) that jobs will be removed if they are still in the pending state.
        By default, jobs stuck in a pending state are not removed`,
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

	validatorList := []func(v1.Job) bool{}
	validatorList = append(validatorList, ExpiredJobs(*expiration, *annotation))
	if *pendingJobExpiration > -1 {
		validatorList = append(validatorList, PendingJobs(*pendingJobExpiration, clientset))
	}

	if *dryrun {
		log.Warnf("!!! DRY RUN (JOBS WON'T BE DISCARDED) !!!")
	}

	targetJobs := FindExpiredJobs(jobList.Items, *annotation, validatorList)

	deletePolicy := metav1.DeletePropagationForeground
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
