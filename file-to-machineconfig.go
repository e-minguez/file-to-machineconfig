package main

import (
	json "encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/e-minguez/file-to-machineconfig/pkg/converter"
)

func printUsage() {
	fmt.Printf("Usage: %s --file /local/path/to/my/file.txt [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Printf("Example:\n%s --file /local/path/to/my/file.txt --filepath /path/to/remote/file.txt --label \"machineconfiguration.openshift.io/role: master\",\"example.com/foo: bar\"\n", os.Args[0])
	os.Exit(1)
}

func main() {

	data := converter.Parameters{}

	// https://coreos.com/ignition/docs/latest/configuration-v2_2.html
	flag.StringVar(&data.LocalPath, "file", "", "The path to the local file [Required]")
	flag.StringVar(&data.RemotePath, "remote", "", "The absolute path to the remote file [Required if running on Windows]")
	flag.StringVar(&data.Name, "name", "", "MachineConfig object name [Required if running on Windows]")
	flag.StringVar(&data.Labels, "labels", "", "MachineConfig metadata labels (separted by ,)")
	flag.StringVar(&data.User, "user", "", "The user name of the owner")
	flag.StringVar(&data.Group, "group", "", "The group name of the owner")
	flag.StringVar(&data.Filesystem, "filesystem", "", "The internal identifier of the filesystem in which to write the file")
	flag.StringVar(&data.APIVer, "apiversion", "", "MachineConfig API version")
	flag.StringVar(&data.IgnitionVer, "ignitionversion", "", "Ignition version")
	flag.IntVar(&data.Mode, "mode", 0, "File's permission mode in octal")

	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 || data.LocalPath == "" {
		printUsage()
	}

	if runtime.GOOS == "windows" && (data.LocalPath == "" || data.RemotePath == "" || data.Name == "") {
		printUsage()
	}

	// Some sanity checks/normalization
	converter.CheckParameters(&data)

	// Fill the machine-config struct
	mc := converter.NewMachineConfig(data)

	// Convert the machine-config struct to json
	// TO-DO: either json or yaml
	b, err := json.Marshal(mc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
