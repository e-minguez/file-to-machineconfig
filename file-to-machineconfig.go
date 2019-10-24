package main

import (
	b64 "encoding/base64"
	json "encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	igntypes "github.com/coreos/ignition/config/v2_2/types"
	MachineConfig "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type parameters struct {
	localpath   string
	remotepath  string
	name        string
	labels      string
	user        string
	group       string
	filesystem  string
	apiver      string
	ignitionver string
	content     string
	mode        int
}

func newMachineConfig(apiver, name, ignitionver, filesystem, username, groupname, remotepath, base64Content string, mode int, labelmap map[string]string) MachineConfig.MachineConfig {
	// So far, a single file is supported
	file := make([]igntypes.File, 1)
	file[0].FileEmbedded1 = igntypes.FileEmbedded1{
		Mode: &mode,
		Contents: igntypes.FileContents{
			Source: base64Content,
		},
	}
	file[0].Node = igntypes.Node{
		Filesystem: filesystem,
		Path:       remotepath,
		User: &igntypes.NodeUser{
			Name: username,
		},
		Group: &igntypes.NodeGroup{
			Name: groupname,
		},
	}

	mc := MachineConfig.MachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineConfig",
			APIVersion: apiver,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labelmap,
		},
		Spec: MachineConfig.MachineConfigSpec{
			Config: igntypes.Config{
				Storage: igntypes.Storage{
					Files: file,
				},
				Ignition: igntypes.Ignition{
					Version: ignitionver,
				},
			},
		},
	}

	return mc
}

func fileToBase64(file string) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return b64.StdEncoding.EncodeToString([]byte(f))
}

func printUsage() {
	fmt.Printf("Usage: %s --file /local/path/to/my/file.txt [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Printf("Example:\n%s --file /local/path/to/my/file.txt --filepath /path/to/remote/file.txt --label \"machineconfiguration.openshift.io/role: master\",\"example.com/foo: bar\"\n", os.Args[0])
	os.Exit(1)
}

func labelsToMap(labels string) map[string]string {
	// Remove blanks and split the labels by the comma
	entries := strings.Split((strings.Replace(labels, " ", "", -1)), ",")

	// https://stackoverflow.com/questions/48465575/easy-way-to-split-string-into-map-in-go
	labelmap := make(map[string]string)
	for _, e := range entries {
		parts := strings.Split(e, ":")
		labelmap[parts[0]] = parts[1]
	}
	return labelmap
}

func checkParameters(rawdata *parameters) {
	// Check for errors first

	// Verify file exists
	file, err := os.Stat(rawdata.localpath)
	if os.IsNotExist(err) {
		log.Fatalf("File %s doesn't exist", rawdata.localpath)
	}

	// Verify is not a directory
	if file.IsDir() {
		log.Fatalf("File %s is a directory", rawdata.localpath)
	}

	// TODO: Verify remotepath is a file path

	// Ignition 2.2 only ¯\_(ツ)_/¯
	if rawdata.ignitionver != "2.2" {
		log.Fatalf("Ignition version must be 2.2")
	}

	// Normalize stuff

	// Remote path = local path if not explicitely used
	if rawdata.remotepath == "" {
		rawdata.remotepath, err = filepath.Abs(rawdata.localpath)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("remote not provided, using '%s' as the original file\n", rawdata.remotepath)
	}

	// Normalize name
	if rawdata.name == "" {
		var nodetype string
		if strings.Contains(rawdata.labels, "master") {
			nodetype = "master"
		} else {
			nodetype = "worker"
		}
		r := strings.NewReplacer("/", "-", ".", "-")
		rawdata.name = strings.TrimSpace("99-" + nodetype + r.Replace(rawdata.remotepath))
		log.Printf("name not provided, using '%s' as name\n", rawdata.name)
	}

	// Copy file mode if not provided
	if rawdata.mode == 0 {
		filemode := file.Mode().Perm()
		log.Printf("mode not provided, using '%#o' as the original file", filemode)
		// Ignition requires decimal
		rawdata.mode = int(filemode)
	}

	// Copy file user if not provided
	if rawdata.user == "" {
		fileuser, _ := user.LookupId(strconv.Itoa(int(file.Sys().(*syscall.Stat_t).Uid)))
		log.Printf("user not provided, using '%s' as the original file", fileuser.Username)
		rawdata.user = fileuser.Username
	}

	// Copy file group if not provided
	if rawdata.group == "" {
		filegroup, _ := user.LookupId(strconv.Itoa(int(file.Sys().(*syscall.Stat_t).Gid)))
		log.Printf("group not provided, using '%s' as the original file", filegroup.Username)
		rawdata.group = filegroup.Username
	}

}

func main() {

	data := parameters{}

	// https://coreos.com/ignition/docs/latest/configuration-v2_2.html
	flag.StringVar(&data.localpath, "file", "", "The path to the local file [Required]")
	flag.StringVar(&data.remotepath, "remote", "", "The absolute path to the remote file")
	flag.StringVar(&data.name, "name", "", "MachineConfig object name")
	flag.StringVar(&data.labels, "labels", "machineconfiguration.openshift.io/role: worker", "MachineConfig metadata labels (separted by ,)")
	flag.IntVar(&data.mode, "mode", 0, "File's permission mode in octal")
	flag.StringVar(&data.user, "user", "", "The user name of the owner")
	flag.StringVar(&data.group, "group", "", "The group name of the owner")
	flag.StringVar(&data.filesystem, "filesystem", "root", "The internal identifier of the filesystem in which to write the file")
	flag.StringVar(&data.apiver, "apiversion", "machineconfiguration.openshift.io/v1", "MachineConfig API version")
	flag.StringVar(&data.ignitionver, "ignitionversion", "2.2", "Ignition version")

	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 || data.localpath == "" {
		printUsage()
	}

	// Some sanity checks/normalization
	checkParameters(&data)

	// Create the base64 data with the proper ignition prefix
	base64Content := "data:text/plain;charset=utf-8;base64," + fileToBase64(data.localpath)

	// Create a map with the labels (as required by the machine-config struct)
	labelmap := labelsToMap(data.labels)

	// Fill the machine-config struct
	mc := newMachineConfig(data.apiver, data.name, data.ignitionver, data.filesystem, data.user, data.group, data.remotepath, base64Content, data.mode, labelmap)

	// Convert the machine-config struct to json
	// TO-DO: either json or yaml
	b, err := json.Marshal(mc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
