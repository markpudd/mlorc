package main

import (

        "fmt"
        "log"
        "io"
        "os"
        "io/ioutil"
        "github.com/markpudd/mlorc/mlorc"

        "net/http"
        "text/template"
 )


 var jq *mlorc.JobQueue
 var dl *mlorc.DataLogger
 var authToken string
 const file_url string = "https://data location"
 const file_loc string = "/app/data/train.csv"
 
  func CheckAuth( r *http.Request) bool{
    token := r.Header.Get("X-Auth-Token")
    return token==authToken

  }
 func DataIn(w http.ResponseWriter, r *http.Request) {
   if !CheckAuth(r) {
     fmt.Fprintf(w, "{status:noauth}")
     return
   }
    title := r.URL.Path[len("/publish/epoch/end/"):]
    dl.AddPoint(r.Body,title)
 }

 func GetGraphs(w http.ResponseWriter, r *http.Request) {
   type JsonPageData struct {
     AccData string
     LossData string
     Val_acc string
     Val_loss string
     Runs []string
   }

  data := JsonPageData{
   AccData: string(dl.GetJson("Acc")),
   LossData: string(dl.GetJson("Loss")),
   Val_acc: string(dl.GetJson("Val_acc")),
   Val_loss: string(dl.GetJson("Val_loss")),
   Runs: uploadList,
  }
//  m := make(map[string][]byte)
//  m["datajson"] = testdata
t, err := template.ParseFiles("mainpage.html")
if err != nil {
    panic(err)
}
  _ = t.Execute(w,data)
 }


 func RunJobs(w http.ResponseWriter, r *http.Request) {
   if !CheckAuth(r) {
     fmt.Fprintf(w, "{status:noauth}")
     return
   }
   jobs :=  mlorc.CreateJobsFromJson(r.Body)
   for _,job  := range jobs {
    jq.AddJob(job)
   }
   fmt.Fprintf(w, "{jobs:queued}")
 }


 func GetJobs(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintf(w, "{jobs:pending}")
 }

 func Health(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintf(w, "{status:ok}")
 }

 func Upload(w http.ResponseWriter, r *http.Request) {
   if !CheckAuth(r) {
     fmt.Fprintf(w, "{status:noauth}")
     return
   }
   title := r.URL.Path[len("/upload/"):]
   r.ParseMultipartForm(100 << 20)
   file, handler, err := r.FormFile("file")
   if err != nil {
     fmt.Println(err)
     return
   }
   defer file.Close()
   fmt.Printf("Uploaded File: %+v\n", handler.Filename)
   fmt.Printf("File Size: %+v\n", handler.Size)
    fmt.Printf("MIME Header: %+v\n", handler.Header)


    f, err := os.Create("/app/data/"+title)
    if err != nil {
       fmt.Println(err)
    }

    defer f.Close()

    fileBytes, err := ioutil.ReadAll(file)
    if err != nil {
       fmt.Println(err)
    }
    f.Write(fileBytes)
    uploadList = append(uploadList,title)
    fmt.Fprintf(w, "Successfully Uploaded File\n")
 }

 func Download(w http.ResponseWriter, r *http.Request) {
   title := r.URL.Path[len("/download/"):]
   http.ServeFile(w, r, "/app/data/"+title)
 }

 func DownloadFile(filepath string, url string) error {

     // Get the data
     resp, err := http.Get(url)
     if err != nil {
         return err
     }
     defer resp.Body.Close()

     // Create the file
     out, err := os.Create(filepath)
     if err != nil {
         return err
     }
     defer out.Close()

     // Write the body to file
     _, err = io.Copy(out, resp.Body)
     return err
 }

func main() {
  uploadList =  make([]string, 0)
  var found bool
  authToken,found = os.LookupEnv("AUTH_TOKEN")
  if !found {
    authToken = "XXXXXX"
  }
  if _, err := os.Stat(file_loc); os.IsNotExist(err) {
    DownloadFile(file_loc,file_url)
  }
   dl = mlorc.NewDataLogger()
   jq = mlorc.NewJobQueue()
   jq.StartJobQueue()
   fs := http.FileServer(http.Dir("static/"))
   http.Handle("/static/", http.StripPrefix("/static/", fs))

   http.HandleFunc("/getgraphs",GetGraphs)
   http.HandleFunc("/runjobs",RunJobs)
   http.HandleFunc("/getjobs",GetJobs)
   http.HandleFunc("/publish/epoch/end/",DataIn)
   http.HandleFunc("/upload/",Upload)
   http.HandleFunc("/download/",Download)
   http.HandleFunc("/",Health)
   log.Fatal(http.ListenAndServe(":8585", nil))
}
