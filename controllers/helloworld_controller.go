/*
Copyright 2024.

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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hellov1 "github.com/judeniroshan/hello-operator/api/v1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hello.org.demo,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hello.org.demo,resources=helloworlds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hello.org.demo,resources=helloworlds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelloWorld object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	log := r.Log.WithValues("helloworld", req.NamespacedName)

	// Fetch the HelloWorld instance
	hw := &hellov1.HelloWorld{}
	err := r.Get(ctx, req.NamespacedName, hw)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Construct the expected deployment name
	deploymentName := fmt.Sprintf("nginx_%s", hw.Name)

	// Check if the deployment exists
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: req.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Deployment not found", "Deployment.Namespace", req.Namespace, "Deployment.Name", deploymentName)
		// Optionally, you can create the deployment here if it does not exist
		return ctrl.Result{}, nil
	} else if err != nil {
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// If deployment is found, log the details
	log.Info("Deployment found", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hellov1.HelloWorld{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *HelloWorldReconciler) createNginxDeployment(ctx context.Context, hw *hellov1.HelloWorld) (*appsv1.Deployment, error) {
	deploymentName := fmt.Sprintf("nginx_added_from_%s", hw.Name)

	labels := map[string]string{
		"app":   "jude_added_nginx",
		"hw_cr": hw.Name,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: hw.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(hw, deployment, r.Scheme); err != nil {
		return nil, err
	}

	err := r.Create(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func int32Ptr(i int32) *int32 {
	return &i
}
