package main

import(
    "fmt"
    "os"
    "log"
    "io"
    "context"
    "bytes"

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

func main() {
    fmt.Println("Welcome to Morpheus")

    config := buildFromKubeConfig()

    // Create client
    clientset, err := kubernetes.NewForConfig(config); if err != nil {
        log.Fatal(err)
    }

    // TODO: iterate over pod list and get logs
    _, err = clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        log.Fatal(err)
    }

    // Get logs IO reader
    req := clientset.CoreV1().Pods("default").GetLogs("scraper", &corev1.PodLogOptions{})
    logs, err := req.Stream(context.TODO()); if err != nil {
        log.Fatal(err)
    }

    // Copy reader to buffer
    buf := new(bytes.Buffer)
    _, err = io.Copy(buf, logs); if err != nil {
        log.Fatal(err)
    }

    logsAsString := buf.String()
    fmt.Println(logsAsString)

}
