package watcher

import (
	"flame/pkg/factory"
	"fmt"
	"github.com/prometheus/prometheus/pkg/rulefmt"
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
type RulesController struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller

	Instance factory.RulesConfigInstance
}

// NewController creates a new Controller.
func NewRulesController(clientSet *kubernetes.Clientset) *RulesController {
	// create the pod watcher
	ruleListWatcher := cache.NewListWatchFromClient(clientSet.CoreV1().RESTClient(), "configmaps", viper.GetString("namespace"), fields.OneTermEqualSelector("metadata.name", viper.GetString("rules-configmap")))

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

	//var ruleInstance factory.RulesConfigInstance
	//ruleInstance.AllRulesGroups = make(map[string]*rulefmt.RuleGroups)
	return &RulesController{
		indexer:  indexer,
		queue:    queue,
		informer: informer,
		//Instance: ruleInstance,
	}
}

func (r *RulesController) RunRulesController() {
	stop := make(chan struct{})
	defer close(stop)
	r.Run(1, stop)
}

func (r *RulesController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := r.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer r.queue.Done(key)

	// Invoke the method containing the business logic
	err := r.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	r.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (r *RulesController) syncToStdout(key string) error {
	obj, exists, err := r.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("rules configmap %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		fmt.Printf("Sync/Add/Update for rules configmap %s\n", obj.(*v1.ConfigMap).GetName())
		info := obj.(*v1.ConfigMap)
		r.Instance.AllRulesGroups = make(map[string]*rulefmt.RuleGroups)
		for k, v := range info.Data {
			groups, errs := rulefmt.Parse([]byte(v))
			if errs != nil {
				for _, err := range errs {
					klog.Error(err)
				}
				return nil
			}
			r.Instance.AllRulesGroups[k] = groups
		}
		r.Instance.Lock.Unlock()
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (r *RulesController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		r.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if r.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing rules configmap %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		r.queue.AddRateLimited(key)
		return
	}

	r.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping rules configmap %q out of the queue: %v", key, err)
}

// Run begins watching and syncing.
func (r *RulesController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer r.queue.ShutDown()
	klog.Info("Starting rules configmap controller")

	go r.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, r.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(r.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping rules configmap controller")
}

func (r *RulesController) runWorker() {
	for r.processNextItem() {
	}
}
