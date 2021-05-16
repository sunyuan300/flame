package k8s

import (
	"context"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ConfigMapUpdate(clientSet *kubernetes.Clientset, name string, data map[string]string) error {
	newConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Data:       data,
	}
	_, err := clientSet.CoreV1().ConfigMaps(viper.GetString("namespace")).Update(context.TODO(), newConfigMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
