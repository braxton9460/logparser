//
//
//
// I formally appologize to any real golang coders for the following
// This is my first go at go (heh)
// @author Braxton Householder <github: braxton9460>
//
package main

import (
  "encoding/json"
  "net/http"
  "bytes"
  "os"
  "fmt"
  "strconv"
  "github.com/hpcloud/tail"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type LogFormat struct {
  JSON  struct {
    Service string    `json:"UPSTREAM_SERVICE"`
    Tts     float64    `json:"RESPONSE_TIME,string"`
    Status  int    `json:"RESPONSE_STATUS,string"`
    Method  string    `json:"REQUEST_METHOD"`
  } `json:"JSON"`
}

var MonitoredServices = map[string]bool {
  "registrationapi": true,
  "exhibitorapi": true,
  "adminapi": true,
  "floorplanapi": true,
  "abstractapi": true,
}

// Fallback "names" to inject values into if they don't match a defined name
var FallbackService = "undefined"
var FallbackStatus int = 99
var FallbackMethod = "UNDEF"

// Counter names
var CounterRequests = "count"
var CounterTts = "total_tts"

// Web configs
var WebPort string = "9332"
var WebEndpoint string = "/metrics"

var MonitoredStatuses = map[int]bool {
    200: true,
    201: true,
    202: true,
    204: true,
    300: true,
    301: true,
    302: true,
    304: true,
    307: true,
    400: true,
    401: true,
    403: true,
    404: true,
    418: true,
    429: true,
    500: true,
    502: true,
    503: true,
    504: true,
}

var MonitoredMethods = map[string]bool {
    "GET": true,
    "POST": true,
    "PUT": true,
    "DELETE": true,
    "OPTION": true,
}

var Stats = make(map[string]map[int]map[string]map[string]float64)
// Do I actually need to do this here?
//var Stats = make(map[string]map[int]map[string]map[string]int)
// Do I actually need to do this here?
//var Stats = make(map[string]map[int]map[string]map[string]int)

func PrintStats(w http.ResponseWriter, r *http.Request) {
  var b bytes.Buffer
  // Print the totals
  b.WriteString("# HELP logstat_request_total logstat_request_total\n")
  b.WriteString("# TYPE logstat_request_total counter\n")
  for service, sdata := range Stats {
    for status, stdata := range sdata {
      for method, mdata := range stdata {
        b.WriteString("logstat_request_total")
        b.WriteString("{service=\"")
        b.WriteString(service)
        b.WriteString("\",status=\"")
        b.WriteString(strconv.Itoa(status))
        b.WriteString("\",method=\"")
        b.WriteString(method)
        b.WriteString("\"} ")
        b.WriteString(strconv.Itoa(int(mdata[CounterRequests])))
        b.WriteString("\n")
      }
    }
  }
  // Print the time to serve
  b.WriteString("# HELP logstat_request_tts_total logstat_request_tts_total\n")
  b.WriteString("# TYPE logstat_request_tts_total counter\n")
  for service, sdata := range Stats {
    for status, stdata := range sdata {
      for method, mdata := range stdata {
        b.WriteString("logstat_request_tts_total")
        b.WriteString("{service=\"")
        b.WriteString(service)
        b.WriteString("\",status=\"")
        b.WriteString(strconv.Itoa(status))
        b.WriteString("\",method=\"")
        b.WriteString(method)
        b.WriteString("\"} ")
        b.WriteString(strconv.FormatFloat(mdata[CounterTts], 'f', 3, 64))
        b.WriteString("\n")
      }
    }
  }
  w.Write([]byte(b.String()))
}
  w.Write([]byte(b.String()))
}

func WebListener() {
  http.HandleFunc(WebEndpoint, PrintStats)
  fmt.Println("Starting web listener on :" + WebPort + WebEndpoint)
  if err := http.ListenAndServe(":" + WebPort, nil); err != nil {
    panic(err)
  }
}

func main() {

  // Start the webserver
  go WebListener()

  // Initialize the Stats map
  // I also appologize for this..
  // It feels like there is probably a nice go way to do this.. but I couldn't figure it out
  Stats[FallbackService] = make(map[int]map[string]map[string]float64)
  Stats[FallbackService][FallbackStatus] = make(map[string]map[string]float64)
  Stats[FallbackService][FallbackStatus][FallbackMethod] = make(map[string]float64)
  for service, _ := range MonitoredServices {
    Stats[service] = make(map[int]map[string]map[string]float64)
    Stats[service][FallbackStatus] = make(map[string]map[string]float64)
    Stats[service][FallbackStatus][FallbackMethod] = make(map[string]float64)
    for status, _ := range MonitoredStatuses {
      Stats[service][status] = make(map[string]map[string]float64)
      Stats[FallbackService][status] = make(map[string]map[string]float64)
      Stats[FallbackService][status][FallbackMethod] = make(map[string]float64)
      for method, _ := range MonitoredMethods {
        Stats[service][status][method] = make(map[string]float64)
        Stats[service][status][FallbackMethod] = make(map[string]float64)
        Stats[service][FallbackStatus][method] = make(map[string]float64)
        Stats[FallbackService][status][method] = make(map[string]float64)
        Stats[FallbackService][FallbackStatus][method] = make(map[string]float64)
      }
    }
  }

//  Stats["registrationapi"][200]["GET"][CounterTts] = 2
//  fmt.Println(Stats["registrationapi"][200]["GET"][CounterTts])
//  fmt.Printf("%+v\n", Stats)
  // Get the first arg
  log_file := os.Args[1]
  // Get the first arg
  log_file := os.Args[1]
  seek := &tail.SeekInfo{
    Offset: 0,
    Whence: 2, // io.SeekEnd
  }
  t, err := tail.TailFile(log_file, tail.Config{Follow: true, Location: seek})
  check(err)
  for line := range t.Lines {
    // Convert from JSON
    line_data := LogFormat{}
    json.Unmarshal([]byte(line.Text), &line_data)

    // Check the service for validity
    if ! MonitoredServices[line_data.JSON.Service] {
      line_data.JSON.Service = FallbackService
    }
    // Check the status for validity
    if ! MonitoredStatuses[line_data.JSON.Status] {
      line_data.JSON.Status = FallbackStatus
    }
    // Check the method for validity
    if ! MonitoredMethods[line_data.JSON.Method] {
      line_data.JSON.Method = FallbackMethod
    }

    // Add it to the stats
    Stats[line_data.JSON.Service][line_data.JSON.Status][line_data.JSON.Method][CounterRequests]++
    Stats[line_data.JSON.Service][line_data.JSON.Status][line_data.JSON.Method][CounterTts] += line_data.JSON.Tts
//    fmt.Printf("%+v\n", Stats[line_data.Service][line_data.Status][line_data.Method])
//    fmt.Printf("%+v\n", line_data)
//    fmt.Println(line.Text)
  }
}
