package controller

import (
	samplecrdv1 "etcd-operator/pkg/apis/samplecrd/v1"
	clientset "etcd-operator/pkg/generated/clientset/versioned"
	etcdclusterscheme "etcd-operator/pkg/generated/clientset/versioned/scheme"
	informers "etcd-operator/pkg/generated/informers/externalversions/samplecrd/v1"
	listers "etcd-operator/pkg/generated/listers/samplecrd/v1"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/golang/glog"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "etcdcluster-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a EtcdCluster is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a EtcdCluster
	// is synced successfully
	MessageResourceSynced = "EtcdCluster synced successfully"
)

// Controller is the controller implementation for EtcdCluster resources
type Controller struct {
	kubeclientset kubernetes.Interface
	clientset     clientset.Interface

	lister listers.EtcdClusterLister
	synced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

// NewController returns a new etcdcluster controller
func NewController(kubeclientset kubernetes.Interface, clientset clientset.Interface, informer informers.EtcdClusterInformer) *Controller {

	utilruntime.Must(etcdclusterscheme.AddToScheme(scheme.Scheme))
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		lister:        informer.Lister(),
		synced:        informer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Etcdclusters"),
		recorder:      recorder,
	}

	glog.Info("Setting up event handlers")
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueEtcdCluster,
		UpdateFunc: func(old, new interface{}) {
			oldEtcdCluster := old.(*samplecrdv1.EtcdCluster)
			newEtcdCluster := new.(*samplecrdv1.EtcdCluster)
			if oldEtcdCluster.ResourceVersion == newEtcdCluster.ResourceVersion {
				return
			}
			controller.enqueueEtcdCluster(new)
		},
		DeleteFunc: controller.enqueueEtcdClusterForDelete,
	})

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("Starting EtcdCluster control loop")

	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	go wait.Until(c.runWorker, time.Second, stopCh)

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

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
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// EtcdCluster resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Network resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the EtcdCluster resource with this namespace/name
	etcdcluster, err := c.lister.EtcdClusters(namespace).Get(name)
	if err != nil {
		// The EtcdCluster resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			glog.Warningf("EtcdCluster: %s/%s does not exist in local cache, will delete it from Neutron ...",
				namespace, name)

			glog.Infof("[Neutron] Deleting network: %s/%s ...", namespace, name)

			// FIX ME: call Neutron API to delete this network by name.
			//
			// neutron.Delete(namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list etcdcluster by: %s/%s", namespace, name))

		return err
	}

	glog.Infof("[Neutron] Try to process etcdcluster: %#v ...", etcdcluster)

	// FIX ME: Do diff().
	//
	// actualNetwork, exists := neutron.Get(namespace, name)
	//
	// if !exists {
	// 	neutron.Create(namespace, name)
	// } else if !reflect.DeepEqual(actualNetwork, network) {
	// 	neutron.Update(namespace, name)
	// }

	c.recorder.Event(etcdcluster, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) enqueueEtcdCluster(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) enqueueEtcdClusterForDelete(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}
