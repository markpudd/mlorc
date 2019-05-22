package mlorc

import (
      "encoding/json"
      "strconv"
      "io"

      "k8s.io/apimachinery/pkg/api/resource"
      batchv1 "k8s.io/api/batch/v1"
      metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
      apiv1 "k8s.io/api/core/v1"
)

type tune struct{
  Key  string
  Values  []string
}


type jobInfo struct {
    Name string
    Image string
    Env  map[string] string
    Tune []tune
}



func CreateJobsFromJson(data io.ReadCloser) []*batchv1.Job {
  var jobinfo jobInfo
  decoder := json.NewDecoder(data)
  err := decoder.Decode(&jobinfo)
  if err != nil {
      panic(err)
  }

  jobs := []*batchv1.Job{};

  for i, param_val := range jobinfo.Tune[0].Values {
    env := []apiv1.EnvVar{};
    e1 := apiv1.EnvVar {
        Name: jobinfo.Tune[0].Key,
        Value: param_val,
     }
    env = append(env,  e1   )
	  for key := range jobinfo.Env {
      e1 := apiv1.EnvVar {
          Name: key,
          Value: jobinfo.Env[key],
       }
      env = append(env,  e1   )
    }
    qq,_ := resource.ParseQuantity("1")
    job := &batchv1.Job{
      ObjectMeta: metav1.ObjectMeta{
        Name: jobinfo.Name+"-"+strconv.Itoa(i),
      },
      Spec: batchv1.JobSpec{
        Template: apiv1.PodTemplateSpec{
          ObjectMeta: metav1.ObjectMeta{
            Labels: map[string]string{
              "app": jobinfo.Name+"-"+strconv.Itoa(i),
              "jtype" : "mljob",
            },
          },
          Spec: apiv1.PodSpec{
            Containers: []apiv1.Container{
              {
                Name:  jobinfo.Name+"-"+strconv.Itoa(i),
                Image: jobinfo.Image,
                Resources: apiv1.ResourceRequirements{
                  Limits: apiv1.ResourceList{
                     "nvidia.com/gpu" : qq,
                  },
                },
                ImagePullPolicy: "Always",
                Env: env,
                VolumeMounts: []apiv1.VolumeMount{
                  {
                    MountPath: "/data",
                    Name: "learn-vol",
                  },
                },
              },
            },
            RestartPolicy: "Never",
            Volumes: []apiv1.Volume{
                {
                  Name: "learn-vol",
                  VolumeSource: apiv1.VolumeSource{
                    PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource {
                      ClaimName: "learn-pvc",
                    },
                },
              },
            },
          },
        },
      },
    }
    jobs = append(jobs,  job   )
  }
  return  jobs
 }
