package mlorc

import (
	"testing"
  "strconv"
)

func TestCreateJobsFromJson(t *testing.T) {
  testdata := []byte(`
      {
          "name": "test_ml",
          "image": "test_image",
          "tune": [
              {
                "key": "lr",
                "values": [ "0.1","0.2","0.5","1"]
              }
          ]
      }
  `)
  vals := []string{"0.1","0.2","0.5","1"}
  jobs :=  CreateJobsFromJson(testdata)
  if(len(jobs) !=4 ) {
    t.Errorf("Jobs len is not 4")
  }

  for i, job := range jobs {
    name := "test_ml-"+strconv.Itoa(i)
  	if job == nil {
  		t.Errorf("Job is nil")
  	} else {
  		if job.Name != name {
  			t.Errorf("Name is not "+name)
      }
      if job.Spec.Template.ObjectMeta.Labels["app"] != name {
          t.Errorf("Label app not set not set to "+name)
      }
      containers := job.Spec.Template.Spec.Containers
      if(len(containers) ==0 ) {
        t.Errorf("No Containers")
      }
      if(containers[0].Name != name) {
        t.Errorf("Container name not "+name)
      }
      if(containers[0].Image != "test_image" ) {
        t.Errorf("Container image not test_image")
      }
      if(containers[0].ImagePullPolicy != "Never" ) {
        t.Errorf("Container ImagePullPolicy not Never")
      }
      env := containers[0].Env
      if(len(env) ==0 ) {
        t.Errorf("No Envronment")
      }
      if(env[0].Name != "lr" ) {
        t.Errorf("Env Name not lr")
      }
      if(env[0].Value != vals[i] ) {
        t.Errorf("Env Value not set correctly")
      }

      if(job.Spec.Template.Spec.RestartPolicy != "Never" ) {
        t.Errorf("RestartPolicy not Never")
      }

    }
	}
}
