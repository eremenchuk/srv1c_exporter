package main

import (
    "bytes"
    "fmt"
    "log"
    "flag"
    "net/http"
    "os/exec"
    "strings"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "time"
)

var (
  addr = flag.String("listen-address", ":9191",
    "The address to listen on for HTTP requests.")

  session_count = prometheus.NewGauge(
    prometheus.GaugeOpts{
      Name: "srv1c_session_count",
      Help: "Number of 1C session.",
    })

  cluster_id = "";
)

func main() {
  flag.Parse()

  prometheus.MustRegister(session_count)

  cmd := exec.Command("/opt/1C/v8.3/x86_64/rac", "cluster", "list")
  //cmd.Stdin = strings.NewReader("cluster list")
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()
  if err != nil {
      log.Fatal(err)
  }

  strs := strings.Split(out.String(),"\n")
  for i := 0; i < len(strs); i++ {
    str := strings.Split(strs[i], ":")
    //fmt.Println(str[0]);
    if strings.TrimSpace(str[0]) == "cluster" {
      cluster_id = strings.TrimSpace(str[1]);
      break;
    }
  }
  fmt.Println(cluster_id);

  go session_list()

  http.Handle("/metrics", promhttp.Handler())

  log.Printf("Starting web server at %s\n", *addr)
  err = http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Printf("http.ListenAndServer: %v\n", err)
  }
}

func session_list() {
//  /opt/1C/v8.3/x86_64/rac session list --cluster=e43cfe60-93c8-11ea-1495-96000053119a
  for {
    cmd := exec.Command("/opt/1C/v8.3/x86_64/rac", "session", "list", "--cluster=" + cluster_id)
    var out bytes.Buffer
    cmd.Stdout = &out;
    err := cmd.Run()
    if err != nil {
      log.Fatal(err)
    }
    //fmt.Println(out.String());

    strs := strings.Split(out.String(),"\n\n")

    session_count.Set(float64(len(strs) - 1))

    time.Sleep(10000 * time.Millisecond)
  }
}
