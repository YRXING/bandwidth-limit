#!/bin/bash

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
	container_ifindex="${ils%%:*}"
	
	# Get the host veth interface attached to a container.
	host_network=$(ip link show)
	prefix="${hostnetwork%%@if${container_ifindex}:*}"
	veth="${prefix##*: }"
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
  local images
  images=($(docker images | grep $1 | awk '{print $1}'))
  echo "INFO: the name from image ID is [ ${images[0]} ]"
  containerIDs=($(docker ps | grep ${images[0]} | awk '{print $1}'))
  echo "INFO: the running containers: [ ${containerIDs[@]} ]"
  #images=($(docker images | grep $1 | awk '{if ($2=="latest") print $1;else print $1":"$2}'))
  #declare -a containerId
  #declare -i i=0
  #for name in "${images[@]}";do
  #        containerId[$i]=$(docker ps | grep $name | awk '{print $1}')
  #        ((i++))
  #done
}


#need parameters: veth rates
function limit(){
	if [ $@ -le 0 || $@ -gt 2 ]; then
		echo "ERROR: limit function --- wrong use of parameters."
		exit 1
	else
		echo "INFO: limit function --- arg: [ $1 ] [ $2 ] "
		echo "INFO: qdisc before is [ $(tc -s qdisc ls dev $1) ]"
		tc qdisc del dev $1 root 
		tc qdisc add dev $1 root tbf rate $2 latency 50ms burst 20k
		if (($? != 0)); then
			echo "ERROR: qdisc set failed!"
			exit 1
		else
			echo "INFO: qisc set successfully!"
			echo "INFO: qdisc now is [ $(tc -s qdisc ls dev $1) ]"
		fi
	fi
}

function use(){
  echo "Useage: script [options] [argument]

  Options:
  -n       Image id
  -d       The bandwidth rate you want to limit
  -h       Print the use information
  
  do you want test effect?
  In addition to checking the correspongding qdisc ruls, you can alse use 'iperf' command:
  you have to start a server: iperf3 -s -p $port
  client: 										docker exec -it $containerID sh -c 'iperf3 -c $ip -p $port -R'
  "
}

if [ $# -lt 2 ];then
	use
	exit 1
fi

while getopts ":d:u:n:hc" opt; do
	case $opt in
		n)
		  getContainer $OPTARG
			getveth ${containerIDs[0]}
      ;;
		d)
			limit $veth $OPTARG
      ;;
    u)
    	rate=$OPTARG:
    	docker exec -it $container sh -c 'tc qdisc del dev eth0 root'
    	docker exec -it $container sh -c 'tc qdisc add dev eth0 root $rate latency 50ms burst 20k'
    	echo "INFO: set upload qdisc successfully!"
    	;;
    h)
    	use
    	;;
    c)
			tc qdisc dev dev $veth root
			docker exec -it $container sh -c 'tc qdisc del dev eth0 root'
			echo "INFO: qdisc clear successfully!"
			;;
    :)
    	echo "Option -$OPTARG requires an argument"
    	exit 1
    	;;
    ?)
    	echo "Invalid option: -$OPTARG requires an argument"
    	;;
  esac
done

