CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o eventroute ../../cmd/eventroute/main.go
rm -rf config.yml && cp -r ../../config/config-test.yml config.yml
rm -rf default.tmpl && cp -r ../../config/templates/default.tmpl default.tmpl
docker build -t hub.xesv5.com/jituan-zhongtai-iaas/event-mesh:v1.0.1-test ./
docker push hub.xesv5.com/jituan-zhongtai-iaas/event-mesh:v1.0.1-test
