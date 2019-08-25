package main

import (
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	deployment "kubeedge_demo/pkg/deployment"
	device "kubeedge_demo/pkg/device"
	devicemodel "kubeedge_demo/pkg/device_model"
	namespace "kubeedge_demo/pkg/namespace"
	node "kubeedge_demo/pkg/node"

	"github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func NewCRDClient(cfg *rest.Config) (*rest.RESTClient, error) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(addDeviceCrds)

	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	config := *cfg
	config.APIPath = "/apis"
	config.GroupVersion = &v1alpha1.SchemeGroupVersion
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		//log.LOGGER.Errorf("Failed to create REST Client due to error %v", err)
		return nil, err
	}

	return client, nil
}

func addDeviceCrds(scheme *runtime.Scheme) error {
	// Add Device
	scheme.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.Device{}, &v1alpha1.DeviceList{})
	v1.AddToGroupVersion(scheme, v1alpha1.SchemeGroupVersion)
	// Add DeviceModel
	scheme.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.DeviceModel{}, &v1alpha1.DeviceModelList{})
	v1.AddToGroupVersion(scheme, v1alpha1.SchemeGroupVersion)
	return nil
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
		return
	}

	crdClient, err := NewCRDClient(cfg)
	if err != nil {
		klog.Fatalf("Error building CRDClient error: %s", err.Error())
		return 
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "DELETE", "POST", "GET"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	nsRouter := namespace.Namespace{ClientSet: kubeClient}
	nodeRouter := node.Node{ClientSet: kubeClient}
	deploymentRouter := deployment.Deployment{ClientSet: kubeClient}
	devicemodelRouter := devicemodel.DeviceModel{Client: crdClient}
	deviceRouter := device.Device{Client: crdClient}

	router.POST("/v1/cluster/id/namespace", nsRouter.AddNamespace)
	router.DELETE("/v1/cluster/id/namespace/:name", nsRouter.DeleteNamespace)
	router.GET("/v1/cluster/id/namespaces", nsRouter.ListNamespace)
	router.GET("/v1/cluster/id/namespace/:name", nsRouter.GetNamespace)
	router.PUT("/v1/cluster/id/namespace/:name", nsRouter.UpdateNamespace)

	router.POST("vl/cluster/id/node", nodeRouter.AddNode)
	router.DELETE("vl/cluster/id/node/:name", nodeRouter.DeleteNode)
	router.GET("vl/cluster/id/nodes", nodeRouter.ListNode)
	router.GET("vl/cluster/id/node/:name", nodeRouter.GetNode)
	router.PUT("vl/cluster/id/node", nodeRouter.UpdateNode)

	router.POST("vl/cluster/id/deployment", deploymentRouter.AddDeployment)
	router.DELETE("vl/cluster/id/deployment/:name", deploymentRouter.DeleteDeployment)
	router.GET("vl/cluster/id/deployments", deploymentRouter.ListDeployment)
	router.GET("vl/cluster/id/deployment/:name", deploymentRouter.GetDeployment)
	router.PUT("vl/cluster/id/deployment/:name", deploymentRouter.UpdateDeployment)

	router.POST("vl/cluster/id/devicemodel", devicemodelRouter.AddDeviceModel)
	router.GET("vl/cluster/id/devicemodel/:name", devicemodelRouter.GetDeviceModel)
	router.GET("vl/cluster/id/devicemodels", devicemodelRouter.ListDeviceModel)
	router.DELETE("vl/cluster/id/devicemodel/:name", devicemodelRouter.DeleteDeviceModel)
	router.PUT("vl/cluster/id/devicemodel/:name", devicemodelRouter.UpdateDeviceModel)

	router.POST("vl/cluster/id/deviceinstance", deviceRouter.AddDevice)
	router.GET("vl/cluster/id/deviceinstance/:name", deviceRouter.GetDevice)
	router.GET("vl/cluster/id/deviceinstances", deviceRouter.ListDevice)
	router.DELETE("vl/cluster/id/deviceinstance/:name", deviceRouter.DeleteDevice)
	router.PUT("vl/cluster/id/deviceinstance/:name", deviceRouter.UpdateDevice)
	
	router.Run(":8000")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluste        r.")
}
