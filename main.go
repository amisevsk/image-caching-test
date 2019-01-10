package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const daemonsetNameEnvVar = "DAEMONSET_NAME"
const imagesEnvVar = "IMAGES"
const namespaceEnvVar = "NAMESPACE"
const intervalEnvVar = "CACHING_INTERVAL_HOURS"

var daemonsetName string
var namespace string
var cachingInterval int

var containerCommand = []string{"/bin/sh", "-c", "sleep 60"}
var propagationPolicy = metav1.DeletePropagationForeground
var terminationGracePeriodSeconds = int64(1)

// Process the images env var (strip whitespace/empty entries). Output is array of strings
// <imagename>=<image>
func processImagesEnvVar() []string {
	envVar := os.Getenv(imagesEnvVar)
	// TODO: error handling
	envVar = strings.TrimSpace(envVar)
	images := strings.Split(envVar, ";")
	for i, image := range images {
		images[i] = strings.TrimSpace(image)
	}
	// If last element is empty, remove it
	if images[len(images)-1] == "" {
		images = images[:len(images)-1]
	}
	return images
}

// Convenience function for making containers.
func makeContainer(name, image string) corev1.Container {
	return corev1.Container{
		Name:    name,
		Image:   image,
		Command: []string{"/bin/sh", "-c", "sleep 5"},
	}
}

// Get array of all images in containers to be cached.
func getContainers() []corev1.Container {
	images := processImagesEnvVar()
	containers := make([]corev1.Container, len(images))
	for i, image := range images {
		nameAndImage := strings.Split(image, "=")
		log.Println(nameAndImage)
		if len(nameAndImage) != 2 {
			log.Printf("Malformed image name/tag in %s env var: %s", imagesEnvVar, image)
			continue
		}
		containers[i] = makeContainer(nameAndImage[0], nameAndImage[1])
	}
	return containers
}

// Create the daemonset, using to-be-cached images as init containers.
func createDaemonset(clientset *kubernetes.Clientset) error {
	log.Printf("Creating daemonset\n")
	toCreate := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: daemonsetName,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test": "daemonset-test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"test": "daemonset-test",
					},
					Name: "test-po",
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					InitContainers:                getContainers(),
					Containers: []corev1.Container{corev1.Container{
						Name: "idle", Image: "centos", Command: containerCommand}},
				},
			},
		},
	}
	_, err := clientset.AppsV1().DaemonSets(namespace).Create(toCreate)
	if err != nil {
		log.Fatalf("Failed to create daemonset: %s", err.Error())
	} else {
		log.Printf("Created daemonset")
	}
	return err
}

// Wait for daemonset to be ready (MODIFIED event with all nodes scheduled)
func waitDaemonsetReady(clientset *kubernetes.Clientset, c <-chan watch.Event) {
	log.Printf("Waiting for daemonset to be ready")
	for ev := range c {
		log.Printf("(DEBUG) Create watch event received: %s\n", ev.Type)
		if ev.Type == watch.Modified {
			daemonset := ev.Object.(*appsv1.DaemonSet)
			// TODO: Not sure if this is the correct logic
			if daemonset.Status.NumberReady == daemonset.Status.DesiredNumberScheduled {
				log.Printf("All nodes scheduled in daemonset, returning")
				return
			}
		} else if ev.Type == watch.Deleted {
			log.Fatalf("Error occurred while waiting for daemonset to be ready -- event %s detected", watch.Deleted)
		}
	}
}

// Delete daemonset with metadata.name daemonsetName
func deleteDaemonset(clientset *kubernetes.Clientset) {
	log.Println("Deleting daemonset")
	err := clientset.AppsV1().DaemonSets(namespace).Delete(daemonsetName, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		log.Fatalf("Failed to delete daemonset %s", err.Error())
	} else {
		log.Printf("Deleted daemonset %s\n", daemonsetName)
	}
}

// Use watch channel to wait for DELETED event on daemonset, then return
func waitDaemonsetDeleted(clientset *kubernetes.Clientset, c <-chan watch.Event) {
	for ev := range c {
		log.Printf("(DEBUG) Delete watch event received: %s\n", ev.Type)
		if ev.Type == watch.Deleted {
			return
		}
	}
}

// Set up watch on daemonset
func watchDaemonset(clientset *kubernetes.Clientset) watch.Interface {
	watch, err := clientset.AppsV1().DaemonSets(namespace).Watch(metav1.ListOptions{
		FieldSelector:        fmt.Sprintf("metadata.name=%s", daemonsetName),
		IncludeUninitialized: true,
	})
	if err != nil {
		log.Fatalf("Failed to set up watch on daemonsets: %s", err.Error())
	}
	return watch
}

// Check if daemonset with daemonsetName exists, and if so, delete it.
func checkIfDaemonsetExists(clientset *kubernetes.Clientset) {
	daemonset, err := clientset.AppsV1().DaemonSets(namespace).Get(daemonsetName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting daemonset %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		log.Panicf(err.Error())
	}
	if daemonset != nil {
		deleteDaemonset(clientset)
		log.Printf("Deleted existing daemonset")
	}
}

func cacheImages(clientset *kubernetes.Clientset) {
	log.Printf("Starting caching process")
	// Set up watch on daemonsets
	dsWatch := watchDaemonset(clientset)
	// Create daemonset, wait for it to be ready, and then delete it.
	watchChan := dsWatch.ResultChan()
	createDaemonset(clientset)
	waitDaemonsetReady(clientset, watchChan)
	time.Sleep(30)
	deleteDaemonset(clientset)
	waitDaemonsetDeleted(clientset, watchChan)
	dsWatch.Stop()
	log.Printf("Done caching images")
}

func main() {
	log.SetOutput(os.Stdout)
	processEnvVars()

	// Set up kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf(err.Error())
	}

	// Clean up existing deployment if necessary
	checkIfDaemonsetExists(clientset)

	for {
		cacheImages(clientset)
		time.Sleep(time.Duration(cachingInterval) * time.Minute)
	}
}

// Check that all required env vars are set, and convert caching interval env var to
// an int.
func processEnvVars() {
	badEnvVar := false
	if daemonsetName = os.Getenv(daemonsetNameEnvVar); daemonsetName == "" {
		log.Printf("Env var %s unset. Aborting", daemonsetNameEnvVar)
		badEnvVar = true
	}
	if namespace = os.Getenv(namespaceEnvVar); namespace == "" {
		log.Printf("Env var %s unset. Aborting", namespaceEnvVar)
		badEnvVar = true
	}
	if cachingIntervalStr := os.Getenv(intervalEnvVar); cachingIntervalStr == "" {
		log.Printf("Env var %s unset. Aborting", intervalEnvVar)
		badEnvVar = true
	} else {
		interval, err := strconv.Atoi(cachingIntervalStr)
		if err != nil {
			log.Printf("Could not parse env var %s to integer. Value is %s", intervalEnvVar, cachingIntervalStr)
			badEnvVar = true
		} else {
			cachingInterval = interval
		}
	}
	if badEnvVar {
		os.Exit(1)
	}
}
