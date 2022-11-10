package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sdev/myutil"
)

func main() {
	clientSet, err := myutil.GetClientSet()
	if err != nil {
		panic(err.Error())
	}
	// 使用clientset获取资源对象 进行CRUD
	deployments, err := clientSet.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, item := range deployments.Items {
		fmt.Printf("%s\t%s\t%s\n", item.Namespace, item.Name, item.APIVersion)
	}

}
