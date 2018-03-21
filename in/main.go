package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	fpath "path/filepath"
	"time"

	"crypto/tls"
	"crypto/x509"

	chelper "github.com/ArthurHlt/go-concourse-helper"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	artutils "github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/dmrlvnt/arti/model"
	"github.com/dmrlvnt/arti/utils"
)

type In struct {
	cmd        *chelper.InCommand
	source     model.Source
	params     model.InParams
	artdetails *config.ArtifactoryDetails
	spec       *artutils.SpecFiles
}

func main() {
	in := &In{
		cmd: chelper.NewInCommand(),
	}
	in.Run()
}
func (c *In) Run() {
	cmd := c.cmd
	msg := c.cmd.Messager()
	err := cmd.Source(&c.source)
	msg.FatalIf("Error when parsing source from concourse", err)
	utils.OverrideLoggerArtifactory(c.source.LogLevel)
	err = cmd.Params(&c.params)
	msg.FatalIf("Error when parsing params from concourse", err)
	c.defaultingParams()
	err = utils.CheckReqParams(c.source)
	if err != nil {
		msg.Fatal(err.Error())
	}
	c.artdetails, err = utils.RetrieveArtDetails(c.source)
	if err != nil {
		msg.Fatal(err.Error())
	}

	filePath := c.cmd.Version().BuildNumber
	dest := utils.AddTrailingSlashIfNeeded(c.cmd.DestinationFolder())
	if c.params.Filename != "" {
		dest += c.params.Filename
	} else {
		dest += fpath.Base(filePath)
	}
	c.spec = artutils.CreateSpec(filePath, dest, c.source.Props, false, !c.params.Notflat, false)
	msg.Log("[blue]Downloading[reset] file '[blue]%s[reset]'...", filePath)
	startDl := time.Now()
	origStdout := os.Stdout
	os.Stdout = os.Stderr
	err = c.Download()
	os.Stdout = origStdout
	msg.FatalIf("Error when downloading", err)
	elapsed := time.Since(startDl)
	msg.Log("[blue]Finished downloading[reset] file '[blue]%s[reset]'.", filePath)

	if c.params.PropsFilename != "" {
		msg.Logln("\n[blue]Downloading properties[reset] file '[blue]%s[reset]'.", c.params.PropsFilename)
		err = c.DownloadProperties()
		msg.FatalIf("Error downloading properties", err)
		msg.Logln("\n[blue]Finished downloading properties[reset] file '[blue]%s[reset]'.", c.params.PropsFilename)
	}

	metadata := []chelper.Metadata{
		{
			Name:  "downloaded_file",
			Value: filePath,
		},
		{
			Name:  "download_time",
			Value: elapsed.String(),
		},
	}
	cmd.Send(metadata)
}

func (c *In) defaultingParams() {
	if c.params.Threads <= 0 {
		c.params.Threads = 3
	}
	if c.params.SplitCount <= 0 {
		c.params.SplitCount = 3
	}
	if c.params.MinSplit <= 0 {
		c.params.MinSplit = 5120
	}
}

func (c In) Download() error {
	return commands.Download(
		c.spec,
		&commands.DownloadFlags{
			ArtDetails:   c.artdetails,
			Threads:      c.params.Threads,
			SplitCount:   c.params.SplitCount,
			MinSplitSize: int64(c.params.MinSplit),
		},
	)
}

func (c In) DownloadProperties() error {
	msg := c.cmd.Messager()

	url := fmt.Sprintf("%sapi/storage/%s?properties", c.artdetails.Url, c.cmd.Version().BuildNumber)

	req, err := http.NewRequest("GET", url, nil)
	msg.FatalIf("Error downloading properties", err)
	req.SetBasicAuth(c.artdetails.GetUser(), c.artdetails.GetPassword())
	client := &http.Client{}
	if c.source.CACert != "" {
		caPool := x509.NewCertPool()
		ok := caPool.AppendCertsFromPEM([]byte(c.source.CACert))
		if !ok {
			msg.Logln(fmt.Sprintf("%v", err))
			msg.Fatal("Error parsing pem certificate")
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caPool,
			},
		}
		client.Transport = tr
	}
	resp, err := client.Do(req)
	msg.FatalIf("Error downloading properties", err)
	defer resp.Body.Close()

	if !(resp.StatusCode == 200 || resp.StatusCode == 404) {
		msg.Fatal(fmt.Sprintf("\nCouldn't get properties info. Response code: %d", resp.StatusCode))
	}

	propsfilePath := path.Dir(c.params.PropsFilename)
	basepath := utils.AddTrailingSlashIfNeeded(c.cmd.DestinationFolder())
	if propsfilePath != "." {
		basepath += propsfilePath
	}
	err = os.MkdirAll(basepath, 0777)
	msg.FatalIf(fmt.Sprintf("\nCouldn't create folder: %s", basepath), err)
	out, err := os.Create(utils.AddTrailingSlashIfNeeded(basepath) + path.Base(c.params.PropsFilename))
	msg.FatalIf("Couldn't create properties file", err)
	defer out.Close()

	io.Copy(out, resp.Body)

	return nil
}
