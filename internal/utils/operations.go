package utils

import (
	"fmt"
	"log"

	conf "github.com/amisevsk/image-caching-test/internal/configuration"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CacheImages creates the daemonset responsible for ensuring images are cached
func CacheImages(clientset *kubernetes.Clientset) {
	log.Printf("Starting caching process")
	// Create daemonset, wait for it to be ready
	createDaemonset(clientset)
	log.Printf("Daemonset ready.")
}

// EnsureDaemonsetExists checks that the daemonset is still present, and
// recreates it if necessary
func EnsureDaemonsetExists(clientset *kubernetes.Clientset) {
	log.Printf("Checking that daemonset exists.")
	daemonset, err :=
		clientset.
			AppsV1().
			DaemonSets(conf.Config.Namespace).
			Get(conf.Config.DaemonsetName, metav1.GetOptions{})
	if err != nil || daemonset == nil {
		log.Printf("Recreating daemonset due to error")
		DeleteDaemonsetIfExists(clientset)
		CacheImages(clientset)
	}
}

// DeleteDaemonsetIfExists first checks if the daemonset exists, and deletes
// it if it does. Useful for ensuring no daemonset is already present from a
// previous rollout.
func DeleteDaemonsetIfExists(clientset *kubernetes.Clientset) {
	daemonset, err :=
		clientset.
			AppsV1().
			DaemonSets(conf.Config.Namespace).
			Get(conf.Config.DaemonsetName, metav1.GetOptions{})
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
