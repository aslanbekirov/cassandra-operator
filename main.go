package main

import(

	"time"
	"fmt"
	"reflect"

	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	co_clientset "github.com/aslanbekirov/cassandra-operator/pkg/client/clientset/versioned"
	co_v1aplha1 "github.com/aslanbekirov/cassandra-operator/pkg/apis/cassandra.database.com/v1alpha1"
	external_versions "github.com/aslanbekirov/cassandra-operator/pkg/client/informers/externalversions"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"flag"
)

func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}


func main(){
	kubeconf := flag.String("kubeconf", "kube.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := GetClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// create clientset and create our crd, this only need to run once
	clientset, err := co_clientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	
	cassandra_cluster := &v1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{Name: co_v1aplha1.FullCRDName},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   co_v1aplha1.CRDGroup,
			Version: co_v1aplha1.CRDVersion,
			Scope:   v1beta1.NamespaceScoped,
			Names:   v1beta1.CustomResourceDefinitionNames{
				Plural: co_v1aplha1.CRDPlural,
				Kind:   reflect.TypeOf(co_v1aplha1.CassandraCluster{}).Name(),
			},
		},
	}

	  
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(config)
    result, err := apiextensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(cassandra_cluster)

	if err!=nil{
		fmt.Println("Erro occured creating cassandra cluster crd, %v", err)
		panic(err)
	}


	time.Sleep(5 * time.Second)

	if err == nil {
		fmt.Printf("CREATED: %#v\n", result)
	} else if apierrors.IsAlreadyExists(err) {
		fmt.Printf("ALREADY EXISTS: %#v\n", result)
	} else {
		panic(err)
	}

	// items, err := clientset.AslangroupV1().Persons("test").List(meta_v1.ListOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	factory:=external_versions.NewSharedInformerFactory(clientset, time.Minute*3)
	informer, err:= factory.ForResource(co_v1aplha1.SchemeGroupVersion.WithResource("people"))
	informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Printf("add: %s \n", obj)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Printf("delete: %s \n", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Printf("Update old: %s \n      New: %s\n", oldObj, newObj)
		},
	},
	)
	stop := make(chan struct{})

	go informer.Informer().Run(stop)

	select{}
}