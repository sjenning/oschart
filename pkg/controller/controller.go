package controller

import (
	"encoding/json"
	"fmt"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	clientset "github.com/openshift/client-go/config/clientset/versioned"
	configinformersv1 "github.com/openshift/client-go/config/informers/externalversions/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/sjenning/oschart/pkg/event"

	"k8s.io/klog/v2"
)

type Controller struct {
	client     clientset.Interface
	coLister   configlistersv1.ClusterOperatorLister
	coSynced   cache.InformerSynced
	workqueue  workqueue.RateLimitingInterface
	eventStore event.Store
}

func New(c clientset.Interface, coInformer configinformersv1.ClusterOperatorInformer, eventStore event.Store) *Controller {
	controller := &Controller{
		client:     c,
		coLister:   coInformer.Lister(),
		coSynced:   coInformer.Informer().HasSynced,
		workqueue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ClusterOperators"),
		eventStore: eventStore,
	}

	coInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueClusterOperator,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueClusterOperator(new)
		},
		DeleteFunc: controller.enqueueClusterOperator,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting ClusterOperator controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.coSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process ClusterOperator resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}

	co, err := c.coLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// co is deleted
			klog.Infof("co %s is deleted", name)
			return nil
		}
	}

	// co is added or updated
	data, _ := json.MarshalIndent(co, "", "  ")
	for _, condition := range co.Status.Conditions {
		if condition.Status == configv1.ConditionTrue {
			c.eventStore.Add(co.GetName(), string(condition.Type), string(condition.Type), string(data))
		} else {
			c.eventStore.Add(co.GetName(), string(condition.Type), string(condition.Status), string(data))
		}
	}

	return nil
}

func (c *Controller) enqueueClusterOperator(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}
