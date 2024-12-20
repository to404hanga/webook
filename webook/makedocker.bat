del webook.sh || true
go mod tidy
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -tags=k8s -o webook.sh .
docker rmi -f to404hanga/webook:v0.0.1
docker build -t to404hanga/webook:v0.0.1 .