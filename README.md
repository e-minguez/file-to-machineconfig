你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
# file-to-machineconfig

Simple tool to convert files to MachineConfig objects to be used with the machine-config-operator in Kubernetes/OpenShift.

> **NOTE**: It only supports [Ignition configuration specification 2.2](https://coreos.com/ignition/docs/latest/configuration-v2_2.html) so far.

## Features

- [x] Linux/OSX/Windows support (and binaries available in [releases/](https://github.com/e-minguez/file-to-machineconfig/releases))
- [x] remote path = local path if not provided
- [x] remote owner/group = local if not provided
- [x] sane defaults (...)
- [x] normalized parameters (...)
- [x] multiple labels support
- [x] base64 file encoded content support
- [x] json output
- [x] yaml output

## To Do

- [ ] Improve normalization and defaults
- [ ] Multiple ignition version
- [ ] Good code
- [ ] Better error handling

## Not working

- [ ] [plain file content support](https://github.com/e-minguez/file-to-machineconfig/issues/13)

## Usage

If using Fedora/CentOS, you can use [this copr](https://copr.fedorainfracloud.org/coprs/eminguez/eminguez-RPMs/),
otherwise download the latest binary release:

```shell
wget -L https://github.com/e-minguez/file-to-machineconfig/releases/download/0.0.3/file-to-machineconfig-linux-amd64 && \
  mv file-to-machineconfig-linux-amd64 ./file-to-machineconfig && \
  chmod a+x ./file-to-machineconfig
```

Or, get the code with:

```shell
go get -u -v github.com/e-minguez/file-to-machineconfig
```

> **NOTE**: This requires golang to be installed and GOPATH properly configured.

Then:

```shell
file-to-machineconfig --file /local/path/to/my/file.txt > ./my-machine-config.json
```

Use `file-to-machineconfig --help` for a more complete usage and flags.

> **WARNING**: Review the generated output first, **do not pipe the output directly to `kubectl` or `oc`**!!!

## Example

```shell
echo "vm.swappiness=10" > ./myswap.conf

file-to-machineconfig --file ./myswap.conf --remote /etc/sysctl.d/swappiness.conf -yaml > myswap.yaml
...[output]...
2019/11/04 14:55:40 name not provided, using '99-worker-etc-sysctl-d-swappiness-conf' as name
2019/11/04 14:55:40 labels not provided, using 'machineconfiguration.openshift.io/role: worker' by default
2019/11/04 14:55:40 filesystem not provided, using 'root' by default
2019/11/04 14:55:40 apiver not provided, using 'machineconfiguration.openshift.io/v1' by default
2019/11/04 14:55:40 user not provided, using 'edu' as the original file
2019/11/04 14:55:40 group not provided, using 'edu' as the original file
2019/11/04 14:55:40 mode not provided, using '0664' as the original file

cat ./myswap.yaml
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  creationTimestamp: null
  labels:
    machineconfiguration.openshift.io/role: worker
  name: 99-worker-etc-sysctl-d-swappiness-conf
spec:
  config:
    ignition:
      config: {}
      security:
        tls: {}
      timeouts: {}
      version: 2.2.0
    networkd: {}
    passwd: {}
    storage:
      files:
      - contents:
          source: data:text/plain;charset=utf-8;base64,dm0uc3dhcHBpbmVzcz0xMAo=
          verification: {}
        filesystem: root
        group:
          name: edu
        mode: 436
        path: /etc/sysctl.d/swappiness.conf
        user:
          name: edu
    systemd: {}
  osImageURL: ""
```

Just to verify:

```shell
echo "dm0uc3dhcHBpbmVzcz0xMAo=" | base64 -d
vm.swappiness=10
```
