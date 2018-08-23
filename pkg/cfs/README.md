# CSI cfs driver


## Kubernetes
### Requirements

The folllowing feature gates and runtime config have to be enabled to deploy the driver

```
FEATURE_GATES=CSIPersistentVolume=true,MountPropagation=true
RUNTIME_CONFIG="storage.k8s.io/v1alpha1=true"
```

Mountprogpation requries support for privileged containers. So, make sure privileged containers are enabled in the cluster.

### Deploy

```kubectl -f deploy/kubernetes create```

### Pre Volume: Example Nginx application
Please update the cfs VolMgrServer & volume uuid information in nginx.yaml file.

```kubectl create -f examples/kubernetes/nginx-pre.yaml```

### Dynamic volume: Example Nginx application
```kubectl create -f examples/kubernetes/cfs-pvc.yaml```
```kubectl create -f examples/kubernetes/cfs-pv.yaml```
```kubectl create -f examples/kubernetes/nginx-dynamic.yaml```
