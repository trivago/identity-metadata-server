set shell := ["/usr/bin/env", "bash", "-euo", "pipefail", "-c"]
set script-interpreter := ["/usr/bin/env", "bash", "-euo", "pipefail"]
set unstable

registry := "ghcr.io/trivago"
image_platform := "linux/arm64,linux/amd64"
buildx_command := "docker buildx build --platform=" + image_platform

# ------------------------------------------------------------------------------

_default:
  @just -l

# Return the current version of a command, optionally with a test index for <version>-test<testIdx>.
[group('misc')]
version cmd testIdx="":
  #!/usr/bin/env bash
  version=$(jq -r '."cmd/{{cmd}}"' .github/release-please-manifest.json)
  if [[ -n "{{ testIdx }}" ]]; then
    echo "${version}-test{{ testIdx }}"
  else
    echo "${version}"
  fi

# push a given target to the registry
[group('containers')]
push target version="local":
  @docker push {{registry}}/{{target}}:{{version}}

# push all targets to the registry
[group('containers')]
push-all version="local":
  @just push metadata-iptables-init {{version}}
  @just push metadata-server {{version}}
  @just push identity-server {{version}}
  @just push debug {{version}}

# build an image for a given target (debug, identity-server, metadata-iptables-init or metadata-server)
[group('containers')]
[script]
build target version:
  case {{target}} in
    debug|metadata-iptables-init)
      docker build \
        -t {{registry}}/{{target}}:{{version}} \
        -f docker/{{target}}.Dockerfile .
      ;;
    metadata-server|identity-server)
      just build-binaries {{target}}
      docker build \
        --build-arg="CMD={{target}}" \
        -t {{registry}}/{{target}}:{{version}} \
        -f docker/cmd.Dockerfile .
      ;;
    *)
      echo "Error: Invalid target '{{target}}'. Must be one of: debug, metadata-iptables-init, metadata-server, identity-server" >&2
      exit 1
      ;;
  esac

# build container images for all tools for a given version
[group('containers')]
build-all version:
  @just build debug {{version}}
  @just build identity-server {{version}}
  @just build metadata-iptables-init {{version}}
  @just build metadata-server {{version}}

[script]
build-binary target arch:
  export CGO_ENABLED=0
  export GOOS=linux
  echo "Building linux/{{arch}} binary"
  export GOARCH={{arch}}
  go build \
    -o ./bin/{{target}}-linux-{{arch}} \
    -tags=jsoniter \
    -ldflags="-s -w" \
    ./cmd/{{target}}

# build binaries for the target (identity-server or metadata-server)
[script]
build-binaries target:
  if [[ ! {{target}} =~ ^(metadata-server|identity-server)$ ]]; then
    echo "Error: Cannot build binaries for '{{target}}'. Must be one of: metadata-server, identity-server" >&2
    exit 1
  fi
  just build-binary {{target}} arm64
  just build-binary {{target}} amd64

# build a multiarch image for a given target (metadata-iptables-init or debug)
[group('ci')]
[script]
build-multiarch target version *args:
  case {{target}} in
    debug|metadata-iptables-init)
      {{buildx_command}} {{args}} \
        --tag {{registry}}/{{target}}:{{version}} \
        --file docker/{{target}}.Dockerfile .
      ;;
    metadata-server|identity-server)
      just build-binaries {{target}}
      {{buildx_command}} {{args}} \
        --build-arg="CMD={{target}}" \
        --tag {{registry}}/{{target}}:{{version}} \
        --file docker/cmd.Dockerfile .
      ;;
    *)
      echo "Error: Invalid target '{{target}}'. Must be one of: debug, metadata-iptables-init, metadata-server, identity-server" >&2
      exit 1
      ;;
  esac

# build a multiarch images for all targets (metadata-iptables-init, debug, metadata-server, identity-server)
[group('ci')]
build-multiarch-all version *args:
  @just build-multiarch debug {{version}} {{args}}
  @just build-multiarch identity-server {{version}} {{args}}
  @just build-multiarch metadata-iptables-init {{version}} {{args}}
  @just build-multiarch metadata-server {{version}} {{args}}

# run unittests for a given target on your local machine
[group('test')]
test target="all":
  @just test-{{target}}

[group('test')]
test-all:
  @just test-metadata-server

# run unittests for the metadata server
[group('test')]
test-metadata-server:
  @go test -v ./cmd/metadata-server/... -coverprofile=metadata-server.coverage

# setup nix with direnv
[group('misc')]
init-nix:
    #!/usr/bin/env bash
    set -euo pipefail

    cat <<-EOF > .envrc
    if ! has nix_direnv_version || ! nix_direnv_version 3.0.4; then
        source_url "https://raw.githubusercontent.com/nix-community/nix-direnv/3.0.4/direnvrc" "sha256-DzlYZ33mWF/Gs8DDeyjr8mnVmQGx7ASYqA5WlxwvBG4="
    fi
    use flake
    EOF

    direnv allow
