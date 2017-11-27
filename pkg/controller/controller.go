package controller

import (
	"errors"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	podLabelUserID  = "user-id"
	podLabelProject = "project-name"
)

// ControllerService defines the primary
// operations for the Controller Service.
type ControllerService interface {
	LaunchContainer(ContainerInfo)
	CallService(Project) (string, error)
}

type ControllerServiceProvider struct {
	client     *kubernetes.Clientset
	launchPort int
}

// InitControllerServiceProvider defines the constructor for
// initializing the provider that interacts with the cluster.
func InitControllerServiceProvider() (*ControllerServiceProvider, error) {
	//config, err := clientcmd.BuildConfigFromFlags("", confPath)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ControllerServiceProvider{
		client: client,
	}, nil
}

// LaunchContainer launches the given container on the cluster.
func (csp *ControllerServiceProvider) LaunchContainer(ci ContainerInfo) {
	pod, err := csp.client.CoreV1().Pods(v1.NamespaceDefault).Create(&v1.Pod{
	//metav1.TypeMeta.Kind: "pod",
	//metav1.ObjectMeta{
	//	Name: pi.UserID + pi.ProjectName,
	//},
	})
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  ci.ProjectInfo.UserID + "-" + ci.ProjectInfo.ProjectName,
				Image: ci.ImageURL,
				Ports: []v1.ContainerPort{
					{
						ContainerPort: int32(csp.launchPort),
					},
				},
			},
		},
	}
	pod.SetLabels(map[string]string{
		podLabelUserID:  ci.ProjectInfo.UserID,
		podLabelProject: ci.ProjectInfo.ProjectName,
	})

	// Update Pod to include the labels
	pod, err = csp.client.Core().Pods(v1.NamespaceDefault).Update(pod)

	svc, err := csp.client.Core().Services(v1.NamespaceDefault).Create(&v1.Service{
		//ObjectMeta: v1.ObjectMeta{
		//	Name: "my-service",
		//},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeClusterIP,
			Selector: pod.Labels,
			Ports: []v1.ServicePort{
				{
					Port: int32(csp.launchPort),
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(svc.Spec.Ports[0].NodePort)
}

func (csp *ControllerServiceProvider) CallService(pi Project) (string, error) {
	serviceIP, err := csp.FindServiceIP(pi)
	if err != nil {
		fmt.Println("ERROR")
		return "", err
	}

	fmt.Println(serviceIP)
	return serviceIP, nil
}

func (csp *ControllerServiceProvider) FindServiceIP(pi Project) (string, error) {
	svcs, err := csp.client.CoreV1().Services(v1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: getLabelSelector(pi),
	})

	if err != nil {
		return "", errors.New("controller: service not found")
	}

	if len(svcs.Items) > 0 {
		return "", errors.New("controller: multiple services found")
	}
	return svcs.Items[0].Spec.ClusterIP, nil
}

func getLabelSelector(pi Project) string {
	return fmt.Sprintf("%s=%s,%s=%s", podLabelUserID, pi.UserID, podLabelProject, pi.ProjectName)
}
