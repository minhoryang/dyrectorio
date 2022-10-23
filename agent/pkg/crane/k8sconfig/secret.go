package k8sconfig

import (
	"context"
	"fmt"
	"log"

	config "github.com/dyrector-io/dyrectorio/agent/internal/config"
	craneConfig "github.com/dyrector-io/dyrectorio/agent/pkg/crane/config"
	"github.com/dyrector-io/dyrectorio/agent/pkg/crane/k8s"

	"k8s.io/apimachinery/pkg/api/errors"
)

const SecretFileName = "private.key"

func TryToGetValidSecret(ctx context.Context, namespace, name string, appConfig *craneConfig.Configuration) (string, error) {
	if namespace == "" || name == "" {
		return "", fmt.Errorf("secret name and namespace can't be empty")
	}

	if err := isNamespaceExisted(namespace, appConfig); err != nil {
		return "", err
	}

	secretFiles, version, err := k8s.GetSecret(ctx, namespace, name, appConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			return addValidSecret(ctx, namespace, name, appConfig)
		}
		return "", fmt.Errorf("k8s stored secret %s/%s was on error: %w", namespace, name, err)
	}

	var foundSecret = ""
	for secretFileName, secretFileContent := range secretFiles {
		if secretFileName == SecretFileName {
			foundSecret = string(secretFileContent)
			break
		}
	}
	if foundSecret == "" {
		return "", fmt.Errorf("k8s stored secret %s/%s was empty (resourceVersion: %s)", namespace, name, version)
	}

	isExpired, err := config.IsExpiredKey(foundSecret)
	if err != nil {
		return "", fmt.Errorf("handling k8s stored secret %s/%s was on error: %w", namespace, name, err)
	}
	if isExpired {
		log.Printf("k8s stored secret %s/%s was expired (resourceVersion: %s), so renewing...", namespace, name, version)
		return addValidSecret(ctx, namespace, name, appConfig)
	}
	return foundSecret, nil
}

func addValidSecret(ctx context.Context, namespace, name string, appConfig *craneConfig.Configuration) (string, error) {
	log.Printf("storing k8s secret at %s/%s", namespace, name)

	keyStr, keyErr := config.GenerateKeyString()
	if keyErr != nil {
		return "", keyErr
	}

	secret := map[string][]byte{}
	secret[SecretFileName] = []byte(keyStr)

	storedVersion, storingErr := k8s.ApplyOpaqueSecret(ctx, namespace, name, secret, appConfig)
	if storingErr != nil {
		return "", storingErr
	}

	log.Printf("k8s secret at %s/%s is updated (resourceVersion: %s)", namespace, name, storedVersion)

	return keyStr, nil
}

func isNamespaceExisted(namespace string, appConfig *craneConfig.Configuration) error {
	var isNamespaceFound bool
	if namespaces, err := k8s.GetNamespaces(appConfig); err != nil {
		return fmt.Errorf("namespace fetching error: %w", err)
	} else {
		for _, item := range namespaces {
			if item.Name == namespace {
				isNamespaceFound = true
				break
			}
		}
	}
	if !isNamespaceFound {
		return fmt.Errorf("namespace not found on k8s: %s", namespace)
	}
	return nil
}

func InjectSecret(secret string, appConfig *craneConfig.Configuration) {
	appConfig.CommonConfiguration.SecretPrivateKey = secret
}
