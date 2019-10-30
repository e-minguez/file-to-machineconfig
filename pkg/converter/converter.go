package converter

import (
	b64 "encoding/base64"
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

type Parameters struct {
	LocalPath   string
	RemotePath  string
	Name        string
	Labels      string
	User        string
	Group       string
	Filesystem  string
	APIVer      string
	IgnitionVer string
	Content     string
	Mode        int
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

func CheckParameters(rawdata *Parameters) {
	// Check for errors first

	// Verify file exists
	file, err := os.Stat(rawdata.LocalPath)
	if os.IsNotExist(err) {
		log.Fatalf("File %s doesn't exist", rawdata.LocalPath)
	}

	// Verify is not a directory
	if file.IsDir() {
		log.Fatalf("File %s is a directory", rawdata.LocalPath)
	}

	// TODO: Verify RemotePath is a file path

	// Ignition 2.2 only ¯\_(ツ)_/¯
	switch {
	case rawdata.IgnitionVer == "":
		rawdata.IgnitionVer = "2.2.0"
	case rawdata.IgnitionVer != "2.2.0":
		log.Fatalf("Ignition version must be 2.2.0")
	default:
		log.Fatalf("You shouldn't fail here...")
	}

	// Normalize stuff

	// Remote path = local path if not explicitely used
	if rawdata.RemotePath == "" {
		rawdata.RemotePath, err = filepath.Abs(rawdata.LocalPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("remote not provided, using '%s' as the original file\n", rawdata.RemotePath)
	}

	// Normalize name
	if rawdata.Name == "" {
		var nodetype string
		if strings.Contains(rawdata.Labels, "master") {
			nodetype = "master"
		} else {
			nodetype = "worker"
		}
		r := strings.NewReplacer("/", "-", ".", "-")
		rawdata.Name = strings.TrimSpace("99-" + nodetype + r.Replace(rawdata.RemotePath))
		log.Printf("name not provided, using '%s' as name\n", rawdata.Name)
	}

	// Copy file mode if not provided
	if rawdata.Mode == 0 {
		filemode := file.Mode().Perm()
		log.Printf("mode not provided, using '%#o' as the original file", filemode)
		// Ignition requires decimal
		rawdata.Mode = int(filemode)
	}

	// Copy file user if not provided
	if rawdata.User == "" {
		fileuser, _ := user.LookupId(strconv.Itoa(int(file.Sys().(*syscall.Stat_t).Uid)))
		log.Printf("user not provided, using '%s' as the original file", fileuser.Username)
		rawdata.User = fileuser.Username
	}

	// Copy file group if not provided
	if rawdata.Group == "" {
		filegroup, _ := user.LookupId(strconv.Itoa(int(file.Sys().(*syscall.Stat_t).Gid)))
		log.Printf("group not provided, using '%s' as the original file", filegroup.Username)
		rawdata.Group = filegroup.Username
	}

	// Set label if not provided
	if rawdata.Labels == "" {
		defaultLabel := "machineconfiguration.openshift.io/role: worker"
		log.Printf("labels not provided, using '%s' by default", defaultLabel)
		rawdata.Labels = defaultLabel
	}

	// Set filesystem if not provided
	if rawdata.Filesystem == "" {
		defaultFilesystem := "root"
		log.Printf("filesystem not provided, using '%s' by default", defaultFilesystem)
		rawdata.Filesystem = defaultFilesystem
	}

	// Set apiver if not provided
	if rawdata.APIVer == "" {
		defaultApiversion := "machineconfiguration.openshift.io/v1"
		log.Printf("apiver not provided, using '%s' by default", defaultApiversion)
		rawdata.APIVer = defaultApiversion
	}

}

func NewMachineConfig(data Parameters) MachineConfig.MachineConfig {

	// Create the base64 data with the proper ignition prefix
	base64Content := "data:text/plain;charset=utf-8;base64," + fileToBase64(data.LocalPath)

	// Create a map with the labels (as required by the machine-config struct)
	labelmap := labelsToMap(data.Labels)

	// So far, a single file is supported
	file := make([]igntypes.File, 1)
	file[0].FileEmbedded1 = igntypes.FileEmbedded1{
		Mode: &data.Mode,
		Contents: igntypes.FileContents{
			Source: base64Content,
		},
	}
	file[0].Node = igntypes.Node{
		Filesystem: data.Filesystem,
		Path:       data.RemotePath,
		User: &igntypes.NodeUser{
			Name: data.User,
		},
		Group: &igntypes.NodeGroup{
			Name: data.Group,
		},
	}

	mc := MachineConfig.MachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineConfig",
			APIVersion: data.APIVer,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   data.Name,
			Labels: labelmap,
		},
		Spec: MachineConfig.MachineConfigSpec{
			Config: igntypes.Config{
				Storage: igntypes.Storage{
					Files: file,
				},
				Ignition: igntypes.Ignition{
					Version: data.IgnitionVer,
				},
			},
		},
	}

	return mc
}
