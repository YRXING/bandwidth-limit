`Before building image, make sure that shell-operator has execute permission on hook file.`

## Run

build image:

```bash
docker build -t "yrxing/bandwidth-limit:1.0" .
docker push yrxing/bandwidth-limit:1.0
```

Apply manifests:

```bash
kubectl create ns bandwidth-limit
kubectl -n bandwidth-limit apply -f bandwidth-limit-rbac.yaml
kubectl -n bandwidth-limit applly -f bandwidth-limit-pod.yaml
```

or you can creat without yaml file:

```bash
kubectl create serviceaccount bandwidth-limit-acc --namespcace bandwidth-limit
kubectl create clusterrole bandwidth-limit --verb=get,watch,list --resource=pods
kubectl create clusterrolebinding bandwidth-limit --clusterrole=bandwidth-limit \
	--serviceaccount=bandwidth-limit:bandwidth-limit-acc
```

See in logs:

```bash
kubectl apply -f test-pod.yaml
kubectl logs -f bandwidth-limit  -n bandwidth-limit
```



## Cleanup

```bash
kubectl delete clusterrolebinding/bandwidth-limit
kubectl delete clusterrole/bandwidth-limit
kubectl delete ns/bandwidth-limit
docker rmi yrxing/bandwidth-limit:1.0
```

