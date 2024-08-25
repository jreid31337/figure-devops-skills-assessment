package main

import (
	"flag"
	"path/filepath"
	"fmt"
	"strings"
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

//	appsV1 "k8s.io/api/apps/v1"
//	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8stypes "k8s.io/apimachinery/pkg/types"
//	"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main(){
	// TODO: move this to an arg if we were to productionify this
	//restart_name_match := "nginx"
	restart_name_match := "database"

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create an API Clientset (k8s.io/client-go/kubernetes)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create an AppsV1Client (k8s.io/client-go/kubernetes/typed/apps/v1)
	appsV1Client := clientset.AppsV1()

	// https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/apps/v1#DeploymentInterface
	deployments, err := appsV1Client.Deployments("").List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d deployments in the cluster\n", len(deployments.Items))
	for _, d := range deployments.Items {
		deployName:=d.GetName()
		deployNamespace := d.GetNamespace()
//		fmt.Println(deployName, deployNamespace)

		if (strings.Contains(deployName, restart_name_match)){
			fmt.Println("matched deployment: ", deployName)
//			fmt.Println("kubeconfig: ", *kubeconfig)

			ctx := context.TODO()

			// https://stackoverflow.com/questions/61335318/how-to-restart-a-deployment-in-kubernetes-using-go-client
			deploy, err := clientset.AppsV1().Deployments(deployNamespace).Get(ctx, deployName, metaV1.GetOptions{})
			fmt.Println(deploy)
			if err != nil {
				panic(err.Error())
			}

			deploymentsClient := clientset.AppsV1().Deployments(deployNamespace)
			data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format(time.RFC3339))
			fmt.Println("Patching deployment with: ", data)
			deployment, err := deploymentsClient.Patch(ctx, deployName, k8stypes.StrategicMergePatchType, []byte(data), metaV1.PatchOptions{})
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("restarted deployment: ", deployment)
		}		
	}
}
