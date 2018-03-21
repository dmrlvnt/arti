package utils

import (
	"errors"
	artlog "github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/mitchellh/colorstring"
	"github.com/orange-cloudfoundry/artifactory-resource/model"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	ART_SECURITY_FOLDER = "security/"
)

func CheckReqParamsWithPattern(source model.Source) error {
	if source.Pattern == "" {
		return errors.New("You must provide a pattern for your file (e.g.: 'local_generic/myfile.txt','local_generic/my*.txt'.")
	}
	return CheckReqParams(source)
}
func CheckReqParams(source model.Source) error {
	if source.Url == "" {
		return errors.New("You must pass an url to artifactory.")
	}
	if source.User == "" && source.ApiKey == "" {
		return errors.New("You must pass user/password pair or apiKey to authnticate over artifactory.")
	}
	return nil
}
func RetrieveArtDetails(source model.Source) (*config.ArtifactoryDetails, error) {
	err := createCert(source.CACert)
	if err != nil {
		return nil, err
	}
	sshKeyPath, err := createSshKeyPath(source.SshKey)
	return &config.ArtifactoryDetails{
		Url:        AddTrailingSlashIfNeeded(source.Url),
		User:       source.User,
		Password:   source.Password,
		ApiKey:     source.ApiKey,
		SshKeyPath: sshKeyPath,
	}, nil

}

func AddTrailingSlashIfNeeded(path string) string {
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
func RemoveStartingSlashIfNeeded(path string) string {
	if path != "" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	return path
}
func createCert(caCert string) error {
	if caCert == "" {
		return nil
	}
	confPath, err := config.GetJfrogHomeDir()
	if err != nil {
		return err
	}
	securityPath := confPath + ART_SECURITY_FOLDER
	os.MkdirAll(securityPath, os.ModePerm)
	return ioutil.WriteFile(securityPath+"cert.pem", []byte(caCert), 0644)
}
func createSshKeyPath(sshKey string) (string, error) {
	if sshKey == "" {
		return "", nil
	}
	file, err := ioutil.TempFile(os.TempDir(), "ssh-key")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(sshKey)
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Dir(file.Name()))
}
func OverrideLoggerArtifactory(logLevel string) {
	logger := artlog.Logger().(*artlog.CliLogger)
	logger.LogLevel = artlog.LogLevel["INFO"]
	if val, ok := artlog.LogLevel[strings.ToUpper(logLevel)]; ok {
		logger.LogLevel = val
	}
	logger.DebugLog = log.New(os.Stderr, colorstring.Color("[cyan][Artifactory Debug] "), 0)
	logger.InfoLog = log.New(os.Stderr, colorstring.Color("[blue][Artifactory Info] "), 0)
	logger.WarnLog = log.New(os.Stderr, colorstring.Color("[yellow][Artifactory Warn] "), 0)
	logger.ErrorLog = log.New(os.Stderr, colorstring.Color("[red][Artifactory Error] "), 0)
}
