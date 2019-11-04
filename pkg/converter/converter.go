package converter

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"

	igntypes "github.com/coreos/ignition/config/v2_2/types"
	MachineConfig "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Parameters Struct containing all the parameters required
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
	Plain       bool
	Yaml        bool
}

// Default values
var defaultFilesystem = "root"
var defaultIgnitionVersion = "2.2.0"
var defaultMachineConfigPrefix = "99-"
var defaultLabel = "machineconfiguration.openshift.io/role: worker"
var defaultApiversion = "machineconfiguration.openshift.io/v1"

// fileToBase64 Encode a file to base64
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

// fileToPlain Create a single string from a file
func fileToPlain(file string) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(f)
}

// labelsToMap Creates a string map with the labels the user provides
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

// CheckParameters Normalize parameters
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
		rawdata.IgnitionVer = defaultIgnitionVersion
	case rawdata.IgnitionVer != defaultIgnitionVersion:
		log.Fatalf("Ignition version must be %s", defaultIgnitionVersion)
	default:
		log.Fatalf("You shouldn't fail here...")
	}

	// Normalize stuff

	// Remote path = local path if not explicitely used
	if rawdata.RemotePath == "" {
		if runtime.GOOS == "windows" {
			log.Fatalf("If running on Windows, remote location is mandatory")
		} else {
			rawdata.RemotePath, err = filepath.Abs(rawdata.LocalPath)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("remote not provided, using '%s' as the original file\n", rawdata.RemotePath)
		}
	} else {
		if filepath.IsAbs(rawdata.RemotePath) == false {
			log.Fatalf("%s is not an absolute path", rawdata.RemotePath)
		}
	}

	// Normalize name
	if rawdata.Name == "" {
		if runtime.GOOS == "windows" {
			log.Fatalf("If running on Windows, name is mandatory")
		} else {
			var nodetype string
			if strings.Contains(rawdata.Labels, "master") {
				nodetype = "master"
			} else {
				nodetype = "worker"
			}
			r := strings.NewReplacer("/", "-", ".", "-")
			rawdata.Name = strings.TrimSpace(defaultMachineConfigPrefix + nodetype + r.Replace(rawdata.RemotePath))
			log.Printf("name not provided, using '%s' as name\n", rawdata.Name)
		}
	}

	// Set label if not provided
	if rawdata.Labels == "" {
		log.Printf("labels not provided, using '%s' by default", defaultLabel)
		rawdata.Labels = defaultLabel
	}

	// Set filesystem if not provided
	if rawdata.Filesystem == "" {
		log.Printf("filesystem not provided, using '%s' by default", defaultFilesystem)
		rawdata.Filesystem = defaultFilesystem
	}

	// Set apiver if not provided
	if rawdata.APIVer == "" {
		log.Printf("apiver not provided, using '%s' by default", defaultApiversion)
		rawdata.APIVer = defaultApiversion
	}

	SetUserGroupMode(file, rawdata)
}

// NewMachineConfig Creates the MachineConfig object
func NewMachineConfig(data Parameters) MachineConfig.MachineConfig {

	// Default content will be base64
	fileContent := "data:text/plain;charset=utf-8;base64,"

	if data.Plain == true {
		fileContent = "data:," + fileToPlain(data.LocalPath)

	} else {
		// Create the base64 data with the proper ignition prefix
		fileContent += fileToBase64(data.LocalPath)
	}

	// Create a map with the labels (as required by the machine-config struct)
	labelmap := labelsToMap(data.Labels)

	// So far, a single file is supported
	file := make([]igntypes.File, 1)
	file[0].FileEmbedded1 = igntypes.FileEmbedded1{
		Mode: &data.Mode,
		Contents: igntypes.FileContents{
			Source: fileContent,
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

// MachineConfigOutput Convert a MachineConfig to a string
func MachineConfigOutput(mc MachineConfig.MachineConfig, mode string) string {

	switch {
	case mode == "json":
		b, err := json.Marshal(mc)
		if err != nil {
			log.Fatal(err)
		}
		return string(b)
	case mode == "yaml":
		b, err := yaml.Marshal(mc)
		if err != nil {
			log.Fatal(err)
		}
		return string(b)
	default:
		log.Fatalf("You shouldn't fail here...")
	}

	return ""
}
