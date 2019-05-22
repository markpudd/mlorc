// Assume only one batch queue running on cluster
package mlorc

import (
      "time"
      "fmt"
      "flag"


      apiv1 "k8s.io/api/core/v1"
      batchv1 "k8s.io/api/batch/v1"
      metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

      "k8s.io/client-go/kubernetes"
      bv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
      cv1 "k8s.io/client-go/kubernetes/typed/core/v1"
      "k8s.io/client-go/tools/clientcmd"


)
//"path/filepath"
//"k8s.io/client-go/util/homedir"
type JobQueue struct {
  jobs chan *batchv1.Job
  running bool
  batchv1client bv1.JobInterface
  podv1client cv1.PodInterface
}



func NewJobQueue() *JobQueue {
	jobq := new(JobQueue)
//  var kubeconfig *string

//  if home := homedir.HomeDir(); home != "" {
//   kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
//  } else {
//   kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
//  }
  flag.Parse()
  config, err := clientcmd.BuildConfigFromFlags("","")
  if err != nil {
   panic(err)
  }
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
   panic(err)
  }
  jobq.batchv1client  = clientset.BatchV1().Jobs(apiv1.NamespaceDefault)
  jobq.podv1client  = clientset.CoreV1().Pods(apiv1.NamespaceDefault)
  jobq.jobs = make(chan *batchv1.Job, 100)
	return jobq
}


func (jq *JobQueue)IsAnyFreeJobs() bool{
  result, err := jq.podv1client.List(metav1.ListOptions{})
  if err != nil {
         panic(err)
  }
  for _,j := range result.Items {
    fmt.Printf("Job %q %q\n",j.GetObjectMeta().GetName(),j.Status.Phase)
  }
  return true
}

func (jq *JobQueue)QueueLoop() {
    for jq.running {
      // is there free jobs on cluster
//      if jq.IsAnyFreeJobs() {
          // Unqueue jobs
          job := <- jq.jobs
          //add job to q
          fmt.Println("Creating job "+job.Name)
          result, err := jq.batchv1client.Create(job)
          if err != nil {
            panic(err)
          }
          fmt.Printf("Created job %q.\n", result.GetObjectMeta().GetName())
//      }
      // sleep to back off
      time.Sleep(1 * time.Second)
    }

}

func (jq *JobQueue)AddJob(job *batchv1.Job) {
  jq.jobs <- job
}


func (jq *JobQueue)StartJobQueue() {
  if !jq.running {
    jq.running = true;
    go jq.QueueLoop();
  }
}

func (jq *JobQueue)StopJobQueue() {
  jq.running = false;

}
