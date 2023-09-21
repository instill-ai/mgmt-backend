# mgmt-backend

[![Integration Test](https://github.com/instill-ai/mgmt-backend/actions/workflows/integration-test.yml/badge.svg)](https://github.com/instill-ai/mgmt-backend/actions/workflows/integration-test.yml)

mgmt-backend manages user-related resources.

## Local dev

On the local machine, clone `base` repository in your workspace, move to the repository folder, and launch all dependent microservices:
```
$ cd <your-workspace>
$ git clone https://github.com/instill-ai/base.git
$ cd base
$ make latest PROFILE=mgmt
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
$ docker exec -it mgmt-backend /bin/bash
```

Create the database
```bash
$ go run ./cmd/migration
```

Initialise the database
```bash
$ go run ./cmd/init
```

Run the server
```bash
$ go run ./cmd/main
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

- **pull_request** to the `main` branch will trigger the **`Integration Test`** workflow running the integration test using the image built on the PR head branch.
- **push** to the `main` branch will trigger
  - the **`Integration Test`** workflow building and pushing the `:latest` image on the `main` branch, following by running the integration test, and
  - the **`Release Please`** workflow, which will create and update a PR with respect to the up-to-date `main` branch using [release-please-action](https://github.com/google-github-actions/release-please-action).

Once the release PR is merged to the `main` branch, the [release-please-action](https://github.com/google-github-actions/release-please-action) will tag and release a version correspondingly.

The images are pushed to Docker Hub [repository](https://hub.docker.com/r/instill/mgmt-backend).

## Contributing

Please refer to the [Contributing Guidelines](./.github/CONTRIBUTING.md) for more details.

## Community support

Please refer to the [community](https://github.com/instill-ai/community) repository.

## License

See the [LICENSE](./LICENSE) file for licensing information.
