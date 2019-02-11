package configuration

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Env vars used for configuration
const (
	intervalEnvVar             = "CACHING_INTERVAL_HOURS"
	daemonsetNameEnvVar        = "DAEMONSET_NAME"
	namespaceEnvVar            = "NAMESPACE"
	impersonateUsersEnvVar     = "IMPERSONATE_USERS"
	proxyURLEnvVar             = "OPENSHIFT_PROXY_URL"
	imagesEnvVar               = "IMAGES"
	serviceAccountIDEnvVar     = "SERVICE_ACCOUNT_ID"
	serviceAccountSecretEnvVar = "SERVICE_ACCOUNT_SECRET"
	oidcProviderEnvVar         = "OIDC_PROVIDER"
)

func getCachingInterval() int {
	cachingIntervalStr := getEnvVarOrExit(intervalEnvVar)
	interval, err := strconv.Atoi(cachingIntervalStr)
	if err != nil {
		log.Fatalf("Could not parse env var %s to integer. Value is %s", intervalEnvVar, cachingIntervalStr)
	}
	return interval
}

func processImagesEnvVar() map[string]string {
	rawImages := getEnvVarOrExit(imagesEnvVar)
	rawImages = strings.TrimSpace(rawImages)
	images := strings.Split(rawImages, ";")
	for i, image := range images {
		images[i] = strings.TrimSpace(image)
	}
	// If last element is empty, remove it
	if images[len(images)-1] == "" {
		images = images[:len(images)-1]
	}

	log.Printf("Processing images from configuration...")
	var imagesMap = make(map[string]string)
	for _, image := range images {
		log.Printf("Image: %s", image)
		nameAndImage := strings.Split(image, "=")
		if len(nameAndImage) != 2 {
			log.Printf("Malformed image name/tag: %s. Ignoring.", image)
			continue
		}
		imagesMap[nameAndImage[0]] = nameAndImage[1]
	}
	return imagesMap
}

func processImpersonateUsers() []string {
	rawUsers := getEnvVarOrExit(impersonateUsersEnvVar)
	users := strings.Split(rawUsers, ",")
	if len(users) == 0 {
		log.Fatalf("No users found in env var %s", impersonateUsersEnvVar)
	}
	return users
}

func getEnvVarOrExit(envVar string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Fatalf("Env var %s unset. Aborting", envVar)
	}
	return val
}
