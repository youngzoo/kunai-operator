package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	generationKey = "nullable.se/configmap-generation"
	resyncPeriod  = 5 * time.Minute
)

var namespace = "kunai"
var configmapName = "kunai-config"
var daemonsetName = "kunai"

func init() {
	if os.Getenv("KUBERNETES_NAMESPACE") != "" {
		namespace = os.Getenv("KUBERNETES_NAMESPACE")
	}
	if os.Getenv("RELEASE_NAME") != "" {
		configmapName = os.Getenv("RELEASE_NAME") + "-config"
		daemonsetName = os.Getenv("RELEASE_NAME")
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create clientset: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go handleSignals(ctx, cancel, &wg)
	go runConfigMapWatcher(ctx, clientset, &wg)

	wg.Wait()
	klog.Infof("Controller stopped.")
}

func handleSignals(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	klog.Infof("Received signal %s, shutting down.", sig)
	cancel()
}

func runConfigMapWatcher(ctx context.Context, clientset *kubernetes.Clientset, wg *sync.WaitGroup) {
	defer wg.Done()

	listWatch := cache.NewFilteredListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"configmaps",
		namespace,
		func(options *metav1.ListOptions) {
			options.FieldSelector = fmt.Sprintf("metadata.name=%s", configmapName)
		},
	)

	informer := cache.NewSharedInformer(listWatch, &corev1.ConfigMap{}, resyncPeriod)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			newCm := newObj.(*corev1.ConfigMap)
			oldCm := oldObj.(*corev1.ConfigMap)
			if newCm.ResourceVersion != oldCm.ResourceVersion {
				klog.Infof("ConfigMap %s updated, restarting DaemonSet %s", configmapName, daemonsetName)
				if err := restartDaemonSet(ctx, clientset); err != nil {
					klog.Errorf("Failed to restart DaemonSet: %v", err)
				}
			}
		},
	})

	informer.Run(ctx.Done())
	klog.Infof("ConfigMap watcher stopped.")
}

func restartDaemonSet(ctx context.Context, clientset *kubernetes.Clientset) error {
	daemonsets := clientset.AppsV1().DaemonSets(namespace)

	ds, err := daemonsets.Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get DaemonSet: %w", err)
	}

	annotations := ds.Spec.Template.ObjectMeta.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	generationStr := annotations[generationKey]
	generation := 0

	if generationStr != "" {
		_, err := fmt.Sscanf(generationStr, "%d", &generation)
		if err != nil {
			klog.V(2).Infof("Error parsing generation: %v, defaulting to 0", err)
			generation = 0
		}
	}

	annotations[generationKey] = fmt.Sprintf("%d", generation+1)
	ds.Spec.Template.ObjectMeta.Annotations = annotations

	_, err = daemonsets.Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update DaemonSet: %w", err)
	}

	klog.Infof("DaemonSet %s restarted.", daemonsetName)
	return nil
}
