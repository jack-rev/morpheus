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

    // wg.Done() is never called, so Morpheus can run indefinitely and listen for new pods
    wg.Add(1)

    // ResultChan provides channel to which new pods are added
    eventChan := pods.ResultChan()
    for event := range eventChan {
        pod := event.Object.(*corev1.Pod)
        switch event.Type {
            case watch.Added:
                fmt.Printf("%v has been added - tailing to begin once pod status is Running\n", pod.ObjectMeta.Name)
                go tailPod(pod.ObjectMeta.Name, namespace, clientset)
            case watch.Deleted:
                fmt.Printf("%v has been deleted - ceasing tailing\n", pod.ObjectMeta.Name)
        }
    }
    wg.Wait()
}

func tailPod(podName string, podNamespace string, clientset *kubernetes.Clientset){

    // Wait until pod is up and running
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

    fmt.Printf("Tailing pod %v\n", podName)

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

    // Generate random colour, omitting darker colours
    r := rand.Intn(175)+80
    g := rand.Intn(175)+80
    b := rand.Intn(175)+80

    // Print logs as they are being streamed via scanner object, giving tail-like functionality
    sc := bufio.NewScanner(logs)
    for sc.Scan() {
        color.Printf("<fg=%v,%v,%v>%v: %v</>\n", r, g, b, podName, sc.Text())
    }

    // EOF in scanner has been reached as pod has been deleted or is no longer available - exit GoRoutine
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
