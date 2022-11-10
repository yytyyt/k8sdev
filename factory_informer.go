package main

import (
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8sdev/myutil"
	"time"
)

func main() {
	clientSet, err := myutil.GetClientSet()
	if err != nil {
		panic(err.Error())
	}
	// 初始化 informer factory
	informerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)
	// 监听想要获取得资源对象informer
	deployInformer := informerFactory.Apps().V1().Deployments()
	// 相当于注册下 informer
	informer := deployInformer.Informer()
	// 创建Lister
	lister := deployInformer.Lister()

	// 注册事件处理程序
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*v1.Deployment)
			fmt.Println("add a deployment:", deployment.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeployment := oldObj.(*v1.Deployment)
			newDeployment := newObj.(*v1.Deployment)
			fmt.Println("update deployment:", oldDeployment.Name, newDeployment.Name)
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*v1.Deployment)
			fmt.Println("update a deployment:", deployment.Name)
		},
	})
	// 启动informer (List & Watch)
	stopCh := make(chan struct{})
	defer close(stopCh)
	informerFactory.Start(stopCh)

	// 等待所有启动得informer同步完成
	informerFactory.WaitForCacheSync(stopCh)

	// 通过Lister获取缓存中得Deployment得数据
	deployments, err := lister.Deployments("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for index, deployment := range deployments {
		fmt.Printf("%d->%s\n", index, deployment.Name)
	}
	<-stopCh
}
