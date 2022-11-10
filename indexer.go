package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// 索引名 或者 索引分类
const NamespaceIndexName = "namespace"
const NodeNameIndexName = "nodeName"

func NamespaceIndexFunc(obj interface{}) ([]string, error) {
	metaObj, _ := meta.Accessor(obj)
	return []string{metaObj.GetNamespace()}, nil
}
func NodeNameIndexFunc(obj interface{}) ([]string, error) {
	pod, _ := obj.(*v1.Pod)
	return []string{pod.Spec.NodeName}, nil
}

func main() {
	/*
		Indexers: {
			"namespace": NamespaceIndexFunc,
			"nodeName": NodeNameIndexFunc,
		}

		Indices: {
			"namespace":{
				"default":["pod-1","pod-3"]  // Index (map) "pod-1" -> PodObject
				"kube-system":["pod-2"]  // Index
			},
			"nodeName":{
				"centos-01":["pod-1","pod-3"] // Index
				"centos-02":["pod-3"] // Index
			}
		}
	**/

	// 初始化indexer
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		NamespaceIndexName: NamespaceIndexFunc,
		NodeNameIndexName:  NodeNameIndexFunc,
	})
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "centos-01",
		},
	}

	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "kube-system",
		},
		Spec: v1.PodSpec{
			NodeName: "centos-01",
		},
	}

	pod3 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-3",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "centos-02",
		},
	}
	// pod加入到indexer中去
	_ = indexer.Add(pod1)
	_ = indexer.Add(pod2)
	_ = indexer.Add(pod3)

	pods, _ := indexer.ByIndex(NamespaceIndexName, "default")
	for _, pod := range pods {
		fmt.Println(pod.(*v1.Pod).Name)
	}
	fmt.Println("********************")
	pods, _ = indexer.ByIndex(NodeNameIndexName, "centos-01")
	for _, pod := range pods {
		fmt.Println(pod.(*v1.Pod).Name)
	}
}
