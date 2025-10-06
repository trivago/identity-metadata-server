# identity-metadata-server

This project holds two components, the `identity-server`, used to give machines
running on-premises identities, and the `metadata-server` used to implement
OIDC, aka. "Workload Identity Federation" for workloads running on Kubernetes
or on-premises servers.

## Maintenance and PRs

This repository is in active development but restricted to the cloud-stack
we run at trivago. We cannot maintain any code that authenticates to other
cloud providers but Google Cloud, as there is no way for us to test these
codepaths.  
If you wish to extend the functionality to other cloud providers, please
fork this repository.

PRs are welcome, but will take some time to be reviewed.

## Documentation

For detailed documentation on the two components hosted in the repository,
please have a look at the [docs](./docs) directory.

## License

All files in the repository are subject to the [Apache 2.0 License](LICENSE)

## Builds and Releases

All commits to the main branch need to use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/).  
Releases will be generated automatically from these commits using [Release Please](https://github.com/googleapis/release-please).

### Required tools

All [required tools](flake.nix) can be installed locally via [nix](https://nixos.org/)
and are loaded on demand via [direnv](https://direnv.net/).  
On MacOS you can install nix via the installer from [determinate systems](https://determinate.systems/).

```shell
curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
```

We provided a [justfile](https://github.com/casey/just) to generate the required `.envrc` file.
Run `just init-nix` to get started, or run the [script](hack/init-nix.sh) directly.

### Running unit-tests

After you have set up your environment, run unittests via `just test` or

```shell
go test ./...
```
