# file-to-machineconfig
Tool to convert files to MachineConfig objects to be used in conjunction with the machine-config-operator in k8s/OpenShift

NOTE: This is using currently [Configuration specification 2.2](https://coreos.com/ignition/docs/latest/configuration-v2_2.html)

## Usage

```
filetomachineconfig --path <myfile> > ./myfile.yaml
```

Use `filetomachineconfig --help` for a more complete usage and flags.
