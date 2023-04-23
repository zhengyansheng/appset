package controllers

import (
	"context"
	"reflect"

	appv1 "github.com/zhengyansheng/appset/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	cpuRequest = "100m"
	cpuLimit   = "100m"
	memRequest = "512Mi"
	memLimit   = "512Mi"
)

func (r *AppSetReconciler) UpCreateDeployment(ctx context.Context, req ctrl.Request, instance *appv1.AppSet) error {
	found := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, found)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// new deployment
			deployment := generatorDeployment(instance)

			klog.Info("set reference by ingress")
			if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
				klog.Error(err, "set deployment controller reference err")
				return err
			}

			// create deployment
			if err := r.Create(ctx, deployment); err != nil {
				klog.Errorf("create deployment err: %v", err)
				return err
			}
			return nil
		}
		return err
	}
	//
	//// update fields replicas/ image
	expected := generatorDeployment(instance)
	klog.Infof("found spec replicas: %+v\n", found.Spec.Replicas)
	klog.Infof("expected spec replicas: %+v\n", expected.Spec.Replicas)
	if !reflect.DeepEqual(found.Spec, expected.Spec) {
		found.Spec = expected.Spec
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}

	return nil
}

func (r *AppSetReconciler) CreateServiceIfNotExists(ctx context.Context, req ctrl.Request, instance *appv1.AppSet) error {
	found := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, found); err != nil {
		if apierrors.IsNotFound(err) {
			// new service
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: instance.Namespace,
					Name:      instance.Name,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(instance, schema.GroupVersionKind{
							Group:   appv1.GroupVersion.Group,
							Version: appv1.GroupVersion.Version,
							Kind:    "AppSet",
						}),
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{
						Name:       "http",
						Port:       instance.Spec.ExposePort,
						TargetPort: intstr.IntOrString{Type: 0, IntVal: 80},
					}},
					Selector: map[string]string{
						"app": instance.Name,
					},
					Type: corev1.ServiceTypeClusterIP,
				},
			}
			if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
				klog.Error(err, "set service controller reference err")
				return err
			}

			// create service
			if err := r.Create(ctx, service); err != nil {
				klog.Error(err, "create service err")
				return err
			}
			klog.Info("create service finish")
		}
	}
	return nil
}

func (r *AppSetReconciler) CreateIngressIfNotExists(ctx context.Context, req ctrl.Request, instance *appv1.AppSet) error {
	found := &networkv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, found); err != nil {
		if apierrors.IsNotFound(err) {
			// new ingress
			pathType := networkv1.PathTypeImplementationSpecific
			ingress := &networkv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: instance.Namespace,
					Name:      instance.Name,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(instance, schema.GroupVersionKind{
							Group:   appv1.GroupVersion.Group,
							Version: appv1.GroupVersion.Version,
							Kind:    "AppSet",
						}),
					},
				},
				Spec: networkv1.IngressSpec{
					Rules: []networkv1.IngressRule{
						{
							Host: instance.Spec.ExposeDomain,
							IngressRuleValue: networkv1.IngressRuleValue{
								HTTP: &networkv1.HTTPIngressRuleValue{
									Paths: []networkv1.HTTPIngressPath{
										{
											Path:     "/",
											PathType: &pathType,
											Backend: networkv1.IngressBackend{
												Service: &networkv1.IngressServiceBackend{
													Name: instance.Name,
													Port: networkv1.ServiceBackendPort{
														//Name:   defaultPortName,
														Number: instance.Spec.ExposePort,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			klog.Info("set reference by ingress")
			if err := controllerutil.SetControllerReference(instance, ingress, r.Scheme); err != nil {
				klog.Error(err, "set ingress controller reference err")
				return err
			}

			// create ingress
			if err := r.Create(ctx, ingress); err != nil {
				klog.Errorf("create ingress err: %v", err)
				return err
			}
			klog.Info("create ingress success")
		}
	}
	return nil
}

// statusUpdate update appset status
func (r *AppSetReconciler) statusUpdate(ctx context.Context, instance *appv1.AppSet) error {
	if instance.Status.Replicas == &instance.Spec.Replicas {
		return nil
	}

	instance.Status.Replicas = &instance.Spec.Replicas
	instance.Status.Ready = true
	if err := r.Status().Update(ctx, instance); err != nil {
		klog.Errorf("appset update err: %v", err)
		return err
	}

	return nil
}

func generatorDeployment(instance *appv1.AppSet) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: instance.Namespace,
			Name:      instance.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instance, schema.GroupVersionKind{
					Group:   appv1.GroupVersion.Group,
					Version: appv1.GroupVersion.Version,
					Kind:    "AppSet",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": instance.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": instance.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            instance.Name,
							Image:           instance.Spec.Image,
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolSCTP,
									ContainerPort: instance.Spec.ExposePort,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse(cpuRequest),
									"memory": resource.MustParse(memRequest),
								},
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse(cpuLimit),
									"memory": resource.MustParse(memLimit),
								},
							},
						},
					},
				},
			},
		},
	}
}
