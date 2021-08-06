



<p align="center">
 <a href="https://github.com/YRXING/bandwidth-limit/discussions"><img src="https://img.shields.io/badge/GitHub-discussions-brightgreen" alt="GH Discussions"/></a>
</p>

## 

## Introduction

When an application occupy a log of network bandwidth, like edge nodes frequently pulling big image files on the cloud, it will hinder the operation of other network applications. At this time, our interactive SSH session may become extreamly slow and unuseable.

This tool is designed to shap the traffic of high bandwidth container application base on `linux netlink socket`.

## Useage

If you just want to limit the container at once, maybe the [shell-script](https://github.com/YRXING/bandwidth-limit/tree/main/shell-script) can meet your requirment. The script wraps up the `tc` command so that you can use it easily.

You can also use it to limit your pod in k8s.

### Running it

```go
go build -o limiter main.go

./limiter
```



### Defining container's traffic in pod's annotations

For example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: registry
  template:
    metadata:
      labels:
        app: registry
      annotations:
        k8s.harmonycloud.com/ingress-bandwidth: 10M
        k8s.harmonycloud.com/egress-bandwidth: 20M
    spec:
      nodeSelector:
        kubernetes.io/hostname: k8s-node-187
      tolerations:
      - key: "noSchedule"
        operator: "Exists"
        effect: "NoSchedule"
      containers:
      - image: nginx:1.21.1
        name: registry-container
```



The traffic unit is bytes and case insensitive. You can define it like "k,m,g,t".

#### Becareful: if you limit the traffic to a small value, you may can not log in to your contianer/host remotely.



### Test it

**Before**

the egress bandwidth is `2.89 Gbits/sec`

```bash
root@registry-deployment-f769d8875-x7fkm:/# iperf3 -c 10.10.102.190 -p 8081
Connecting to host 10.10.102.190, port 8081
[  5] local 10.100.121.215 port 41640 connected to 10.10.102.190 port 8081
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec   323 MBytes  2.71 Gbits/sec  170    427 KBytes       
[  5]   1.00-2.00   sec   300 MBytes  2.52 Gbits/sec  233    309 KBytes       
[  5]   2.00-3.00   sec   299 MBytes  2.51 Gbits/sec  156    352 KBytes       
[  5]   3.00-4.00   sec   309 MBytes  2.59 Gbits/sec  136    413 KBytes       
[  5]   4.00-5.00   sec   299 MBytes  2.51 Gbits/sec  157    306 KBytes       
[  5]   5.00-6.00   sec   313 MBytes  2.63 Gbits/sec   95    331 KBytes       
[  5]   6.00-7.00   sec   333 MBytes  2.80 Gbits/sec  168    335 KBytes       
[  5]   7.00-8.00   sec   220 MBytes  1.84 Gbits/sec  112    351 KBytes       
[  5]   8.00-9.00   sec   286 MBytes  2.40 Gbits/sec  126    388 KBytes       
[  5]   9.00-10.00  sec   281 MBytes  2.36 Gbits/sec  135    270 KBytes       
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec  2.89 GBytes  2.49 Gbits/sec  1488             sender
[  5]   0.00-10.00  sec  2.89 GBytes  2.48 Gbits/sec                  receiver
```



the ingress bandwidth is `3.82 Gbits/sec`

```
root@registry-deployment-f769d8875-x7fkm:/# iperf3 -c 10.10.102.190 -p 8081 -R
Connecting to host 10.10.102.190, port 8081
Reverse mode, remote host 10.10.102.190 is sending
[  5] local 10.100.121.215 port 42340 connected to 10.10.102.190 port 8081

[ ID] Interval           Transfer     Bitrate
[  5]   0.00-1.00   sec   320 MBytes  2.68 Gbits/sec                  
[  5]   1.00-2.00   sec   408 MBytes  3.43 Gbits/sec                  
[  5]   2.00-3.00   sec   370 MBytes  3.10 Gbits/sec                  
[  5]   3.00-4.00   sec   424 MBytes  3.56 Gbits/sec                  
[  5]   4.00-5.00   sec   401 MBytes  3.36 Gbits/sec                  
[  5]   5.00-6.00   sec   303 MBytes  2.54 Gbits/sec                  
[  5]   6.00-7.00   sec   423 MBytes  3.55 Gbits/sec                  
[  5]   7.00-8.00   sec   569 MBytes  4.77 Gbits/sec                  
[  5]   8.00-9.00   sec   301 MBytes  2.53 Gbits/sec                  
[  5]   9.00-10.00  sec   395 MBytes  3.31 Gbits/sec                  
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec  3.82 GBytes  3.29 Gbits/sec  2165             sender
[  5]   0.00-10.00  sec  3.82 GBytes  3.28 Gbits/sec                  receiver
```



**After**

```bash
[root@k8s-node-187 ~]# ./limiter 
I0806 17:16:23.911828   14620 pods_controller.go:96] Add event: container exists 
I0806 17:16:23.912179   14620 pods_controller.go:98] pod's contaier id is: 5d272abaf7e0e507f666ccb8b7540fe0c85e43c32f41fa49348a62b069a8c487
I0806 17:16:23.912665   14620 pods_controller.go:104] pod's container pid is: 1956
I0806 17:16:23.914207   14620 pods_controller.go:107] Tc set up rules: &{10M 20M 176 4 false 0xc000209c20}
I0806 17:16:23.914384   14620 pods_controller.go:109] start setting tc rules....
I0806 17:16:23.914653   14620 utils.go:33] the host veth name is calie68221e6903
I0806 17:16:23.914704   14620 utils.go:40] the rate translated is 10000000 bytes per second
I0806 17:16:23.915335   14620 utils.go:50] set qdisc on host veth 176: calie68221e6903 successfully
I0806 17:16:23.917189   14620 utils.go:79] the rate translated is 20000000 bytes per second
I0806 17:16:23.917419   14620 utils.go:89] set qdisc on container's veth 4: eth0 successfully
```



the egress bandwidth is `148 Mbits/sec`

```bash
root@registry-deployment-f769d8875-x7fkm:/# iperf3 -c 10.10.102.190 -p 8081
Connecting to host 10.10.102.190, port 8081
[  5] local 10.100.121.215 port 44230 connected to 10.10.102.190 port 8081
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec  17.0 MBytes   143 Mbits/sec    2    178 KBytes       
[  5]   1.00-2.00   sec  18.0 MBytes   151 Mbits/sec    0    214 KBytes       
[  5]   2.00-3.00   sec  17.5 MBytes   147 Mbits/sec    0    229 KBytes       
[  5]   3.00-4.00   sec  18.1 MBytes   152 Mbits/sec    0    244 KBytes       
[  5]   4.00-5.00   sec  17.4 MBytes   146 Mbits/sec    0    244 KBytes       
[  5]   5.00-6.00   sec  17.7 MBytes   149 Mbits/sec    0    255 KBytes       
[  5]   6.00-7.00   sec  18.2 MBytes   152 Mbits/sec    0    287 KBytes       
[  5]   7.00-8.00   sec  17.0 MBytes   143 Mbits/sec    0    304 KBytes       
[  5]   8.00-9.00   sec  18.0 MBytes   151 Mbits/sec    0    304 KBytes       
[  5]   9.00-10.00  sec  17.4 MBytes   146 Mbits/sec    0    304 KBytes       
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec   176 MBytes   148 Mbits/sec    2             sender
[  5]   0.00-10.00  sec   175 MBytes   147 Mbits/sec                  receiver
```



the ingress bandwidth is `76.7 Gbits/sec`

```bash
root@registry-deployment-f769d8875-x7fkm:/# iperf3 -c 10.10.102.190 -p 8081 -R
Connecting to host 10.10.102.190, port 8081
Reverse mode, remote host 10.10.102.190 is sending
[  5] local 10.100.121.215 port 44450 connected to 10.10.102.190 port 8081
[ ID] Interval           Transfer     Bitrate
[  5]   0.00-1.00   sec  9.16 MBytes  76.8 Mbits/sec                  
[  5]   1.00-2.00   sec  9.08 MBytes  76.2 Mbits/sec                  
[  5]   2.00-3.00   sec  9.09 MBytes  76.2 Mbits/sec                  
[  5]   3.00-4.00   sec  9.07 MBytes  76.1 Mbits/sec                  
[  5]   4.00-5.00   sec  9.08 MBytes  76.1 Mbits/sec                  
[  5]   5.00-6.00   sec  9.09 MBytes  76.3 Mbits/sec                  
[  5]   6.00-7.00   sec  9.10 MBytes  76.3 Mbits/sec                  
[  5]   7.00-8.00   sec  9.03 MBytes  75.7 Mbits/sec                  
[  5]   8.00-9.00   sec  8.98 MBytes  75.4 Mbits/sec                  
[  5]   9.00-10.00  sec  9.03 MBytes  75.8 Mbits/sec                  
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec  91.5 MBytes  76.7 Mbits/sec   16             sender
[  5]   0.00-10.00  sec  90.7 MBytes  76.1 Mbits/sec                  receiver
```



You are welcome to make new issues and pull requests.
