package zaproxy

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	zapv1alpha1 "github.com/omerlh/zap-operator/pkg/apis/zaproxy/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/api/extensions/v1beta1"
)

var log = logf.Log.WithName("controller_zaproxy")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Zaproxy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZaproxy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("zaproxy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Zaproxy
	err = c.Watch(&source.Kind{Type: &zapv1alpha1.Zaproxy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Zaproxy
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zapv1alpha1.Zaproxy{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileZaproxy implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileZaproxy{}

// ReconcileZaproxy reconciles a Zaproxy object
type ReconcileZaproxy struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Zaproxy object and makes changes based on the state read
// and what is in the Zaproxy.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileZaproxy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zaproxy")

	// Fetch the Zaproxy instance
	instance := &zapv1alpha1.Zaproxy{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	targetIngress := &v1beta1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.TargetIngress, Namespace: instance.Spec.TargetNamespace}, targetIngress)
	if err != nil {
		reqLogger.Info("Target ingress not found", "Ingress.Namespace", instance.Spec.TargetNamespace, "Ingress.Name", instance.Spec.TargetIngress)
		return reconcile.Result{}, err
	}

	serviceName := targetIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName
	targetHost := targetIngress.Spec.Rules[0].Host

	targetService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: instance.Spec.TargetNamespace}, targetService)
	if err != nil {
		reqLogger.Info("Target service not found", "Service.Namespace", instance.Spec.TargetNamespace, "Service.Name", serviceName)
		return reconcile.Result{}, err
	}

	serviceIp := targetService.Spec.ClusterIP

	// Define a new Pod object
	pod := newPodForCR(instance, serviceIp, targetHost)

	// Set Zaproxy instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	}

	// Define a new Pod object
	service := newServiceForCR(instance)

	// Set Zaproxy instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found2 := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: pod.Namespace}, found2)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}

	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
	}

	// Define a new Pod object
	ingress := newCanaryIngressForCR(instance, targetHost)

	// Set Zaproxy instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ingress, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found3 := &v1beta1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ingress.Name, Namespace: pod.Namespace}, found3)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Canary Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
		err = r.client.Create(context.TODO(), ingress)
		if err != nil {
			return reconcile.Result{}, err
		}

	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Ingress already exists",  "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *zapv1alpha1.Zaproxy, targetServiceIp string, targetHost string) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "zaproxy",
					Image:   "soluto/zap-ci:1551816665660",
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							Protocol:      corev1.ProtocolTCP,
							ContainerPort: 8090,
						},
					},
					LivenessProbe:  &corev1.Probe{
						Handler:  httpGetHandler("/", 8090),
						InitialDelaySeconds: 60,
						FailureThreshold: 10,
					},
				},
			},
			HostAliases: []corev1.HostAlias{
				{
					IP: targetServiceIp,
					Hostnames: []string{targetHost},
				},
			},
		},
	}
}

// newServiceForCR returns a busybox service with the same name/namespace as the cr
func newServiceForCR(cr *zapv1alpha1.Zaproxy) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-service",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": cr.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(8090),
				},
			},
		},
	}
}

// newServiceForCR returns a busybox service with the same name/namespace as the cr
func newCanaryIngressForCR(cr *zapv1alpha1.Zaproxy, targetHost string) *v1beta1.Ingress {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-ingress",
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string {
				"nginx.ingress.kubernetes.io/canary": "true",
    			"nginx.ingress.kubernetes.io/canary-weight": "5",
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: targetHost,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: cr.Name + "-service",
										ServicePort: intstr.IntOrString{IntVal: 80},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return nil
}

//source: https://github.com/divinerapier/learn-kubernetes/blob/a481964b876e07f255915699b6ab522d279329a1/test/e2e/common/container_probe.go
func httpGetHandler(path string, port int) corev1.Handler {
	return corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: path,
			Port: intstr.FromInt(port),
		},
	}
}
