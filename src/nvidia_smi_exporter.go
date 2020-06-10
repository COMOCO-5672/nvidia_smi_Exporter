package main

import (
    "bytes"
    "encoding/csv"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/common/promlog/flag"
    "os"
    "strconv"
    "strings"

    //"encoding/json"
    "fmt"
    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/log/level"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/common/promlog"
    "github.com/prometheus/common/version"
    "gopkg.in/alecthomas/kingpin.v2"
    "net/http"
    //"log"
    //"os"
    "os/exec"
)

const (
    namespace = "nvidia"
)
var (
    up = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "up"),
        "Was the last query of Consul successful.",
        nil, nil,
    )
    GPU = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "GPU"),
        "GPUs Info",
        []string{"index","name","utilization_gpu","utilization_memory","memory_total","memory_free","memory_used",
            "temperature_gpu","power_draw","power_limit"}, nil,
    )

     /*
     index = prometheus.NewDesc(
         prometheus.BuildFQName(namespace,"","GPU_Index"),
         "Zero based index of the GPU. Can change at each boot.",
         []string{"addr","index"},nil)

     name= prometheus.NewDesc(
         prometheus.BuildFQName(namespace, "", "Name"),
         "The official product name of the GPU. This is an alphanumeric string. For all products.",
         []string{"addr","index","name"}, nil,
     )
     */

    utilization_gpu = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "utilization_gpu"),
        "Percent of time over the past sample period during which one or more kernels was executing on the GPU.",
        []string{"index","name"}, nil,
    )
    utilization_memory=prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "utilization_memory"),
            "Percent of time over the past sample period during which global (device) memory was being read or written.",
            []string{"index","name"},nil,)
    memory_total=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","memory_total"),
        "Total installed GPU memory.",
        []string{"index","name"},nil)
    memory_free=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","memory_free"),
        "",
        []string{"index","name"},nil)
    memory_used=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","memory_used"),
        "Total memory allocated by active contexts",
        []string{"index","name"},nil)
    temperature_gpu=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","temperature_gpu"),
        "Core GPU temperature. in degrees C.",
        []string{"index","name"},nil)
    power_draw=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","power_draw"),
        "The last measured power draw for the entire board, in watts. Only available if power management is supported. This reading is accurate to within +/- 5 watts.",
        []string{"index","name"},nil)
    power_limit=prometheus.NewDesc(
       prometheus.BuildFQName(namespace,"","power_limit"),
       "The software power limit in watts. Set by software like nvidia-smi. On Kepler devices Power Limit can be adjusted using [-pl | --power-limit=] switches.",
       []string{"index","name"},nil)

    video_encode=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","video_encode"),
        "Video Encode",
        []string{"index","name"},nil)
    video_decode=prometheus.NewDesc(
        prometheus.BuildFQName(namespace,"","video_decode"),
        "Video Decode",
        []string{"index","name"},nil)


)
type promHTTPLogger struct {
    logger log.Logger
}

func (l promHTTPLogger) Println(v ...interface{}) {
    level.Error(l.logger).Log("msg", fmt.Sprint(v...))
}

type Exporter struct {
    //   client        *consul_api.Client
    //   kvPrefix      string
    //   kvFilter      *regexp.Regexp
    healthSummary bool
    logger        log.Logger
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
    ch <- up
    ch <- GPU
    ch <- utilization_gpu
    ch <- utilization_memory
    ch <- memory_total
    ch <- temperature_gpu
    ch <- power_draw
    ch <- power_limit
    ch <- video_encode
    ch <- video_decode
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
    ok := e.collectMetrics(ch)
    if ok {
        ch<-prometheus.MustNewConstMetric(
            up,prometheus.GaugeValue,1.0)
    } else {
        ch<-prometheus.MustNewConstMetric(
            up,prometheus.GaugeValue,0.0)
    }
}

func (e *Exporter) collectMetrics(ch chan <-prometheus.Metric) bool {
    //# 第一种查询方式
    out, err := exec.Command(
        "nvidia-smi",
        "--query-gpu=index,name,utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,temperature.gpu,power.draw,power.limit",
        "--format=csv,noheader,nounits").Output()

    if err != nil {
        fmt.Printf("%s\n", err)
        return false
    }

    csvReader := csv.NewReader(bytes.NewReader(out))
    csvReader.TrimLeadingSpace = true
    records, _ := csvReader.ReadAll()

    for _,row :=range records  {
        if len(row) > 0 {
            // 第一种查询可以查到名称等信息，但是查不到视频编解码利用率
            // 这里再查一下就可以了,前两行为符号位，不需要(需优化)
            video,err:=exec.Command("nvidia-smi","dmon","-i",row[0],"-c","1").Output()
            if err==nil {
                csvReader := csv.NewReader(bytes.NewReader(video))
                csvReader.TrimLeadingSpace = true
                info, _ := csvReader.ReadAll()
                if len(info) > 2 {
                    for i:=2;i < len(info); i++ {
                        split:=strings.Fields(info[i][0])
                        //fmt.Print(split[1])
                        videoEncode,err:= strconv.ParseFloat(split[6],64) //GPU利用率
                        videoDecode,err:=strconv.ParseFloat(split[7],64) //GPU利用率
                        if err!=nil {
                            fmt.Print(err)
                        }
                        ch<-prometheus.MustNewConstMetric(
                            video_encode,prometheus.GaugeValue,videoEncode,row[0],row[1])
                        ch<-prometheus.MustNewConstMetric(
                            video_decode,prometheus.GaugeValue,videoDecode,row[0],row[1])
                    }
                }
            }
            utilGpu,err:= strconv.ParseFloat(row[2],64) //GPU利用率
            utilMemory,err:=strconv.ParseFloat(row[3],64) //显存利用率
            memTotal,err:=strconv.ParseFloat(row[4],64) //显存总大小
            tempGpu,err:=strconv.ParseFloat(row[7],64) //GPU 温度
            powDraw,err:=strconv.ParseFloat(row[8],64) // GPU功率
            powLimit,err:=strconv.ParseFloat(row[9],64) //GPU最大功率
            if err!=nil {
                fmt.Print(err)
            }
            ch<- prometheus.MustNewConstMetric(
                GPU,prometheus.GaugeValue,1.0,row[0],row[1],row[2],row[3],row[4],row[5],row[6],row[7],row[8],row[9])
            ch<-prometheus.MustNewConstMetric(
                utilization_gpu,prometheus.GaugeValue, utilGpu,row[0],row[1])
            ch<-prometheus.MustNewConstMetric(
                utilization_memory,prometheus.GaugeValue,utilMemory,row[0],row[1])
            ch<-prometheus.MustNewConstMetric(
                memory_total,prometheus.GaugeValue,memTotal,row[0],row[1])
            ch<-prometheus.MustNewConstMetric(
                 temperature_gpu,prometheus.GaugeValue,tempGpu,row[0],row[1])
            ch<-prometheus.MustNewConstMetric(
                power_draw,prometheus.GaugeValue,powDraw,row[0],row[1])
            ch<-prometheus.MustNewConstMetric(
                power_limit,prometheus.GaugeValue,powLimit,row[0],row[1])
        }
    }


    return true
}
// name, index, temperature.gpu, utilization.gpu,
// utilization.memory, memory.total, memory.free, memory.used

func NewExporter(healthSummary bool, logger log.Logger) (*Exporter, error) {
    _, err := exec.Command("nvidia-smi",
        "-q").Output()
    if err != nil {
        fmt.Printf("%s\n", err)
        return nil,err
    }
    return &Exporter{
        healthSummary: healthSummary,
        logger: logger,
    },nil
    //fmt.Fprintf(response, strings.Replace(result, ".", "_", -1))
}

func init() {
    prometheus.MustRegister(version.NewCollector("nvidia_smi_exporter"))
}

func main() {

    var (
        listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9101").String()
        metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
        healthSummary = kingpin.Flag("consul.health-summary", "Generate a health summary for each service instance. Needs n+1 queries to collect all information.").Default("true").Bool()
    )

    promlogConfig:=&promlog.Config{}
    flag.AddFlags(kingpin.CommandLine,promlogConfig)
    kingpin.Version(version.Print("Nvidia_exporter"))
    kingpin.HelpFlag.Short('h')
    kingpin.Parse()

    logger := promlog.New(promlogConfig)

    level.Info(logger).Log("msg","Starting Nvidia_exporter","version",version.Info())
    level.Info(logger).Log("build_context", version.BuildContext())

    exporter, err := NewExporter(*healthSummary, logger)
    if err!=nil {
        level.Error(logger).Log("msg","Error marshaling query options", "err", err)
        os.Exit(1)
    }
    prometheus.MustRegister(exporter)


    http.Handle(*metricsPath,
        promhttp.InstrumentMetricHandler(
            prometheus.DefaultRegisterer,
            promhttp.HandlerFor(
                prometheus.DefaultGatherer,
                promhttp.HandlerOpts{
                    ErrorLog: &promHTTPLogger{
                        logger: logger,
                    },
                },
            ),
        ),
    )



    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`<html>
             <head><title>NVIDIA_SMI_Exporter</title></head>
             <body>
             <h1>NVIDIA_SMI_Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             <h2>Options</h2>
             </dl>
             <h2>Build</h2>
             <pre>` + version.Info() + ` ` + version.BuildContext() + `</pre>
             </body>
             </html>`))
        // <pre>` + string(queryOptionsJson) + `</pre>
    })

    http.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "OK")
    })

    http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "OK")
    })

    //http.HandleFunc(*metricsPath, metrics)
    level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
    if err := http.ListenAndServe( *listenAddress, nil); err != nil {
        level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
        os.Exit(1)
    }
}
