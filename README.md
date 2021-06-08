# k8s-test
A simple application for testing a k8s environment.

## Prepare

1. Install ingress

```shell
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install -f helm/ingress/values.yaml ingress-nginx ingress-nginx/ingress-nginx
```

2. Add repository `bitnamy` for postgres

```shell
helm repo add bitnami https://charts.bitnami.com/bitnami
```

## Run application

### Skaffold

```shell
skaffold run
```

## Test

Get a public IP address from an admin console. For example, it's 89.22.183.246.

Add a public IP to /etc/hosts
```shell
echo "89.22.183.246 cluster-test" >> /etc/hosts
```

Check endpoints

```shell
curl http://cluster-test/
root handler

curl http://cluster-test/user/count
{count of user: 10000}

curl http://cluster-test/user/random
{"Id":2837,"Name":"fake-2836","Emails":["fake-master-2836@email.com","fake-slave-2836@email.com"]}
```

Run performance test

```shell
ab -n 1000 -c 5 http://cluster-test/user/random
for i in `seq 1 20`; do echo $i; ab -n 1000 -c 5 http://cluster-test/user/random ; done
```
