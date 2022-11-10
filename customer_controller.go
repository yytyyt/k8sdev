package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8sdev/myutil"
	"time"
)

// Controller Pod 控制器
type Controller struct {
	workqueue workqueue.RateLimitingInterface
	indexer   cache.Indexer
	informer  cache.Controller // 里面包含 Reflector\DeltaFIFO\ListWatcher
}

func NewController(workqueue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		workqueue: workqueue,
		indexer:   indexer,
		informer:  informer,
	}
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()
	// 停止控制器后需要关掉队列
	defer c.workqueue.ShutDown()

	// 启动控制器
	klog.Info("starting pod controller")

	// 启动通用控制器框架
	go c.informer.Run(stopCh)

	// 等待所有相关得缓存同步完成， 然后再开始处理workqueue中得数据
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for cahchs to sync"))
	}

	// 启动worker处理元素
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	<-stopCh
	klog.Info("stopping pod controller")
}

// 处理元素
func (c *Controller) runWorker() {
	for c.processNextItem() {

	}

}

// 处理元素
func (c *Controller) processNextItem() bool {
	// 从 workqueue里面取出一个元素
	key, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	// 告诉队列我们已经处理了该key
	defer c.workqueue.Done(key)

	// 根据key 去处理我们得业务逻辑
	err := c.syncToStdout(key.(string))
	// 错误处理
	c.handleErr(err, key)
	return true
}

// 业务逻辑处理
func (c *Controller) syncToStdout(key string) error {
	// 从indexer获取元素数据
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetch object with key %s from indexer failed with %v", key, err)
	}
	if !exists {
		fmt.Printf("Pod %s does not exists anymore \n", key)
	} else {
		fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*v1.Pod).GetName())
	}
	return nil
}

// 错误处理
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.workqueue.Forget(key)
		return
	}
	// 如果出现了问题 我们运行当前控制器重试5次
	if c.workqueue.NumRequeues(key) < 5 {
		// 重新入队列
		c.workqueue.AddRateLimited(key)
		return
	}
	c.workqueue.Forget(key)
	runtime.HandleError(err)
}

func main() {
	clientSet, err := myutil.GetClientSet()
	if err != nil {
		klog.Fatal(err)
	}
	// 创建Pod得 ListWatch
	podListWatcher := cache.NewListWatchFromClient(clientSet.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	// 创建队列
	workqueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				workqueue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				workqueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				workqueue.Add(key)
			}
		},
	}, cache.Indexers{})

	// 实例化Pod控制器
	controller := NewController(workqueue, indexer, informer)

	stopCh := make(chan struct{})
	defer close(stopCh)
	go controller.Run(1, stopCh)
	select {}

}
