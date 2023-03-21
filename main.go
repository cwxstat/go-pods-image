package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func PodLogs(clientset *kubernetes.Clientset, pod *corev1.Pod) error {

	if pod.Status.Phase != corev1.PodRunning {
		log.Printf("Pod not running")
		return nil

	}

	for _, container := range pod.Spec.Containers {
		podLogs := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			TypeMeta:  metav1.TypeMeta{},
			Container: container.Name,
		})
		logs, err := podLogs.Stream(context.Background())
		if err != nil {
			log.Printf("Error getting logs for pod %s: %v", pod.Name, err)
			return err
		}
		defer logs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, logs)
		if err != nil {
			log.Printf("Error reading logs for pod %s: %v", pod.Name, err)
			return err
		}
		fmt.Printf("  Logs: %s\n", buf.String())
	}
	return nil
}

func main() {
	// Load the Kubernetes configuration from the default location or a specified path.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{})

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Fatalf("Error loading Kubernetes configuration: %v", err)
	}

	// Create a Kubernetes client using the configuration.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// List all the pods in the "dev" namespace.
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	podImageMap := make(map[string]string)
	// Iterate over the pods
	for _, pod := range pods.Items {
		fmt.Printf("%s: Pod %s: status: %v,\n", pod.Namespace, pod.Name, pod.Status.Phase)
		for _, container := range pod.Spec.Containers {
			key := fmt.Sprintf("%s, %s, %s", pod.Namespace, pod.Name, container.Image)
			podImageMap[key] = ""
		}

	}

	var keys []string
	for k := range podImageMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Println(k)
	}
}
