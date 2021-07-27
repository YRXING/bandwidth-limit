#!/bin/bash

container=""
rates=""
veth=""
port=""
nic=""
ip=""

#get host veth corresponding to container
#Input: the container's ID
function getveth() {
  if [ -z "$1" ]; then
    echo "ERROR: getveth function --- a container ID is neccessary!"
    exit 1
  fi
  # Get the PID of a docker container
  container_pid=$(docker inspect --format '{{.State.Pid}}' "$1")
  # Export container's namespcae to make a docker container's networking info available to 'ip netns',so that we can access this container's namespace
  if [ ! -d /var/run/netns ];then
    mkdir -p /var/run/netns
  fi
  ln -sf "/proc/${container_pid}/ns/net" "/var/run/netns/ns-${container_pid}"
  
  # Get the index number of a docker container's first veth interface (typically eth0)
  ils=$(ip netns exec "ns-${container_pid}" ip link show type veth)
  container_host_ifindex="${ils#*@if}"
  container_host_ifindex="${container_host_ifindex%%:*}"
  # Get the host veth interface attached to a container.
  host_network=$(ip link show | grep "${container_host_ifindex}: ")
  postfix="${host_network##${container_host_ifindex}: }"
  veth="${postfix%%@if*}"
  echo "INFO: the veth of [ $1 ] on the host is [ $veth ]"
}

#get container id according to image id
#Input：image id
function getContainer(){
  if [ -z "$1" ]; then
    echo "ERROR: getContainer function --- a image id is neccessary!"
    exit 1
  fi
  echo "INFO: getContainer function --- arg: [ $1 ]"
  containerIDs=($(docker ps -qf "ancestor=$1"))
  echo "INFO: the running containers: [ ${containerIDs[@]} ]"
  container=${containerIDs[0]}
}

#get the local nic name
function get_nic(){
  local str=$(ip link show)
  str=${str#*2: }
  nic=${str%%:*}
  echo "INFO: the local nic name is ${nic}"
}

#redirect eth0 to ifb0 to limit ingress traffic
#Input: the NIC name (typically eth0),
function redirect_nic(){
  #open ifb0 network interface
  modprobe ifb numifbs=1
  ip link set ifb0 up
  #redirect eth0 to ifb0
  tc qdisc add dev $1 ingress handle ffff:
  tc filter add dev $1 parent ffff: protocol ip prio 0 u32 match u32 0 0 flowid ffff: action mirred egress redirect dev ifb0
}

#Input: veth,rates
function limit_veth(){
  echo "INFO: limit_veth function called --- arg: [ $1 ] [ $2 ] "
  echo "INFO: qdisc before is [ $(tc -s qdisc ls dev $1) ]"
  tc qdisc del dev $1 root 
  tc qdisc add dev $1 root tbf rate $2 latency 50ms burst 20k
  if (($? != 0)); then
    echo "ERROR: qdisc set failed!"
    exit 1:
  else
    echo "INFO: qisc set successfully!"
    echo "INFO: qdisc now is [ $(tc -s qdisc ls dev $1) ]"
  fi
}

#Input: rates,port
function limit_host(){
  echo "INFO: limit_host function called --- arg: [ $1 ] [ $2 ] "
  echo "INFO: qdisc before is [ $(tc -s qdisc ls dev ifb0) ]"
  #limit egress bandwidth of ifb0
  tc qdisc del dev ifb0 root
  tc qdisc add dev ifb0 root handle 1:0 htb default 1
  
  tc class add dev ifb0 parent 1:0 classid 1:1 htb rate ${2} burst 20k
  
  tc class add dev ifb0 parent 1:1 classid 1:10 htb rate ${2}
  tc qdisc add dev ifb0 parent 1:10 handle 10: sfq perturb 10
  
  tc filter add dev ifb0 parent 1:0 prio 1 u32 match ip dport ${3} 0xffff flowid 1:10
  echo "INFO: qisc set successfully!"
  echo "INFO: qdisc now is [ $(tc -s qdisc ls dev ifb0) ]"
}

#Input: rate,ip/marsk（for example: 10.10.10.10/24）
function limit_ip(){
  echo "INFO: limit_ip function called --- arg: [ $1 ] [ $2 ] "
  echo "INFO: qdisc before is [ $(tc -s qdisc ls dev ifb0) ]"
  tc qdisc del dev ifb0 root
  tc qdisc add dev ifb0 root handle 1: htb default 1
  
  tc class add dev ifb0 parent 1:0 classid 1:1 htb rate ${1} burst 20k
  
  tc class add dev ifb0 parent 1:1 classid 1:10 htb rate ${1} ceil ${1} burst 20k
  tc qdisc add dev ifb0 parent 1:10 handle 10: sfq perturb 10
  
  tc filter add dev ifb0 protocol ip parent 1:0 prio 1 u32 match ip src ${2} flowid 1:10
  echo "INFO: qisc set successfully!"
  echo "INFO: qdisc now is [ $(tc -s qdisc ls dev ifb0) ]"
}


function use(){
  echo "Useage: script [options] [argument]

  Options:
  -i       Image id
  -c       container id or container name
  -p       container port
  -r       The bandwidth rate you want to limit
  -d       del the qdisc exist
  -h       Print the use information
  
  do you want test effect?
  In addition to checking the correspongding qdisc rules, you can alse use 'iperf' command:
  you have to start a server: iperf3 -s -p $port
  client:                     docker exec -it $containerID sh -c 'iperf3 -c $ip -p $port -R'
  "
}

if [ $# -lt 2 ];then
  use
  exit 1
fi

while getopts ":i:c:p:r:hd:n:" opt; do
  case $opt in
    i)
      getContainer $OPTARG
      getveth ${container}
      ;;
    c)
      container=${OPTARG}
      getveth ${container}
      ;;
    p)
      port=${OPTARG};;
    n)
      ip=${OPTARG};;
    r)
      rates=$OPTARG;;
    h)
      use;;
    d)
      tc qdisc dev dev ${OPTARG} root
      echo "INFO: qdisc clear successfully!"
      ;;
    :)
      echo "Option -$OPTARG requires an argument"
      exit 1
      ;;
    ?)
      echo "Invalid option: -$OPTARG";;
  esac
done

echo "INFO: the args--${container} ${port} ${rates} ${ip}"

if [ ${ip} -o ${port} ];then
  get_nic
  redirect_nic ${nic}
fi

if [ ${ip} ];then
  echo "INFO: limit_ip"
  limit_ip ${rates} ${ip}
elif [ ${port} ];then
  echo "INFO: limit_host"
  limit_host ${rates} ${port}
else
  echo "INFO: limit_veth"
  limit_veth ${veth} ${rates}
fi