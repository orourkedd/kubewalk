package main

import (
	"fmt"
	"os"
	"path/filepath"

	v1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	envs := []string{"test", "staging", "demo", "integration", "infraprelive"}
	// envs := []string{"live", "infralive"}
	configPath := filepath.Join(homeDir(), ".kube", "config")

	// Out-of-cluster config
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		panic(err)
	}

	// Get ingresses for all environments
	ingresses := []v1beta1.Ingress{}
	for _, env := range envs {
		list, err := clientset.ExtensionsV1beta1().Ingresses(env).List(v1.ListOptions{})
		if err != nil {
			panic(err)
		}
		ingresses = append(ingresses, list.Items...)
	}

	// Update ingress annotation
	for _, ing := range ingresses {
		annotation, ok := ing.Annotations["kubernetes.io/ingress.class"]
		if !ok {
			continue
		}

		if annotation == "nginx" {
			continue
		}

		ing.Annotations["kubernetes.io/ingress.class"] = "nginx"

		_, err := clientset.ExtensionsV1beta1().Ingresses(ing.Namespace).Update(&ing)
		if err != nil {
			panic(err)
		}

		fmt.Println(fmt.Sprintf("[%s] %s - %s", ing.Namespace, ing.Name, annotation))
	}
}

func homeDir() string {
	return os.Getenv("HOME")
}
