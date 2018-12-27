minikube start \
   --kubernetes-version=v1.10.1 \
   --cpus=4 \
   --memory=6144 \
    --extra-config=kubelet.authentication-token-webhook=true
