package main

import(
    "fmt"
    "os"
    "log"
    "context"
    "bufio"
    "sync"
    "math/rand"
    "time"
    "github.com/gookit/color"

    "k8s.io/client-go/rest"
    watch "k8s.io/apimachinery/pkg/watch"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

var wg sync.WaitGroup
var Reset  = "\033[0m"

func buildFromKubeConfig() *rest.Config {
    home := os.Getenv("HOME")
    config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home)); if err != nil {
        fmt.Println(err)
    }
    return config
}

func iteratePods(clientset *kubernetes.Clientset){
    namespace := "default"

    pods, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
            Watch : true,
        })
    if err != nil {
        log.Fatal(err)
    }

    wg.Add(1)
    eventChan := pods.ResultChan()
    for event := range eventChan {
        pod := event.Object.(*corev1.Pod)
        switch event.Type {
            case watch.Added:
                fmt.Printf("Pod %v has been added - tailing to begin once pod status is Running\n", pod.ObjectMeta.Name)
                go tailPod(pod.ObjectMeta.Name, namespace, clientset)
            case watch.Deleted:
                fmt.Printf("Pod %v has been deleted - stopping tailing\n", pod.ObjectMeta.Name)
        }
    }
    wg.Wait()
}

func tailPod(podName string, podNamespace string, clientset *kubernetes.Clientset){
    for {
        // This works but feel like it could be better written... are chans an option here?
        pod, err := clientset.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
        if err != nil {
            log.Fatal(err)
        }
        time.Sleep(time.Second)
        if pod.Status.Phase == corev1.PodRunning {
            break
        }
    }

    fmt.Printf("Now tailing pod %v\n", podName)

    sinceSeconds := int64(5)

    // Get logs IO reader
    logs, err := clientset.CoreV1().Pods(podNamespace).GetLogs(podName, &corev1.PodLogOptions{
        Follow: true,
        SinceSeconds: &sinceSeconds,
    }).Stream(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    r := rand.Intn(155)+100
    g := rand.Intn(155)+100
    b := rand.Intn(155)+100

    // Tail logs via scanner object
    // Logs are not being received from the eventChan
    // As a challenge, see if you can rewrite this using channels rather than a stream?
    sc := bufio.NewScanner(logs)
    for sc.Scan() {
        color.Printf("<fg=%v,%v,%v>%v: %v</>\n", r, g, b, podName, sc.Text())
    }

    return
}

func main() {

    fmt.Println("Welcome to Morpheus")

    config := buildFromKubeConfig()

    // Create client
    clientset, err := kubernetes.NewForConfig(config); if err != nil {
        log.Fatal(err)
    }

    iteratePods(clientset)

}
