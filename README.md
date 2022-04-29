# mgmt-backend

mgmt-backend manages the user resources in [Visual Data Preparation](https://github.com/instill-ai/vdp) project.

## Development

Pre-requirements:

- Go v1.17 or later installed on your development machine

### Binary build

```bash
$ make
```

### Docker build

```bash
# Build images with BuildKit
$ DOCKER_BUILDKIT=1 docker build -t instill/mgmt-backend:dev .
```

The latest image will be published to Docker Hub [repository](https://hub.docker.com/r/instill/mgmt-backend) at release time.

### License

See the [LICENSE](./LICENSE) file for licensing information.
