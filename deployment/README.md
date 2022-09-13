## Local Setup

#### Requirement
1. python3
2. npm (v7.21.0+)
3. solc (0.8.15)
4. docker, docker-compose(option)
5. helm(option)

#### Docker-Compose

Start...
```bash
blockNr=$(bash ./deployment/tool/tool.sh blockHeight)
bash ./deployment/tool/tool.sh prepare new
bash ./deployment/tool/tool.sh deployContracts
export L1_ENDPOINT=http://localhost:8545 && bash ./deployment/docker-compose/docker-compose.sh up $blockNr
```

Stop...
```bash
bash ./deployment/docker-compose/docker-compose.sh down
```

#### Helm (v3)

1. Prepare
```bash
export BLOCK_NUMBER=$(bash ./deployment/tool/tool.sh blockHeight)
bash ./deployment/tool/tool.sh prepare new
bash ./deployment/tool/tool.sh deployContracts
```

2. Install
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
## deploy redis
helm install redis bitnami/redis --namespace redis --create-namespace
## get redis password
export REDIS_PASSWORD=$(kubectl get secret --namespace redis redis -o jsonpath="{.data.redis-password}" | base64 -d)
## deploy postgresql
helm install postgresql bitnami/postgresql --namespace postgres --create-namespace \
    --set auth.username=postgres
## get postgresql password
export POSTGRES_PASSWORD=$(kubectl get secret --namespace postgres postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)

## initialize database
kubectl port-forward --namespace postgres svc/postgresql 5432:5432

./build/bin/zkbnb db initialize --dsn "host=localhost user=postgres password=${POSTGRES_PASSWORD} dbname=zkbnb port=5432 sslmode=disable" --contractAddr ./deployment/configs/contractaddr.yaml

## deploy application
export KEY_FILE_PATH=$(pwd)/deployment/.zkbnb
helm install zkbnb \
    -f ./deployment/helm/local-value/values.yaml \
    --post-renderer ./deployment/helm/local-value/post-render.sh \
    --namespace zkbnb --create-namespace \
    ./deployment/helm/zkbnb

```

3. Uninstall
```bash
helm uninstall zkbnb -n zkbnb
helm uninstall postgresql -n postgres
# kubectl delete pvc --all -n postgres
helm uninstall redis -n redis
# kubectl delete pvc --all -n redis
```

