package main

import (
	b64 "encoding/base64"
	json "encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	igntypes "github.com/coreos/ignition/config/v2_2/types"
	MachineConfig "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

var (
	file     string
	filepath string
	name     string
	labels   string
	mode     int
	/*user        string
	group       string*/
	filesystem  string
	apiver      string
	ignitionver string
	content     string
)

func init() {
	// https://coreos.com/ignition/docs/latest/configuration-v2_2.html
	flag.StringVar(&file, "file", "", "The absolute path to the local file [Required]")
	flag.StringVar(&filepath, "filepath", "", "The absolute path to the remote file [Required]")
	flag.StringVar(&name, "name", "", "MachineConfig object name")
	flag.StringVar(&labels, "labels", "machineconfiguration.openshift.io/role: worker", "MachineConfig metadata labels (separted by ,)")
	flag.IntVar(&mode, "mode", 420, "File's permission mode in octal")
	/*flag.StringVar(&user, "user", "root", "The user name of the owner")
	flag.StringVar(&group, "group", "root", "The group name of the owner")*/
	flag.StringVar(&filesystem, "filesystem", "root", "The internal identifier of the filesystem in which to write the file")
	flag.StringVar(&apiver, "apiversion", "machineconfiguration.openshift.io/v1", "MachineConfig API version")
	flag.StringVar(&ignitionver, "ignitionversion", "2.2", "Ignition version")
}

func newMachineConfig(apiver string, name string, ignitionver string, filesystem string, mode int, filepath string, base64Content string, labelmap map[string]string) MachineConfig.MachineConfig {

	filecontent := igntypes.FileContents{
		Source: "data:text/plain;charset=utf-8;base64," + base64Content,
	}

	fileembedded1 := igntypes.FileEmbedded1{
		Mode:     &mode,
		Contents: filecontent,
	}

	node := igntypes.Node{
		Filesystem: filesystem,
		Path:       filepath,
	}

	file := make([]igntypes.File, 1)
	file[0].FileEmbedded1 = fileembedded1
	file[0].Node = node

	storage := igntypes.Storage{
		Files: file,
	}

	mcspec := MachineConfig.MachineConfigSpec{}
	mcspec.Config.Ignition.Version = ignitionver
	mcspec.Config.Storage = storage

	mc := MachineConfig.MachineConfig{}
	mc.APIVersion = apiver
	mc.Kind = "MachineConfig"
	mc.Name = name
	mc.Labels = labelmap
	mc.Spec = mcspec

	return mc
}

func fileToBase64(file string) (string, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return b64.StdEncoding.EncodeToString([]byte(f)), nil
}

func printUsage() {
	fmt.Printf("Usage: %s --file /local/path/to/my/file.txt --filepath /path/to/remote/file.txt [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Printf("Example:\n%s --file /local/path/to/my/file.txt --filepath /path/to/remote/file.txt --label \"machineconfiguration.openshift.io/role: master\",\"example.com/foo: bar\"\n", os.Args[0])
	os.Exit(1)
}

func labelsToMap(labels string) map[string]string {
	// https://stackoverflow.com/questions/48465575/easy-way-to-split-string-into-map-in-go
	labelmap := make(map[string]string)
	entries := strings.Split(labels, ",")
	for _, e := range entries {
		parts := strings.Split(e, ":")
		labelmap[parts[0]] = parts[1]
	}
	return labelmap
}

func main() {

	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 {
		printUsage()
	}

	// if file is not provided, print usage
	if file == "" || filepath == "" {
		printUsage()
	}

	if name == "" {
		r := strings.NewReplacer("/", "-", ".", "-")
		name = "99-worker" + r.Replace(filepath)
		fmt.Printf("name not provided, using %s as name\n", name)
	}
	name = strings.TrimSpace(name)

	base64Content, err := fileToBase64(file)
	if err != nil {
		log.Fatal(err)
	}
	base64Content = "data:text/plain;charset=utf-8;base64," + base64Content

	labelmap := labelsToMap(strings.Replace(labels, " ", "", -1))

	mc := newMachineConfig(apiver, name, ignitionver, filesystem, mode, filepath, base64Content, labelmap)
	b, err := json.Marshal(mc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
