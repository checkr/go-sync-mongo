# Go-Sync-Mongo Helm Chart
* Deploy MongoDB sync program
* Tested on Kubernetes 1.7
## Configuration

| Parameter                  | Description                         | Default                                                 |
|----------------------------|-------------------------------------|---------------------------------------------------------|
| `replicas`                     | Number of replicas | `1` |
| `image.repository`             | Image repository | `thinkpoet/go-sync-mongo` |
| `image.tag`                    | Image tag. Possible values listed [here](https://hub.docker.com/r/thinkpoet/go-sync-mongo/tags/).| `v1.1`|
| `image.pullPolicy`             | Image pull policy | `Always` |
| `resources`                    | CPU/Memory resource requests/limits | `{}` |
| `nodeSelector`                 | Node labels for pod assignment | `{}` |
| `tolerations`                  | Toleration labels for pod assignment | `[]` |
| `affinity`                     | Affinity settings for pod assignment | `{}` |
| `annotations`                  | Deployment annotations | `{}` |
| `podAnnotations`               | Pod annotations | `{}` |
| `commandline.args.src`         | Endpoint for source database | `127.0.0.1:27017` |
| `commandline.args.srcUsername` | Username for source database | `admin` |
| `commandline.args.srcPassword` | Password for source database | `admin` |
| `commandline.args.srcEnableSSL`| Enable SSL in source database | `false` |
| `commandline.args.dst`         | Endpoint for destination database | `127.0.0.1:27017` |
| `commandline.args.dstUsername` | Username for destination database | `admin` |
| `commandline.args.dstPassword` | Password for destination database | `admin` |
| `commandline.args.dstEnableSSL`| Enable SSL in destination database | `false` |
| `commandline.args.since`       | Seconds since the Unix epoch | `1` |
