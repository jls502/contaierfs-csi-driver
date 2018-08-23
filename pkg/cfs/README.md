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

```kubectl create -f deploy/examples/nginx-pre.yaml```

### Dynamic volume: Example Nginx application
```kubectl create -f deploy/examples/cfs-pvc.yaml```
```kubectl create -f deploy/examples/cfs-pv.yaml```
```kubectl create -f deploy/examples/nginx-dynamic.yaml```
