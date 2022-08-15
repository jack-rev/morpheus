package main

import(
    "fmt"
    "os"
    "log"
    "context"
    "bufio"
    "math/rand"
    "flag"

    "github.com/gookit/color"

    "k8s.io/client-go/rest"
    "k8s.io/apimachinery/pkg/watch"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

var namespace string

func buildFromKubeConfig() *rest.Config {
    home := os.Getenv("HOME")
    config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home)); if err != nil {
        fmt.Println(err)
    }
    return config
}

func iteratePods(clientset *kubernetes.Clientset, namespace string){
    pods, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
            Watch : true,
        })
    if err != nil {
        log.Fatal(err)
    }

    // ResultChan provides channel to which new pods are added
    eventChan := pods.ResultChan()
    for event := range eventChan {
        pod := event.Object.(*corev1.Pod)
        switch event.Type {
            case watch.Added:
                fmt.Printf("%v has been detected by Morpheus - tailing will begin once pod status is Running\n", pod.ObjectMeta.Name)
            case watch.Modified:
                // Wait until pod is running, but not marked for deletion
                // TODO test against other use cases such as patches/updates, jobs, etc.
                if pod.Status.Phase == corev1.PodRunning && pod.ObjectMeta.DeletionTimestamp == nil {
                    go tailPod(pod.ObjectMeta.Name, namespace, clientset)
                }
            case watch.Deleted:
                fmt.Printf("%v has been deleted - no longer tailing\n", pod.ObjectMeta.Name)
        }
    }
}

func tailPod(podName string, podNamespace string, clientset *kubernetes.Clientset){
    // TODO add a timeout to deal with non-deletion events such as lost connection to cluster
    fmt.Printf("Tailing %v\n", podName)

    // Include logs from the past 5 seconds
    //TODO parameterise this
    sinceSeconds := int64(5)

    // Get logs as a Request object then convert to IO reader via Stream()
    logs, err := clientset.CoreV1().Pods(podNamespace).GetLogs(podName, &corev1.PodLogOptions{
        Follow: true,
        SinceSeconds: &sinceSeconds,
    }).Stream(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    // Generate random colour RGB values, omitting darker ones
    r := rand.Intn(175)+80
    g := rand.Intn(175)+80
    b := rand.Intn(175)+80

    // Print logs in random colour as they are being streamed to scanner
    sc := bufio.NewScanner(logs)
    for sc.Scan() {
        color.Printf("<fg=%v,%v,%v>%v: %v</>\n", r, g, b, podName, sc.Text())
    }
    // EOF in scanner has been reached - exit GoRoutine
    return
}

func main() {
    config := buildFromKubeConfig()
    // Create client
    clientset, err := kubernetes.NewForConfig(config); if err != nil {
        log.Fatal(err)
    }

    flag.StringVar(&namespace, "namespace", "default", "namespace to run morpheus in")
    flag.Parse()

    fmt.Printf("Welcome to Morpheus - running in %v namespace\n", namespace)
    iteratePods(clientset, namespace)
}
