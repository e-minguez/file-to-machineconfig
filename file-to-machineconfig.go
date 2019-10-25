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

func printUsage() {
	fmt.Printf("Usage: %s --file /local/path/to/my/file.txt [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Printf("Example:\n%s --file /local/path/to/my/file.txt --filepath /path/to/remote/file.txt --label \"machineconfiguration.openshift.io/role: master\",\"example.com/foo: bar\"\n", os.Args[0])
	os.Exit(1)
}

func fileToBase64(file string) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	encodedcontent := b64.StdEncoding.EncodeToString([]byte(f))
	if encodedcontent == "" {
		log.Fatal("The content of the file couldn't be encoded in base64")
	}
	return encodedcontent
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
	switch {
	case rawdata.ignitionver == "":
		rawdata.ignitionver = "2.2"
	case rawdata.ignitionver != "2.2":
		log.Fatalf("Ignition version must be 2.2")
	default:
		log.Fatalf("You shouldn't fail here...")
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

	// Set label if not provided
	if rawdata.labels == "" {
		defaultLabel := "machineconfiguration.openshift.io/role: worker"
		log.Printf("labels not provided, using '%s' by default", defaultLabel)
		rawdata.labels = defaultLabel
	}

	// Set filesystem if not provided
	if rawdata.filesystem == "" {
		defaultFilesystem := "root"
		log.Printf("filesystem not provided, using '%s' by default", defaultFilesystem)
		rawdata.filesystem = defaultFilesystem
	}

	// Set apiver if not provided
	if rawdata.apiver == "" {
		defaultApiversion := "machineconfiguration.openshift.io/v1"
		log.Printf("apiver not provided, using '%s' by default", defaultApiversion)
		rawdata.apiver = defaultApiversion
	}

}

func newMachineConfig(data parameters) MachineConfig.MachineConfig {

	// Create the base64 data with the proper ignition prefix
	base64Content := "data:text/plain;charset=utf-8;base64," + fileToBase64(data.localpath)

	// Create a map with the labels (as required by the machine-config struct)
	labelmap := labelsToMap(data.labels)

	// So far, a single file is supported
	file := make([]igntypes.File, 1)
	file[0].FileEmbedded1 = igntypes.FileEmbedded1{
		Mode: &data.mode,
		Contents: igntypes.FileContents{
			Source: base64Content,
		},
	}
	file[0].Node = igntypes.Node{
		Filesystem: data.filesystem,
		Path:       data.remotepath,
		User: &igntypes.NodeUser{
			Name: data.user,
		},
		Group: &igntypes.NodeGroup{
			Name: data.group,
		},
	}

	mc := MachineConfig.MachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineConfig",
			APIVersion: data.apiver,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   data.name,
			Labels: labelmap,
		},
		Spec: MachineConfig.MachineConfigSpec{
			Config: igntypes.Config{
				Storage: igntypes.Storage{
					Files: file,
				},
				Ignition: igntypes.Ignition{
					Version: data.ignitionver,
				},
			},
		},
	}

	return mc
}

func main() {

	data := parameters{}

	// https://coreos.com/ignition/docs/latest/configuration-v2_2.html
	flag.StringVar(&data.localpath, "file", "", "The path to the local file [Required]")
	flag.StringVar(&data.remotepath, "remote", "", "The absolute path to the remote file")
	flag.StringVar(&data.name, "name", "", "MachineConfig object name")
	flag.StringVar(&data.labels, "labels", "", "MachineConfig metadata labels (separted by ,)")
	flag.StringVar(&data.user, "user", "", "The user name of the owner")
	flag.StringVar(&data.group, "group", "", "The group name of the owner")
	flag.StringVar(&data.filesystem, "filesystem", "", "The internal identifier of the filesystem in which to write the file")
	flag.StringVar(&data.apiver, "apiversion", "", "MachineConfig API version")
	flag.StringVar(&data.ignitionver, "ignitionversion", "", "Ignition version")
	flag.IntVar(&data.mode, "mode", 0, "File's permission mode in octal")

	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 || data.localpath == "" {
		printUsage()
	}

	// Some sanity checks/normalization
	checkParameters(&data)

	// Fill the machine-config struct
	mc := newMachineConfig(data)

	// Convert the machine-config struct to json
	// TO-DO: either json or yaml
	b, err := json.Marshal(mc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
