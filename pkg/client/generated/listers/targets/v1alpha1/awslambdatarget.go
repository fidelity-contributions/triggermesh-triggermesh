/*
Copyright 2021 TriggerMesh Inc.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/triggermesh/triggermesh/pkg/apis/targets/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AWSLambdaTargetLister helps list AWSLambdaTargets.
// All objects returned here must be treated as read-only.
type AWSLambdaTargetLister interface {
	// List lists all AWSLambdaTargets in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AWSLambdaTarget, err error)
	// AWSLambdaTargets returns an object that can list and get AWSLambdaTargets.
	AWSLambdaTargets(namespace string) AWSLambdaTargetNamespaceLister
	AWSLambdaTargetListerExpansion
}

// aWSLambdaTargetLister implements the AWSLambdaTargetLister interface.
type aWSLambdaTargetLister struct {
	indexer cache.Indexer
}

// NewAWSLambdaTargetLister returns a new AWSLambdaTargetLister.
func NewAWSLambdaTargetLister(indexer cache.Indexer) AWSLambdaTargetLister {
	return &aWSLambdaTargetLister{indexer: indexer}
}

// List lists all AWSLambdaTargets in the indexer.
func (s *aWSLambdaTargetLister) List(selector labels.Selector) (ret []*v1alpha1.AWSLambdaTarget, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AWSLambdaTarget))
	})
	return ret, err
}

// AWSLambdaTargets returns an object that can list and get AWSLambdaTargets.
func (s *aWSLambdaTargetLister) AWSLambdaTargets(namespace string) AWSLambdaTargetNamespaceLister {
	return aWSLambdaTargetNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AWSLambdaTargetNamespaceLister helps list and get AWSLambdaTargets.
// All objects returned here must be treated as read-only.
type AWSLambdaTargetNamespaceLister interface {
	// List lists all AWSLambdaTargets in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AWSLambdaTarget, err error)
	// Get retrieves the AWSLambdaTarget from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.AWSLambdaTarget, error)
	AWSLambdaTargetNamespaceListerExpansion
}

// aWSLambdaTargetNamespaceLister implements the AWSLambdaTargetNamespaceLister
// interface.
type aWSLambdaTargetNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AWSLambdaTargets in the indexer for a given namespace.
func (s aWSLambdaTargetNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.AWSLambdaTarget, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AWSLambdaTarget))
	})
	return ret, err
}

// Get retrieves the AWSLambdaTarget from the indexer for a given namespace and name.
func (s aWSLambdaTargetNamespaceLister) Get(name string) (*v1alpha1.AWSLambdaTarget, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("awslambdatarget"), name)
	}
	return obj.(*v1alpha1.AWSLambdaTarget), nil
}