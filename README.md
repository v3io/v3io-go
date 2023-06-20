# Project: v3io-go

## Description:
**v3io-go** is an interface to V3IO API (data & control planes) for Golang

## Prerequisites:
1. Cap'n proto compiler
    * MacOS: `brew install capnp`
    * Debian / Ubuntu: `apt-get install capnproto`
    * Other: Follow the guidelines https://capnproto.org/install.html

## Dependencies:
1. Using go modules
    * `go mod download`

## Running the tests
1. **Define the following environment variables:**
    - V3IO_DATAPLANE_URL=http://<app-node>:8081
    - V3IO_DATAPLANE_USERNAME=<username>
    - V3IO_DATAPLANE_ACCESS_KEY=<access-key>
    - V3IO_CONTROLPLANE_URL=http://<data-node>:8001
    - V3IO_CONTROLPLANE_USERNAME=<username>
    - V3IO_CONTROLPLANE_PASSWORD=<password>
    - V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD=<igz_admin-password>
2. Run ``make test``
3. Alternatively you can pass environment variables inline as you can see in the following example: 
    ```
    V3IO_DATAPLANE_URL=http://<app-node>:8081 \
    V3IO_DATAPLANE_USERNAME=<username> \
    V3IO_DATAPLANE_ACCESS_KEY=<access-key> \
    V3IO_CONTROLPLANE_URL=http://<data-node>:8001 \
    V3IO_CONTROLPLANE_USERNAME=<admin-username> \
    V3IO_CONTROLPLANE_PASSWORD=<admin-password> \
    V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD=<igz_admin-password> \
    make test
    ```

## Mocking
We used [mockery](https://vektra.github.io/mockery/) to generate mocks for the interfaces in the `v3io` package.
To generate mock for interface in specific path you can run:
```bash
mockery --dir <path_to_dir_contains_interface> --name <interface_name>
```

For example, to generate mock for `Session` interface in `controlplane` you can run:
```bash
mockery --dir pkg/controlplane --name Session
```