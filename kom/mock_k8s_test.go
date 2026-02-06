package kom

import (
	"fmt"
	"reflect"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// RegisterFakeCluster 注册一个包含 Fake Client 的集群用于测试
func RegisterFakeCluster(id string, objects ...runtime.Object) *Kubectl {
	// 1. 创建 Fake Clientset
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	// Populate fake discovery
	if fd, ok := fakeClient.Discovery().(*discoveryfake.FakeDiscovery); ok {
		fd.FakedServerVersion = &version.Info{Major: "1", Minor: "29", GitVersion: "v1.29.0"}
		fd.Resources = []*metav1.APIResourceList{
			{
				GroupVersion: "v1",
				APIResources: []metav1.APIResource{
					{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"list", "get", "create", "delete", "update", "patch"}},
					{Name: "services", Namespaced: true, Kind: "Service", Verbs: []string{"list", "get", "create", "delete", "update", "patch"}},
				},
			},
			{
				GroupVersion: "apps/v1",
				APIResources: []metav1.APIResource{
					{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: []string{"list", "get", "create", "delete", "update", "patch"}},
				},
			},
		}
	}

	// 2. 创建 Fake Dynamic Client
	s := scheme.Scheme
	_ = apiextensionsv1.AddToScheme(s)
	// scheme := runtime.NewScheme()
	// _ = v1.AddToScheme(scheme)
	// _ = appsv1.AddToScheme(scheme)
	// metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	// 将传入的对象转换为 Unstructured 以便 Dynamic Client 使用
	// 注意：fake.NewSimpleDynamicClient 需要 runtime.Object，如果传入的是 typed object (如 v1.Pod)，
	// 它内部会自动处理，但为了保险起见，我们可以确保 scheme 包含了这些类型。
	// 这里简化处理，直接传入 objects。
	fakeDynamicClient := fake.NewSimpleDynamicClient(s, objects...)

	// 3. 初始化 ClusterInst
	// 注意：我们需要创建一个 ClusterInst 并注册到全局 Clusters 中，或者直接返回一个绑定了 fake client 的 Kubectl
	// 由于 kom 的设计是通过 ID 查找集群，我们需要注册它。

	k := initKubectl(&rest.Config{}, id) // config 为 empty struct

	cluster := &ClusterInst{
		ID:            id,
		Kubectl:       k,
		Client:        fakeClient,
		DynamicClient: fakeDynamicClient,
		Config:        &rest.Config{Host: "https://fake-cluster"},
		apiResources: []*metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod", Group: "", Version: "v1"},
			{Name: "deployments", Namespaced: true, Kind: "Deployment", Group: "apps", Version: "v1"},
			{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet", Group: "apps", Version: "v1"},
			{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet", Group: "apps", Version: "v1"},
			{Name: "daemonsets", Namespaced: true, Kind: "DaemonSet", Group: "apps", Version: "v1"},
			{Name: "controllerrevisions", Namespaced: true, Kind: "ControllerRevision", Group: "apps", Version: "v1"},
			{Name: "services", Namespaced: true, Kind: "Service", Group: "", Version: "v1"},
			{Name: "nodes", Namespaced: false, Kind: "Node", Group: "", Version: "v1"},
			{Name: "namespaces", Namespaced: false, Kind: "Namespace", Group: "", Version: "v1"},
			{Name: "configmaps", Namespaced: true, Kind: "ConfigMap", Group: "", Version: "v1"},
			{Name: "secrets", Namespaced: true, Kind: "Secret", Group: "", Version: "v1"},
			{Name: "events", Namespaced: true, Kind: "Event", Group: "", Version: "v1"},
			{Name: "serviceaccounts", Namespaced: true, Kind: "ServiceAccount", Group: "", Version: "v1"},
			{Name: "roles", Namespaced: true, Kind: "Role", Group: "rbac.authorization.k8s.io", Version: "v1"},
			{Name: "rolebindings", Namespaced: true, Kind: "RoleBinding", Group: "rbac.authorization.k8s.io", Version: "v1"},
			{Name: "clusterroles", Namespaced: false, Kind: "ClusterRole", Group: "rbac.authorization.k8s.io", Version: "v1"},
			{Name: "clusterrolebindings", Namespaced: false, Kind: "ClusterRoleBinding", Group: "rbac.authorization.k8s.io", Version: "v1"},
			{Name: "persistentvolumes", Namespaced: false, Kind: "PersistentVolume", Group: "", Version: "v1"},
			{Name: "persistentvolumeclaims", Namespaced: true, Kind: "PersistentVolumeClaim", Group: "", Version: "v1"},
			{Name: "storageclasses", Namespaced: false, Kind: "StorageClass", Group: "storage.k8s.io", Version: "v1"},
			{Name: "ingresses", Namespaced: true, Kind: "Ingress", Group: "networking.k8s.io", Version: "v1"},
			{Name: "customresourcedefinitions", Namespaced: false, Kind: "CustomResourceDefinition", Group: "apiextensions.k8s.io", Version: "v1"},
			{Name: "cronjobs", Namespaced: true, Kind: "CronJob", Group: "batch", Version: "v1"},
		},
	}

	// 4. 注册 fake 回调
	cluster.callbacks = k.initializeCallbacks()
	registerFakeHandlers(cluster.callbacks)

	// Initialize other fields for status testing
	cluster.serverVersion = &version.Info{Major: "1", Minor: "29", GitVersion: "v1.29.0"}
	// cluster.openAPISchema = k.getOpenAPISchema() // fake client might return nil or empty
	// cluster.describerMap = k.initializeDescriberMap()

	// 注册到全局 map
	Clusters().clusters.Store(id, cluster)

	return k
}

func registerFakeHandlers(c *callbacks) {
	c.Get().Register("fake:get", fakeGet)
	c.List().Register("fake:list", fakeList)
	c.Delete().Register("fake:delete", fakeDelete)
	c.Create().Register("fake:create", fakeCreate)
	c.Update().Register("fake:update", fakeUpdate)
	c.Patch().Register("fake:patch", fakePatch)
	c.Exec().Register("fake:exec", fakeExec)
	c.Logs().Register("fake:logs", fakeLogs)
}

func fakeExec(k *Kubectl) error {
	if k.Statement.Command == "" {
		return fmt.Errorf("command is empty")
	}
	return nil
}

func fakeLogs(k *Kubectl) error {
	if k.Statement.PodLogOptions == nil {
		return fmt.Errorf("PodLogOptions is nil")
	}
	return nil
}

func fakeCreate(k *Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	ctx := stmt.Context

	// Convert Dest to Unstructured
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(stmt.Dest)
	if err != nil {
		return err
	}
	u := &unstructured.Unstructured{Object: obj}

	// Ensure name and namespace are set if missing (though usually set in stmt)
	if u.GetName() == "" && stmt.Name != "" {
		u.SetName(stmt.Name)
	}
	if u.GetNamespace() == "" && stmt.Namespace != "" {
		u.SetNamespace(stmt.Namespace)
	}

	var res *unstructured.Unstructured

	if stmt.Namespaced {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Create(ctx, u, metav1.CreateOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Create(ctx, u, metav1.CreateOptions{})
	}

	if err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
}

func fakeUpdate(k *Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	ctx := stmt.Context

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(stmt.Dest)
	if err != nil {
		return err
	}
	u := &unstructured.Unstructured{Object: obj}

	var res *unstructured.Unstructured
	if stmt.Namespaced {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Update(ctx, u, metav1.UpdateOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Update(ctx, u, metav1.UpdateOptions{})
	}

	if err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
}

func fakePatch(k *Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	patchType := stmt.PatchType
	patchData := []byte(stmt.PatchData)

	// Workaround for fake client: convert StrategicMergePatchType to MergePatchType
	// because fake dynamic client deals with Unstructured and might fail with SMPT
	// if schema information is missing or not fully supported in the fake context.
	if patchType == types.StrategicMergePatchType {
		patchType = types.MergePatchType
	}

	var res *unstructured.Unstructured
	var err error

	if stmt.Namespaced {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	}

	if err != nil {
		return err
	}

	if stmt.Dest != nil {
		return runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
	}
	return nil
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

	opts := metav1.ListOptions{}
	if len(stmt.ListOptions) > 0 {
		opts = stmt.ListOptions[0]
	}

	var list *unstructured.UnstructuredList
	var err error

	if stmt.Namespaced {
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, opts)
	} else {
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).List(ctx, opts)
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

	var err error

	if stmt.Namespaced {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	return err
}
