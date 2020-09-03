# nvidia_smi_Exporter
[  描述 ]

根据Prometheus标准，创建NVIDIA_SMI 数据采集器，

[  条件 ] 

需要提前将英伟达驱动安装完毕，并将NVIDIA_SMI文件加入环境变量。

[  使用方式 ]

`  ./nvidia_smi_exporter.exe -h `

> usage: nvidia_smi_exporter.exe [<flags>]
>
> Flags:
>   -h, --help                   Show context-sensitive help (also try --help-long
>                                and --help-man).
>       --web.listen-address=":9101"
>                                Address to listen on for web interface and
>                                telemetry.
>       --web.telemetry-path="/metrics"
>                                Path under which to expose metrics.
>       --consul.health-summary  Generate a health summary for each service
>                                instance. Needs n+1 queries to collect all
>                                information.
>       --log.level=info         Only log messages with the given severity or
>                                above. One of: [debug, info, warn, error]
>       --log.format=logfmt      Output format of log messages. One of: [logfmt,
>                                json]
>       --version                Show application version.

[  参数标准  ]

>|             metric             | - index：显卡序号  - name：显卡型号                          | value              |
>| :----------------------------: | ------------------------------------------------------------ | ------------------ |
>|           nvidia_GPU           | {index="0",memory_free="14738",memory_total="16384",memory_used="1646",name="Quadro RTX 5000",power_draw="45.51",power_limit="230.00",temperature_gpu="44",utilization_gpu="15",utilization_memory="5"} |                    |
>|      nvidia_memory_total       | {index="0",name="Quadro RTX 5000"}                           | 显存容量（MB）     |
>|       nvidia_power_draw        | {index="0",name="Quadro RTX 5000"}                           | 功率（W）          |
>|       nvidia_power_limit       | {index="0",name="Quadro RTX 5000"}                           | 最大功率（W）      |
>| nvidia_smi_exporter_build_info | {branch="54cm06",goversion="go1.13.6",revision="0.0.1",version="0.1.0"} | build信息          |
>|     nvidia_temperature_gpu     | {index="0",name="Quadro RTX 5000"}                           | 显卡温度（摄氏度） |
>|     nvidia_utilization_gpu     | {index="0",name="Quadro RTX 5000"}                           | 显卡利用率（%）    |
>|   nvidia_utilization_memory    | {index="0",name="Quadro RTX 5000"}                           | 显存使用率（%）    |
>|      nvidia_video_decode       | {index="0",name="Quadro RTX 5000"}                           | 视频解码率（%）    |
>|      nvidia_video_encode       | {index="0",name="Quadro RTX 5000"}                           | 视频编码率（%）    |
>
>

