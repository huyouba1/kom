package kom

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// RegisterFakeCluster 注册一个包含 Fake Client 的集群用于测试
func RegisterFakeCluster(id string, objects ...runtime.Object) *Kubectl {
	// 1. 创建 Fake Clientset
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	// 2. 创建 Fake Dynamic Client
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	// 将传入的对象转换为 Unstructured 以便 Dynamic Client 使用
	// 注意：fake.NewSimpleDynamicClient 需要 runtime.Object，如果传入的是 typed object (如 v1.Pod)，
	// 它内部会自动处理，但为了保险起见，我们可以确保 scheme 包含了这些类型。
	// 这里简化处理，直接传入 objects。
	fakeDynamicClient := fake.NewSimpleDynamicClient(scheme, objects...)

	// 3. 初始化 ClusterInst
	// 注意：我们需要创建一个 ClusterInst 并注册到全局 Clusters 中，或者直接返回一个绑定了 fake client 的 Kubectl
	// 由于 kom 的设计是通过 ID 查找集群，我们需要注册它。

	k := initKubectl(&rest.Config{}, id) // config 为 empty struct

	cluster := &ClusterInst{
		ID:            id,
		Kubectl:       k,
		Client:        fakeClient,
		DynamicClient: fakeDynamicClient,
		apiResources: []*metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod", Group: "", Version: "v1"},
		},
	}

	// 4. 注册 fake 回调
	cluster.callbacks = k.initializeCallbacks()
	registerFakeHandlers(cluster.callbacks)

	// 注册到全局 map
	Clusters().clusters.Store(id, cluster)

	return k
}

func registerFakeHandlers(c *callbacks) {
	c.Get().Register("fake:get", fakeGet)
	c.List().Register("fake:list", fakeList)
	c.Delete().Register("fake:delete", fakeDelete)
}

func fakeGet(k *Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context

	var res *unstructured.Unstructured
	var err error

	if stmt.Namespaced {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
}

func fakeList(k *Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	ctx := stmt.Context

	var list *unstructured.UnstructuredList
	var err error

	if stmt.Namespaced {
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
	} else {
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return err
	}

	// 关键：处理切片转换
	destValue := reflect.ValueOf(stmt.Dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("fakeList: dest must be a pointer to a slice")
	}

	// 获取切片元素类型
	// elemType := destValue.Elem().Type().Elem() // e.g., v1.Pod

	// 创建一个新的切片来存储结果
	// slice := reflect.MakeSlice(destValue.Elem().Type(), 0, len(list.Items))

	// 直接使用 runtime.DefaultUnstructuredConverter 无法转换 List 到 Slice，需要手动遍历
	// 参考 callbacks/list.go 的实现

	// 这里我们简化：假设 dest 是 *[]v1.Pod
	// 我们遍历 list.Items，将每个 item 转换为 v1.Pod，然后 append

	// 通用做法：
	sliceValue := destValue.Elem()
	sliceValue.SetLen(0) // clear

	for _, item := range list.Items {
		// 创建元素的新实例
		newElem := reflect.New(sliceValue.Type().Elem()).Interface()

		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, newElem)
		if err != nil {
			return fmt.Errorf("convert error: %v", err)
		}

		// append to slice
		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(newElem).Elem()))
	}

	if stmt.TotalCount != nil {
		*stmt.TotalCount = int64(len(list.Items))
	}

	return nil
}

func fakeDelete(k *Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context

	if stmt.Namespaced {
		return stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return stmt.Kubectl.DynamicClient().Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
}

func TestGetPodWithFakeClient(t *testing.T) {
	podName := "test-pod"
	ns := "default"
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}

	RegisterFakeCluster("test-cluster", pod)

	var res v1.Pod
	err := Cluster("test-cluster").Resource(&v1.Pod{}).Namespace(ns).Name(podName).Get(&res).Error
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if res.Name != podName {
		t.Errorf("Expected name %s, got %s", podName, res.Name)
	}
}

func TestListPodsWithFakeClient(t *testing.T) {
	pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}}
	pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"}}

	RegisterFakeCluster("list-cluster", pod1, pod2)

	var pods []v1.Pod
	// 必须设置 GVK，因为 Resource(&v1.Pod{}) 会解析 GVK
	err := Cluster("list-cluster").Resource(&v1.Pod{}).Namespace("default").List(&pods).Error
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("Expected 2 pods, got %d", len(pods))
	}
}

func TestDeletePodWithFakeClient(t *testing.T) {
	podName := "del-pod"
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"}}
	RegisterFakeCluster("del-cluster", pod)

	err := Cluster("del-cluster").Resource(&v1.Pod{}).Namespace("default").Name(podName).Delete().Error
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	var res v1.Pod
	err = Cluster("del-cluster").Resource(&v1.Pod{}).Namespace("default").Name(podName).Get(&res).Error
	if err == nil {
		t.Errorf("Expected error after deletion, got nil")
	}
}

func TestChainSideEffect(t *testing.T) {
	RegisterFakeCluster("side-effect")
	k := Cluster("side-effect")

	// Create a ctl/pod chain
	c1 := k.Ctl().Pod()
	c1.ContainerName("c1")

	// Branch off
	c2 := c1.ContainerName("c2")

	// Check if c1 was modified
	// Note: We need to access internal state.
	// c1 and c2 are *pod, which has unexported field kubectl *Kubectl.
	// In package kom, we can access unexported fields.

	if c1 == c2 {
		t.Errorf("c1 and c2 are the same pointer: Mutable implementation confirmed")
	} else {
		t.Logf("c1 and c2 are different pointers: Immutable implementation confirmed")
	}

	// c1.kubectl should be accessible
	if c1.kubectl.Statement.ContainerName == "c2" {
		t.Errorf("Side Effect Observed: c1 container name changed to c2")
	} else {
		t.Logf("No Side Effect Observed")
	}
}

func TestStdinBug(t *testing.T) {
	RegisterFakeCluster("stdin-cluster")
	k := Cluster("stdin-cluster")
	reader := strings.NewReader("test")

	p := k.Ctl().Pod()
	p2 := p.Stdin(reader)

	if p2.kubectl.Statement.Stdin != reader {
		t.Errorf("Confirmed Bug: Stdin did not set reader on returned pod")
	} else {
		t.Logf("Bug fixed: Stdin set reader correctly")
	}
}
