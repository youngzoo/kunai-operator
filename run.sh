docker run -t -i --privileged \
 -v /etc/machine-id:/etc/machine-id:ro \
 -v /proc:/proc \
 -v /var/run/containerd/containerd.sock:/var/run/containerd/containerd.sock:ro \
 -v /var/run/podman/podman.sock:/var/run/podman/podman.sock:ro \
 -v ./config.yaml:/etc/kunai/config.yaml \
 kunai:main run --config /etc/kunai/config.yaml


curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
