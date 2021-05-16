package watcher

import (
	"flame/pkg/factory"
	"fmt"
	"github.com/prometheus/prometheus/config"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// PromController demonstrates how to implement a controller with client-go.
type PromController struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller

	Instance factory.PromConfigInstance
}

// NewController creates a new Controller.
func NewPromController(clientSet *kubernetes.Clientset) *PromController {
	// create the pod watcher
	ruleListWatcher := cache.NewListWatchFromClient(clientSet.CoreV1().RESTClient(), "configmaps", viper.GetString("namespace"), fields.OneTermEqualSelector("metadata.name", viper.GetString("prometheus-configmap")))

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(ruleListWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	return &PromController{
		indexer:  indexer,
		queue:    queue,
		informer: informer,
	}
}

func (p *PromController) RunPromController() {
	stop := make(chan struct{})
	defer close(stop)
	p.Run(1, stop)

	//// Wait forever
	//select {}
}

//func NewAndRunPromController(clientSet *kubernetes.Clientset) {
//	// create the pod watcher
//	ruleListWatcher := cache.NewListWatchFromClient(clientSet.CoreV1().RESTClient(), "configmaps", viper.GetString("namespace"), fields.OneTermEqualSelector("metadata.name", viper.GetString("prometheus-configmap")))
//
//	// create the workqueue
//	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
//
//	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
//	// whenever the cache is updated, the pod key is added to the workqueue.
//	// Note that when we finally process the item from the workqueue, we might see a newer version
//	// of the Pod than the version which was responsible for triggering the update.
//	indexer, informer := cache.NewIndexerInformer(ruleListWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
//		AddFunc: func(obj interface{}) {
//			key, err := cache.MetaNamespaceKeyFunc(obj)
//			if err == nil {
//				queue.Add(key)
//			}
//		},
//		UpdateFunc: func(old interface{}, new interface{}) {
//			key, err := cache.MetaNamespaceKeyFunc(new)
//			if err == nil {
//				queue.Add(key)
//			}
//		},
//		DeleteFunc: func(obj interface{}) {
//			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
//			// key function.
//			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
//			if err == nil {
//				queue.Add(key)
//			}
//		},
//	}, cache.Indexers{})
//
//	controller := PromController{
//		indexer:  indexer,
//		queue:    queue,
//		informer: informer,
//	}
//	// Now let's start the controller
//	stop := make(chan struct{})
//	defer close(stop)
//	go controller.Run(1, stop)
//
//	// Wait forever
//	select {}
//}

func (p *PromController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := p.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer p.queue.Done(key)

	// Invoke the method containing the business logic
	err := p.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	p.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (p *PromController) syncToStdout(key string) error {
	obj, exists, err := p.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Pod %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*v1.ConfigMap).GetName())
		info := obj.(*v1.ConfigMap)
		res, err := config.Load(info.Data[viper.GetString("prometheus.yml")])
		if err != nil {
			return err
		}
		p.Instance.Config = res
		p.Instance.UpdateScrapeCache()
		//p.Instance.Lock.Unlock()
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (p *PromController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		p.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if p.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		p.queue.AddRateLimited(key)
		return
	}

	p.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

// Run begins watching and syncing.
func (p *PromController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer p.queue.ShutDown()
	klog.Info("Starting Pod controller")

	go p.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, p.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(p.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping Pod controller")
}

func (p *PromController) runWorker() {
	for p.processNextItem() {
	}
}
