package k8s

import (
	"context"
	"log"

	"github.com/dyrector-io/dyrectorio/agent/internal/crypt"
	"github.com/dyrector-io/dyrectorio/agent/pkg/crane/config"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// namespace wrapper for the facade
type secret struct {
	ctx context.Context

	avail     []string
	appConfig *config.Configuration
}

func newSecret(ctx context.Context, cfg *config.Configuration) *secret {
	secr := secret{ctx: ctx, appConfig: cfg, avail: []string{}}
	return &secr
}

func (s *secret) applySecrets(namespace, name string, values map[string]string) error {
	cli, err := getSecretClient(namespace, s.appConfig)
	if err != nil {
		return err
	}

	decrypted, err := crypt.DecryptSecrets(values, &s.appConfig.CommonConfiguration)

	if err != nil {
		return err
	}

	secrets := corev1.Secret(name, namespace).WithData(decrypted)

	result, err := cli.Apply(s.ctx, secrets, metav1.ApplyOptions{
		FieldManager: s.appConfig.FieldManagerName,
		Force:        s.appConfig.ForceOnConflicts,
	})

	if result != nil {
		log.Println("Secret updated: " + result.Name)
		if len(values) > 0 {
			s.avail = append(s.avail, name)
		}
	}

	return err
}

func ListSecrets(ctx context.Context, namespace, name string, appConfig *config.Configuration) ([]string, error) {
	cli, err := getSecretClient(namespace, appConfig)
	if err != nil {
		return nil, err
	}

	list, err := cli.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	names := []string{}

	for index := range list.Items {
		secret := &list.Items[index]
		if secret.Name == name {
			for key := range secret.Data {
				names = append(names, key)
			}
			break
		}
	}

	return names, nil
}

func getSecretClient(namespace string, cfg *config.Configuration) (v1.SecretInterface, error) {
	clientset, err := GetClientSet(cfg)

	if err != nil {
		return nil, err
	}

	return clientset.CoreV1().Secrets(namespace), nil
}

func ApplyOpaqueSecret(ctx context.Context, namespace, name string, values map[string][]byte, appConfig *config.Configuration) (
	version string, err error,
) {
	cli, err := getSecretClient(namespace, appConfig)
	if err != nil {
		return "", err
	}

	secrets := corev1.Secret(name, namespace).WithData(values)

	result, err := cli.Apply(ctx, secrets, metav1.ApplyOptions{
		FieldManager: appConfig.FieldManagerName,
		Force:        appConfig.ForceOnConflicts,
	})

	if err != nil {
		return "", err
	}

	return result.ResourceVersion, nil
}

func GetSecret(ctx context.Context, namespace, name string, appConfig *config.Configuration) (
	secretFiles map[string][]byte, version string, err error,
) {
	cli, err := getSecretClient(namespace, appConfig)
	if err != nil {
		return nil, "", err
	}

	item, err := cli.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}
	return item.Data, item.ResourceVersion, nil
}
