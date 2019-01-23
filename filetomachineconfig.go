package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
  "strings"
	//machineconfig "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

var (
	file        string
	format      string
	mode        int
	signature   bool
	user        string
	group       string
	filesystem  string
	apiver      string
	labels      string
	name        string
	ignitionver string
)

func init() {
	// https://coreos.com/ignition/docs/latest/configuration-v2_2.html
	flag.StringVar(&file, "file", "", "The absolute path to the file [Required]")
	flag.StringVar(&format, "format", "base64", "Format to convert the file")
	flag.IntVar(&mode, "mode", 420, "File's permission mode in octal")
	flag.BoolVar(&signature, "signature", false, "Include sha512 signature")
	flag.StringVar(&user, "user", "", "The user name of the owner")
	flag.StringVar(&group, "group", "", "The group name of the owner")
	flag.StringVar(&filesystem, "filesystem", "root", "The internal identifier of the filesystem in which to write the file")
	flag.StringVar(&apiver, "apiversion", "machineconfiguration.openshift.io/v1", "MachineConfig API version")
	flag.StringVar(&labels, "labels", "", "MachineConfig metadata labels (separted by ,)")
	flag.StringVar(&name, "name", "", "MachineConfig object name")
	flag.StringVar(&ignitionver, "ignitionversion", "2.2", "Ignition version")
}

func main() {

	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 {
		printUsage()
	}

	// if file is not provided, print usage
	if file == "" {
		printUsage()
	}

	if labels != "" {
    labellist := strings.Split(labels, ",")
    fmt.Printf("Labels %s", labellist[0])
  }



	//fmt.Printf("filePtr: %s, formatPtr: %s, modePtr: %i, verifyPtr: %b, userPtr: %s, groupPtr: %s, filesystemPtr: %s\n", *filePtr, *formatPtr, *modePtr, *verifyPtr, *userPtr, *groupPtr, *filesystemPtr)

	/*
		apiVersion: machineconfiguration.openshift.io/v1
		kind: MachineConfig
		metadata:
		  labels:
		    machineconfiguration.openshift.io/role: master
		  name: masters-chrony-configuration
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
		          source: data:text/plain;charset=utf-8;base64,c2VydmVyIGNsb2NrLnJlZGhhdC5jb20gaWJ1cnN0CmRyaWZ0ZmlsZSAvdmFyL2xpYi9jaHJvbnkvZHJpZnQKbWFrZXN0ZXAgMS4wIDMKcnRjc3luYwpsb2dkaXIgL3Zhci9sb2cvY2hyb255Cg==
		          verification: {}
		        filesystem: root
		        mode: 420
		        path: /etc/chrony.conf
		  osImageURL: ""

	*/
}

func toBase64(file string) (string, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString([]byte(f)), nil
}

func printUsage() {
	fmt.Printf("Usage: %s --file /path/to/my/file.txt [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Printf("Example:\n%s --file /path/to/my/file.txt --label \"machineconfiguration.openshift.io/role: master\",\"example.com/labela: value\"\n", os.Args[0])
	os.Exit(1)
}
