# file-to-machineconfig

Simple tool to convert files to MachineConfig objects to be used with the machine-config-operator in Kubernetes/OpenShift.

> **NOTE**: It only supports [Ignition configuration specification 2.2](https://coreos.com/ignition/docs/latest/configuration-v2_2.html) so far.

## Usage

Download the latest binary release:

```shell
wget -L https://github.com/e-minguez/file-to-machineconfig/releases/download/0.0.1/file-to-machineconfig-linux-amd64 && \
  mv file-to-machineconfig-linux-amd64 ./file-to-machineconfig && \
  chmod a+x ./file-to-machineconfig
```

Or, get the code with:

```shell
go get -u -v github.com/e-minguez/file-to-machineconfig
```

Then:

```shell
file-to-machineconfig --file /local/path/to/my/file.txt > ./my-machine-config.json
```

Use `file-to-machineconfig --help` for a more complete usage and flags.

> **WARNING**: Review the generated output first, **do not pipe the output directly to `kubectl` or `oc`**!!!

## Example

```shell
echo "vm.swappiness=10" > ./myswap.conf

file-to-machineconfig --file ./myswap.conf --remote /etc/sysctl.d/swappiness.conf > myswap.json
...[output]...
2019/10/24 15:16:10 name not provided, using '99-worker-etc-sysctl-d-swappiness-conf' as name
2019/10/24 15:16:10 mode not provided, using '0664' as the original file
2019/10/24 15:16:10 user not provided, using 'edu' as the original file
2019/10/24 15:16:10 group not provided, using 'edu' as the original file

cat myswap.json | jq .
{
  "kind": "MachineConfig",
  "apiVersion": "machineconfiguration.openshift.io/v1",
  "metadata": {
    "name": "99-worker-etc-sysctl-d-swappiness-conf",
    "creationTimestamp": null,
    "labels": {
      "machineconfiguration.openshift.io/role": "worker"
    }
  },
  "spec": {
    "osImageURL": "",
    "config": {
      "ignition": {
        "config": {},
        "security": {
          "tls": {}
        },
        "timeouts": {},
        "version": "2.2"
      },
      "networkd": {},
      "passwd": {},
      "storage": {
        "files": [
          {
            "filesystem": "root",
            "group": {
              "name": "edu"
            },
            "path": "/etc/sysctl.d/swappiness.conf",
            "user": {
              "name": "edu"
            },
            "contents": {
              "source": "data:text/plain;charset=utf-8;base64,dm0uc3dhcHBpbmVzcz0xMAo=",
              "verification": {}
            },
            "mode": 436
          }
        ]
      },
      "systemd": {}
    }
  }
}
```

Just to verify:

```shell
echo "dm0uc3dhcHBpbmVzcz0xMAo=" | base64 -d
vm.swappiness=10
```

## To do

* Improve the code (like A LOT)
* Error handling