# Overview
HLoad is an HTTP benchmark tool similar to hey or wrk.

The main features of HLoad
- Basic HTTP benchmark features you would expect in any HTTP benchmarking tool.
  - Concurrent number of connections.
  - Benchmark duration.
  - Abort the benchmark after n number of failed requests.
  - Rate limit the requests per second.
  - TLS.
- Running HTTP benchmark using the configuration file. We can save the configuration file and use it later to run the benchmark.
- Configuration file to control default parameter values. We can for example define different default connection counts.
- We can set multiple HTTP parameters, and they will be chosen randomly for every request.
- We can switch between HTTP libraries to make HTTP calls. You can choose net/http or fasthttp.
- We can store benchmark configuration and results in the local SQLite database.
- Usage of the T-Digest data structure to calculate request latency percentiles. In other HTTP benchmark tools, it usually works by storing all of the request latency information in the memory and, at the end, getting percentile from it. This approach makes it very resource-hungry when we want to make long-term benchmarks. Using the T-Digest method, we don't have to worry about memory usage when performing a long-running benchmark.
- Collect aggregated stats results in the provided time windows. For example, we can gather information like average request time, max request time, min request time, and requests count into a 10s window.
- Collect all requests stats from every request. THIS CAN CAUSE LARGE MEMORY USAGE IN LONG-RUNNING BENCHMARKS.
- Find saved benchmark configuration and results by name, description and time range.

# Examples

## Simple benchmark
Execute simple benchmark using all default values
This can be done by `loader run` subcommand, the only mandatory parameter is the URL.
<details>
<summary>Example</summary>

```bash
$ hload loader run --host http://192.168.50.147:8080
* Running loader...
* Loader basic parameters:
  * Created at: 2023-11-23 00:08:00.00641 +0100 CET m=+0.017385459
  * Target host: http://192.168.50.147:8080
  * Concurrent connections: 10
  * Requests count: 1000


99.90% [#################.] [1.00K in 1.207s] ... Requests progress

Summary
* Basic
  * URL:      http://192.168.50.147:8080
  * Start:    2023-11-23 00:08:00 +0100 CET
  * End:      2023-11-23 00:08:01 +0100 CET
  * Duration: 1.213677s
* Requests count
  * Total requests count: 1000
  * Success requests:     1000
 * Failed requests       0
  * Data transferred      600.6K
  * Request per second    823.942449/sec
* Requests latency
  * Average time 11.886475ms
  * Min time     5.262417ms
  * Max time     55.007083ms
  * P50 time     10.198735ms
  * P75 time     11.697522ms
  * P90 time     14.170362ms
  * P99 time     37.257256ms
* HTTP Codes
  * HTTP Code 200: 1000
```
</details>

## Simple benchmark, save the results in the database
<details>
<summary>Example</summary>

```bash
$ hload loader run --host http://192.168.50.147:8080 --save
* Running loader...
* Loader basic parameters:
  * Created at: 2023-11-23 00:08:57.280027 +0100 CET m=+0.026288501
  * Target host: http://192.168.50.147:8080
  * Concurrent connections: 10
  * Requests count: 1000


Requests progress ... done! [1.00K in 1.347s]

Summary
* Basic
  * UUID:     c5227df2-d050-4e19-8279-6b9d01ab80b2
  * URL:      http://192.168.50.147:8080
  * Start:    2023-11-23 00:08:57 +0100 CET
  * End:      2023-11-23 00:08:58 +0100 CET
  * Duration: 1.627858s
* Requests count
  * Total requests count: 1000
  * Success requests:     1000
  * Failed requests       0
  * Data transferred      600.6K
  * Request per second    614.304196/sec
* Requests latency
  * Average time 13.305065ms
  * Min time     5.346334ms
  * Max time     85.775ms
  * P50 time     10.320336ms
  * P75 time     14.210937ms
  * P90 time     21.658041ms
  * P99 time     31.65373ms
* HTTP Codes
  * HTTP Code 200: 1000

New loader configuration saved: 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac

$ hload loader find -u 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac -s
* Loader 1:
  * Loader UUID: 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac
  * Target host: http://192.168.50.147:8080
  * Created at: 2023-11-23 00:08:57 +0100 CET
  * Name: Configuration Thu, 23 Nov 2023 00:08:57.270
  * HTTP Engine: fast_http
  * Description: Default loader description
  * Concurrent connection: 10
  * Request count: 1000
  * Summaries:
    * Summary 1
      * Basic
        * UUID:     c5227df2-d050-4e19-8279-6b9d01ab80b2
        * URL:      http://192.168.50.147:8080
        * Start:    2023-11-23 00:08:57 +0100 CET
        * End:      2023-11-23 00:08:58 +0100 CET
        * Duration: 1.627858s
      * Requests count
        * Total requests count: 1000
        * Success requests:     1000
        * Failed requests       0
        * Data transferred      600.6K
        * Request per second    614.304196/sec
      * Requests latency
        * Average time 13.305065ms
        * Min time     5.346334ms
        * Max time     85.775ms
        * P50 time     10.320336ms
        * P75 time     14.210937ms
        * P90 time     21.658041ms
        * P99 time     31.65373ms
      * HTTP Codes
        * HTTP Code 200: 1000
```
</details>

## Make a benchmark with 20000000 requests but set the benchamrk timeout to 1s. After 1s it will stop and return the results
<details>
<summary>Example</summary>

```bash
$ hload loader run --host http://192.168.50.147:8080 -c 1 -r 20000000 --benchmark-timeout 1s
* Running loader...
* Loader basic parameters:
  * Created at: 2023-11-28 23:56:25.134087 +0100 CET m=+0.004248167
  * Target host: http://192.168.50.147:8080
  * Concurrent connections: 1
  * Requests count: 20000000


 0.00% [..................] [110 in 1.04698s] ... Requests progress

Summary
* Basic
  * URL:      http://192.168.50.147:8080
  * Start:    2023-11-28 23:56:25 +0100 CET
  * End:      2023-11-28 23:56:26 +0100 CET
  * Duration: 1.181138s
* Requests count
  * Total requests count: 110
  * Success requests:     110
  * Failed requests       0
  * Data transferred      66.1K
  * Request per second    93.130523/sec
* Requests latency
  * Average time 9.493764ms
  * Min time     5.145916ms
  * Max time     32.013ms
  * P50 time     8.546312ms
  * P75 time     10.230958ms
  * P90 time     12.90762ms
  * P99 time     23.091838ms
* HTTP Codes
  * HTTP Code 200: 110
```
</details>

## Find latest loader configuration and run it again.
<details>
<summary>Example</summary>

```bash
$ hload loader find -l 1
* Loader 1:
  * Loader UUID: 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac
  * Target host: http://192.168.50.147:8080
  * Created at: 2023-11-23 00:08:57 +0100 CET
  * Name: Configuration Thu, 23 Nov 2023 00:08:57.270
  * HTTP Engine: fast_http
  * Description: Default loader description
  * Concurrent connection: 10
  * Request count: 1000
  * Summaries:
$ hload loader start -u 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac
* Running loader...
* Loader basic parameters:
  * Created at: 2023-11-23 00:22:37.820439 +0100 CET m=+0.003357834
  * Target host: http://192.168.50.147:8080
  * Concurrent connections: 10
  * Requests count: 1000


99.90% [#################.] [999 in 1.480152s] ... Requests progress

Summary
* Basic
  * UUID:     84a7f472-6fb0-4c38-bcc8-8371d8bc58f2
  * URL:      http://192.168.50.147:8080
  * Start:    2023-11-23 00:22:37 +0100 CET
  * End:      2023-11-23 00:22:39 +0100 CET
  * Duration: 2.300805s
* Requests count
  * Total requests count: 1000
  * Success requests:     1000
  * Failed requests       0
  * Data transferred      600.6K
  * Request per second    434.630488/sec
* Requests latency
  * Average time 14.557348ms
  * Min time     6.023125ms
  * Max time     49.033958ms
  * P50 time     10.972124ms
  * P75 time     16.402839ms
  * P90 time     26.535145ms
  * P99 time     38.869669ms
* HTTP Codes
  * HTTP Code 200: 1000

New loader configuration saved: 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac
```
</details>

## Find the loader configuration created within 1hour
We can use  --from and --to options to specify the time range. 
It accept the `time.ParseDuration` format and also the following time formats
- DD.MM.YYYY
- DD.MM.YY
- MM/DD/YYYY
- MM/DD/YY
- MMDDYYYY
- MMDDYY
- HH:MM_YYYYMMDD
- YYYYMMDD
- MM/DD/YY
 
<details>
<summary>Example</summary>

```bash
./hload loader find --from=1h
* Loader 1:
  * Loader UUID: 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac
  * Target host: http://192.168.50.147:8080
  * Created at: 2023-11-23 00:08:57 +0100 CET
  * Name: Configuration Thu, 23 Nov 2023 00:08:57.270
  * HTTP Engine: fast_http
  * Description: Default loader description
  * Concurrent connection: 10
  * Request count: 1000
  * Summaries:
```
</details>

## Find the loader configuration summaries created within 1hour
If we provide `-u` in `loader find` command, the `--from` and `--to` options will be applied to summaries.

<details>
<summary>Example</summary>

```bash
$ hload loader find -u 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac -s --from=1h
* Loader 1:
  * Loader UUID: 7bd14cba-d27e-4e4f-9a21-d9e8bb9a58ac
  * Target host: http://192.168.50.147:8080
  * Created at: 2023-11-23 00:08:57 +0100 CET
  * Name: Configuration Thu, 23 Nov 2023 00:08:57.270
  * HTTP Engine: fast_http
  * Description: Default loader description
  * Concurrent connection: 10
  * Request count: 1000
  * Summaries:
    * Summary 1
      * Basic
        * UUID:     84a7f472-6fb0-4c38-bcc8-8371d8bc58f2
        * URL:      http://192.168.50.147:8080
        * Start:    2023-11-23 00:22:37 +0100 CET
        * End:      2023-11-23 00:22:39 +0100 CET
        * Duration: 2.300805s
      * Requests count
        * Total requests count: 1000
        * Success requests:     1000
        * Failed requests       0
        * Data transferred      600.6K
        * Request per second    434.630488/sec
      * Requests latency
        * Average time 14.557348ms
        * Min time     6.023125ms
        * Max time     49.033958ms
        * P50 time     10.972124ms
        * P75 time     16.402839ms
        * P90 time     26.535145ms
        * P99 time     38.869669ms
      * HTTP Codes
        * HTTP Code 200: 1000
    * Summary 2
      * Basic
        * UUID:     c5227df2-d050-4e19-8279-6b9d01ab80b2
        * URL:      http://192.168.50.147:8080
        * Start:    2023-11-23 00:08:57 +0100 CET
        * End:      2023-11-23 00:08:58 +0100 CET
        * Duration: 1.627858s
      * Requests count
        * Total requests count: 1000
        * Success requests:     1000
        * Failed requests       0
        * Data transferred      600.6K
        * Request per second    614.304196/sec
      * Requests latency
        * Average time 13.305065ms
        * Min time     5.346334ms
        * Max time     85.775ms
        * P50 time     10.320336ms
        * P75 time     14.210937ms
        * P90 time     21.658041ms
        * P99 time     31.65373ms
      * HTTP Codes
        * HTTP Code 200: 1000
```
</details>
