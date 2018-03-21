package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	chelper "github.com/ArthurHlt/go-concourse-helper"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	artutils "github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
  "github.com/orange-cloudfoundry/artifactory-resource/model"
	"github.com/orange-cloudfoundry/artifactory-resource/utils"
	"strings"
)

type Out struct {
	cmd        *chelper.OutCommand
	source     model.Source
	params     model.OutParams
	artdetails *config.ArtifactoryDetails
	spec       *artutils.SpecFiles
}

func main() {
	Out := &Out{
		cmd: chelper.NewOutCommand(),
	}
	Out.Run()
}
func (c *Out) Run() {
	cmd := c.cmd
	msg := c.cmd.Messager()
	err := cmd.Source(&c.source)
	msg.FatalIf("Error when parsing source from concourse", err)
	utils.OverrideLoggerArtifactory(c.source.LogLevel)

	err = cmd.Params(&c.params)
	msg.FatalIf("Error when parsing params from concourse", err)
	if c.params.Target == "" {
		msg.Fatal("You must set a target (in the form of: [repository_name]/[repository_path]) in out parameter.")
	}

	c.defaultingParams()

	err = utils.CheckReqParams(c.source)
	if err != nil {
		msg.Fatal(err.Error())
	}
	c.artdetails, err = utils.RetrieveArtDetails(c.source)
	if err != nil && !strings.Contains(err.Error(), "You must provide a pattern") {
		msg.Fatal(err.Error())
	}
	src := c.folderPath(c.params.Source)
	target := utils.AddTrailingSlashIfNeeded(c.params.Target)

	props := c.mergeProps()

	c.spec = artutils.CreateSpec(src, target, props, true, true, c.source.Regexp)
	msg.Log("[blue]Uploading[reset] file(s) to target '[blue]%s[reset]'...", target)
	startDl := time.Now()
	origStdout := os.Stdout
	os.Stdout = os.Stderr
	totalUploaded, totalFailed, err := c.Upload()
	os.Stdout = origStdout
	msg.FatalIf("Error when uploading", err)
	if totalFailed > 0 {
		msg.Fatal(fmt.Sprintf("%d files failed to upload", totalFailed))
	}
	elapsed := time.Since(startDl)
	msg.Log("[blue]Finished uploading[reset] file(s) to target '[blue]%s[reset]'.", target)

	cmd.Send([]chelper.Metadata{
		{
			Name:  "total_uploaded",
			Value: fmt.Sprintf("%d", totalUploaded),
		},
		{
			Name:  "upload_time",
			Value: elapsed.String(),
		},
	})
}

func (c *Out) defaultingParams() {
	if c.params.Threads <= 0 {
		c.params.Threads = 3
	}
}
func (c Out) folderPath(p string) string {
	src := utils.AddTrailingSlashIfNeeded(c.cmd.SourceFolder())
	src += utils.RemoveStartingSlashIfNeeded(p)
	return src
}
func (c Out) Upload() (totalUploaded, totalFailed int, err error) {
	return commands.Upload(
		c.spec,
		&commands.UploadFlags{
			ArtDetails:     c.artdetails,
			Threads:        c.params.Threads,
			ExplodeArchive: c.params.ExplodeArchive,
		},
	)
}
func (c Out) mergeProps() string {
	msg := c.cmd.Messager()
	props := ""
	if c.params.Props != "" {
		props = c.params.Props
	}
	if c.params.Props != "" && c.params.PropsFromFile != "" {
		props += ";"
	}
	if c.params.PropsFromFile != "" {
		dat, err := ioutil.ReadFile(c.folderPath(c.params.PropsFromFile))
		if err != nil {
			msg.Logln("Could not read file with props from path: %s; %v", c.params.PropsFromFile, err)
			msg.Fatal("error opening file")
		}
		props += string(dat)
	}

	return props
}
