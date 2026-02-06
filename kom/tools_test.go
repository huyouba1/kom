package kom

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func TestTools(t *testing.T) {
	// Setup
	cfg := &rest.Config{Host: "https://test-tools"}
	id := "test-tools"
	k, err := Clusters().RegisterByConfigWithID(cfg, id, RegisterDisableCRDWatch())
	assert.NoError(t, err)

	// Mock APIResources
	cluster := Clusters().GetClusterById(id)
	cluster.apiResources = []*metav1.APIResource{
		{
			Name:         "pods",
			SingularName: "pod",
			Namespaced:   true,
			Kind:         "Pod",
			Group:        "",
			Version:      "v1",
			ShortNames:   []string{"po"},
		},
		{
			Name:         "deployments",
			SingularName: "deployment",
			Namespaced:   true,
			Kind:         "Deployment",
			Group:        "apps",
			Version:      "v1",
			ShortNames:   []string{"deploy"},
		},
	}

	// Mock CRDList
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "crontabs.stable.example.com",
			},
			"spec": map[string]interface{}{
				"group": "stable.example.com",
				"versions": []interface{}{
					map[string]interface{}{
						"name":    "v1",
						"served":  true,
						"storage": true,
					},
				},
				"scope": "Namespaced",
				"names": map[string]interface{}{
					"plural":     "crontabs",
					"singular":   "crontab",
					"kind":       "CronTab",
					"shortNames": []interface{}{"ct"},
				},
			},
		},
	}
	cluster.crdList = []*unstructured.Unstructured{crd}
	// Inject CRDList into cache so Status().CRDList() returns it
	k.ClusterCache().Set("crdList", []*unstructured.Unstructured{crd}, 1)
	k.ClusterCache().Wait()
	time.Sleep(10 * time.Millisecond)

	tools := k.Tools()

	t.Run("DebugCRD", func(t *testing.T) {
		fmt.Printf("k.ID: %s\n", k.ID)
		clusterFromMgr := Clusters().GetClusterById(id)
		fmt.Printf("Cluster from Mgr: %p, CRDList len: %d\n", clusterFromMgr, len(clusterFromMgr.crdList))

		parentCluster := k.ParentCluster()
		fmt.Printf("Parent Cluster: %p, CRDList len: %d\n", parentCluster, len(parentCluster.crdList))

		crdList := k.Status().CRDList()
		fmt.Printf("CRDList from Status: len: %d\n", len(crdList))
		if len(crdList) > 0 {
			c := crdList[0]
			spec, found, err := unstructured.NestedMap(c.Object, "spec")
			fmt.Printf("Spec found: %v, err: %v\n", found, err)
			if found {
				kind, found, err := unstructured.NestedString(spec, "names", "kind")
				fmt.Printf("Kind: %s, found: %v, err: %v\n", kind, found, err)
			}
		}
	})

	t.Run("IsBuiltinResource", func(t *testing.T) {
		assert.True(t, tools.IsBuiltinResource("Pod"))
		assert.True(t, tools.IsBuiltinResource("Deployment"))
		assert.False(t, tools.IsBuiltinResource("CronTab"))
	})

	t.Run("GetGVRByKind", func(t *testing.T) {
		gvr, namespaced := tools.GetGVRByKind("Pod")
		assert.True(t, namespaced)
		assert.Equal(t, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, gvr)

		gvr, namespaced = tools.GetGVRByKind("Deployment")
		assert.True(t, namespaced)
		assert.Equal(t, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, gvr)

		gvr, _ = tools.GetGVRByKind("Unknown")
		assert.Equal(t, schema.GroupVersionResource{}, gvr)
	})

	t.Run("GetGVRByGVK", func(t *testing.T) {
		gvk := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}
		gvr, namespaced, ok := tools.GetGVRByGVK(gvk)
		assert.True(t, ok)
		assert.True(t, namespaced)
		assert.Equal(t, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, gvr)
	})

	t.Run("GetCRD", func(t *testing.T) {
		foundCRD, err := tools.GetCRD("CronTab", "stable.example.com")
		assert.NoError(t, err)
		assert.NotNil(t, foundCRD)

		_, err = tools.GetCRD("Unknown", "unknown")
		assert.Error(t, err)
	})

	t.Run("GetGVRFromCRD", func(t *testing.T) {
		gvr := tools.GetGVRFromCRD(crd)
		assert.Equal(t, schema.GroupVersionResource{Group: "stable.example.com", Version: "v1", Resource: "crontabs"}, gvr)
	})

	t.Run("FindGVKByTableNameInApiResources", func(t *testing.T) {
		gvk := tools.FindGVKByTableNameInApiResources("pods")
		assert.NotNil(t, gvk)
		assert.Equal(t, schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, *gvk)

		gvk = tools.FindGVKByTableNameInApiResources("po")
		assert.NotNil(t, gvk)
		assert.Equal(t, schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, *gvk)
	})

	t.Run("FindGVKByTableNameInCRDList", func(t *testing.T) {
		gvk := tools.FindGVKByTableNameInCRDList("crontabs")
		if assert.NotNil(t, gvk) {
			assert.Equal(t, schema.GroupVersionKind{Group: "stable.example.com", Version: "v1", Kind: "CronTab"}, *gvk)
		}

		gvk = tools.FindGVKByTableNameInCRDList("ct")
		if assert.NotNil(t, gvk) {
			assert.Equal(t, schema.GroupVersionKind{Group: "stable.example.com", Version: "v1", Kind: "CronTab"}, *gvk)
		}
	})

	t.Run("ListAvailableTableNames", func(t *testing.T) {
		names := tools.ListAvailableTableNames()
		assert.Contains(t, names, "pod")
		assert.Contains(t, names, "po")
		assert.Contains(t, names, "deployment")
		assert.Contains(t, names, "deploy")
	})

	t.Run("ConvertRuntimeObjectToTypedObject", func(t *testing.T) {
		pod := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			},
		}
		var result map[string]interface{}
		err := tools.ConvertRuntimeObjectToTypedObject(pod, &result)
		assert.NoError(t, err)
		assert.Equal(t, "test-pod", result["metadata"].(map[string]interface{})["name"])
	})

	t.Run("ParseGVK2GVR", func(t *testing.T) {
		// Test Builtin
		gvks := []schema.GroupVersionKind{{Group: "", Version: "v1", Kind: "Pod"}}
		gvr, namespaced := tools.ParseGVK2GVR(gvks)
		assert.True(t, namespaced)
		assert.Equal(t, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, gvr)

		// Test CRD
		crdGVKs := []schema.GroupVersionKind{{Group: "stable.example.com", Version: "v1", Kind: "CronTab"}}
		gvr, namespaced = tools.ParseGVK2GVR(crdGVKs)
		assert.True(t, namespaced)
		assert.Equal(t, schema.GroupVersionResource{Group: "stable.example.com", Version: "v1", Resource: "crontabs"}, gvr)
	})

	t.Run("ClearCache", func(t *testing.T) {
		assert.NotNil(t, k.ClusterCache())
		// Add something to cache
		k.ClusterCache().Set("key", "value", 1)
		k.ClusterCache().Wait()
		time.Sleep(10 * time.Millisecond) // Give it a bit more time for async ops

		val, found := k.ClusterCache().Get("key")
		assert.True(t, found)
		assert.Equal(t, "value", val)

		tools.ClearCache()
		k.ClusterCache().Wait()
		time.Sleep(10 * time.Millisecond)

		_, found = k.ClusterCache().Get("key")
		assert.False(t, found)
	})
}
