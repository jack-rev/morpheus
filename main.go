package main

import(
    "fmt"
    "os"
    "log"
    "time"
    _ "io"
    "context"
    _ "bytes"
    _ "strings"
    "bufio"

    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

func buildFromKubeConfig() *rest.Config {
    home := os.Getenv("HOME")
    config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home)); if err != nil {
        fmt.Println(err)
    }
    return config
}

func tailPod(podName string, podNamespace string, clientset *kubernetes.Clientset){
    // Get logs IO reader
    logs, err := clientset.CoreV1().Pods(podNamespace).GetLogs(podName, &corev1.PodLogOptions{
        Follow: true,
    }).Stream(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    // Tail logs via scanner object
    sc := bufio.NewScanner(logs)

    for sc.Scan() {
        fmt.Printf("%v pod: %v\n", podName, sc.Text())
    }
}

func main() {
    fmt.Println("Welcome to Morpheus")

    config := buildFromKubeConfig()

    // Create client
    clientset, err := kubernetes.NewForConfig(config); if err != nil {
        log.Fatal(err)
    }

    // TODO: iterate over pod list and get logs
    _, err = clientset.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        log.Fatal(err)
    }

    go tailPod("scraper", "default", clientset)
    go tailPod("scraper-down", "default", clientset)

    time.Sleep(time.Minute)

}
