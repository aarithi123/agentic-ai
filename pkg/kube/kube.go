package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctlr "sigs.k8s.io/controller-runtime"

	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	config        *rest.Config
	ingressGVR    = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
)

func GetServices(namespace string) ([]string, error) {
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	list, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	var services []string
	for _, svc := range list.Items {
		services = append(services, svc.Name)
	}
	return services, nil
}

func GetDeployments(namespace string) ([]string, error) {
	// Ensure the namespace exists
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("namespace '%s' not found: %v", namespace, err)
	}

	// List deployments
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	var names []string
	for _, d := range deployments.Items {
		names = append(names, d.Name)
	}
	return names, nil
}

func GetIngresses(namespace string) ([]string, error) {
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	list, err := dynamicClient.Resource(ingressGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %v", err)
	}

	var ingresses []string
	for _, ing := range list.Items {
		ingresses = append(ingresses, ing.GetName())
	}
	return ingresses, nil
}

func GetNamespaceLabels(namespace string) (map[string]string, error) {
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ns.Labels, nil
}

func GetNamespaces() ([]string, error) {
	list, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var namespaces []string
	for _, ns := range list.Items {
		namespaces = append(namespaces, ns.Name)
	}
	return namespaces, nil
}

type RestartOutput struct {
	Message string `json:"message"`
	Pods    []Pod  `json:"pods"`
}

type Pod struct {
	Name string `json:"name"`
	Node string `json:"node"`
}

func RestartApplication(namespace, app string) (string, error) {
	namespace = strings.ToLower(namespace)
	deployment := strings.ToLower(app)

	list, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "{}", fmt.Errorf("failed to list pods: %v", err)
	}

	var pods []Pod
	prefix := deployment + "-"
	for _, pod := range list.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			pods = append(pods, Pod{Name: pod.Name, Node: pod.Spec.NodeName})
		}
	}

	now := time.Now().Format(time.RFC3339)
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, now)

	_, err = clientset.AppsV1().Deployments(namespace).Patch(context.TODO(), deployment, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return "{}", fmt.Errorf("failed to patch deployment: %v", err)
	}

	output := RestartOutput{Message: "Restart initiated", Pods: pods}
	jsonDoc, err := json.Marshal(output)
	if err != nil {
		return "{}", fmt.Errorf("failed to marshal output: %v", err)
	}
	return string(jsonDoc), nil
}

type PodCpuMemory struct {
	Name   string `json:"name"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

func GetPodCpuMemory(namespace string) ([]PodCpuMemory, error) {
	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %v", err)
	}

	var results []PodCpuMemory
	for _, item := range podMetricsList.Items {
		var totalCPU, totalMem string
		for _, container := range item.Containers {
			totalCPU = container.Usage.Cpu().String()
			totalMem = container.Usage.Memory().String()
		}
		results = append(results, PodCpuMemory{
			Name:   item.Name,
			CPU:    totalCPU,
			Memory: totalMem,
		})
	}

	return results, nil
}

func init() {
	config = ctlr.GetConfigOrDie()
	clientset = kubernetes.NewForConfigOrDie(config)
	dynamicClient = dynamic.NewForConfigOrDie(config)
}
