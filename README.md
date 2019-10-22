# file-to-machineconfig
Simple tool to convert files to MachineConfig objects to be used with the machine-config-operator in k8s/OpenShift.

NOTE: This is using currently [Configuration specification 2.2](https://coreos.com/ignition/docs/latest/configuration-v2_2.html)

## Usage

```
file-to-machineconfig --file /local/path/to/my/file.txt --filepath /path/to/remote/file.txt > ./my-machine-config.json
```

Use `file-to-machineconfig --help` for a more complete usage and flags.

## Example

```
$ echo "vm.swappiness=10" > ./myswap.conf

$ file-to-machineconfig --file ./myswap.conf --filepath /etc/sysctl.d/swappiness.conf | jq .
name not provided, using 99-worker-etc-sysctl-d-swappiness-conf as name
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
            "path": "/etc/sysctl.d/swappiness.conf",
            "contents": {
              "source": "data:text/plain;charset=utf-8;base64,dm0uc3dhcHBpbmVzcz0xMAo=",
              "verification": {}
            },
            "mode": 420
          }
        ]
      },
      "systemd": {}
    }
  }
}
```

Just to verify:

```
$ echo "dm0uc3dhcHBpbmVzcz0xMAo=" | base64 -d
vm.swappiness=10
```

# To do
* Verify parameters
* Improve the code (like A LOT)
* Error handling