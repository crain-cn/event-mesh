##configure alertmanager
kubectl create secret generic alertmanager-prometheus-operator-alertmanager  --from-file=alertmanager.yaml

##install sample app
kubectl create ns jituan-zhongtai-iaas
kubectl apply -f crd/eventmesh_eventroute.yaml
kubectl apply -f crd/notification_receiver.yaml
kubectl apply -f app/deployment.yaml 

##install  promethus rule
kubectl apply -f promethus/rules/pod.yaml
