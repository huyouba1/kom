package kom

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCronJob(t *testing.T) {
	RegisterFakeCluster("cj-cluster")
	k := Cluster("cj-cluster")
	ns := "default"
	name := "test-cj"

	// Create CronJob
	cj := &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{Kind: "CronJob", APIVersion: "batch/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "* * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "c", Image: "busybox"}},
							RestartPolicy: corev1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}
	err := k.Resource(cj).Create(cj).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched batchv1.CronJob
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var cjList []batchv1.CronJob
	err = k.Resource(&batchv1.CronJob{}).Namespace(ns).List(&cjList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(cjList) != 1 {
		t.Errorf("Expected 1 CJ, got %d", len(cjList))
	}

	// Test Update
	fetched.Spec.Schedule = "*/5 * * * *"
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated batchv1.CronJob
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Spec.Schedule != "*/5 * * * *" {
		t.Errorf("Expected schedule */5 * * * *, got %s", updated.Spec.Schedule)
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}
