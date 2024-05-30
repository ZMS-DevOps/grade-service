# zms-devops-grade


Build and push to DockerHub
```shell
docker build -t devopszms2024/zms-devops-grade-service:latest .
docker push devopszms2024/zms-devops-grade-service:latest
```

Create grade-service infrastructure

```shell
minikube addons enable ingress
istioctl install --set profile=demo -y
```
First time you can use apply
```shell
kubectl apply -R -f grade-k8s 
kubectl apply -R -f grade-istio
```
When you want to replace existing pod, svc... you should use this command
```shell
kubectl replace --force -f grade-k8s
kubectl replace --force -f grade-istio
```

```shell
kubectl get pods -n backend
kubectl describe pods POD -n backend
```