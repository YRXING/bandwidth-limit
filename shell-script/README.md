# bandwidth-limit

What it can do：

- limit the bandwidth of the corresponding running container according to the image id ( if there are multiple containers, only the first one is limited)
- limit bandwidth directly according to contaienr name or container ID
- It can automatically select the corresponding bandwidth limiting policy according to the network mode of running container
- limit the traffic of specifies IP
- clear the qdisc rules manually by specifying the network interface card (automatic mode under development)

Useage：

```bash
./limit.sh -c [container_name/container_id] -r [rates]
```

```bash
./limit.sh -i [image-id] -r [rates]
```

```bash
./limit.sh -I 10.10.102.190/24 -r [rates]
```

```bash
./limit.sh -d ifb0
```


