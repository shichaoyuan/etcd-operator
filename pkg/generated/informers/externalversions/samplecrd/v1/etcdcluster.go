/*
Copyright The Kubernetes Authors.

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
// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	samplecrdv1 "etcd-operator/pkg/apis/samplecrd/v1"
	versioned "etcd-operator/pkg/generated/clientset/versioned"
	internalinterfaces "etcd-operator/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "etcd-operator/pkg/generated/listers/samplecrd/v1"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EtcdClusterInformer provides access to a shared informer and lister for
// EtcdClusters.
type EtcdClusterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.EtcdClusterLister
}

type etcdClusterInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEtcdClusterInformer constructs a new informer for EtcdCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEtcdClusterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEtcdClusterInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEtcdClusterInformer constructs a new informer for EtcdCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEtcdClusterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecrdV1().EtcdClusters(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecrdV1().EtcdClusters(namespace).Watch(context.TODO(), options)
			},
		},
		&samplecrdv1.EtcdCluster{},
		resyncPeriod,
		indexers,
	)
}

func (f *etcdClusterInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEtcdClusterInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *etcdClusterInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&samplecrdv1.EtcdCluster{}, f.defaultInformer)
}

func (f *etcdClusterInformer) Lister() v1.EtcdClusterLister {
	return v1.NewEtcdClusterLister(f.Informer().GetIndexer())
}
