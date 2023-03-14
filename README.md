# mgmt-backend

mgmt-backend communicates with [Versatile Data Pipeline (VDP)](https://github.com/instill-ai/vdp) to manage the user resources.

## Local dev

On the local machine, clone `vdp` repository in your workspace, move to the repository folder, and launch all dependent microservices:
```
$ cd <your-workspace>
$ git clone https://github.com/instill-ai/vdp.git
$ cd vdp
$ make dev PROFILE=mgmt ITMODE=true
```

Clone `mgmt-backend` repository in your workspace and move to the repository folder:
```
$ cd <your-workspace>
$ git clone https://github.com/instill-ai/mgmt-backend.git
$ cd mgmt-backend
```

### Build the dev image

```bash
$ make build
```

### Run the dev container

```bash
$ make dev
```

Now, you have the Go project set up in the container, in which you can compile and run the binaries together with the integration test in each container shell.

### Run the server in the container

Enter the container
```bash
$ docker exec -it mgmt-backend-cloud /bin/bash
```

Create the database
```bash
$ go run ./cmd/migration
```

Initialise the database
```bash
$ go run ./cmd/init
```

Build binaries
```bash
$ go build -o /mgmt-backend/exec/mgmt-backend ./cmd/main
$ go build -o /mgmt-backend/exec/plugin ./cmd/plugin
```

Run the server
```bash
$ /mgmt-backend/exec/mgmt-backend
```

### Run the integration test

``` bash
$ docker exec -it mgmt-backend /bin/bash
$ make integration-test
```

### Stop the dev container

```bash
$ make stop
```

### CI/CD

The latest images will be published to Docker Hub [repository](https://hub.docker.com/r/instill/mgmt-backend) at release.

## License

See the [LICENSE](./LICENSE) file for licensing information.
