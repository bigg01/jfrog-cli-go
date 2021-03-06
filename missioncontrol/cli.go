package missioncontrol

import (
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/commands"
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/commands/rtinstances"
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"fmt"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "rt-instances",
			Aliases: []string{"rti"},
			Usage:   "Artifactory instances",
			Subcommands: getRtiSubCommands(),
		},
		{
			Name:    "config",
			Flags:   getConfigFlags(),
			Aliases: []string{"c"},
			Usage:   "Configure Mission Control details",
			Action: configure,
		},
	}
}

func getRtiSubCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "add",
			Flags:  getAddInstanceFlags(),
			Usage:  "Add an instance",
			Action: addInstance,
		},
		{
			Name:   "remove",
			Flags:  getRemoveInstanceFlags(),
			Usage:  "Remove an instance",
			Action: removeInstance,
		},
		{
			Name:   "attach-lic",
			Flags:  getAttachLicenseFlags(),
			Usage:  "Attach license to an instance",
			Action: attachLicense,
		},
		{
			Name:   "detach-lic",
			Flags:  getDetachLicenseFlags(),
			Usage:  "Dettach license from an instance",
			Action: detachLicense,
		},
	}
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Mission Control URL",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Mission Control username",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Mission Control password",
		},
	}
}

func getRemoveInstanceFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "quiet",
			Value: "",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.",
		},
	}...)
}

func getAddInstanceFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "rt-url",
			Usage: "[Mandatory] Artifactory URL.",
		},
		cli.StringFlag{
			Name:  "rt-user",
			Usage: "[Mandatory] Artifactory admin username.",
		},
		cli.StringFlag{
			Name:  "rt-password",
			Usage: "[Mandatory] Artifactory admin password - optionally encrypted.",
		},
		cli.StringFlag{
			Name:  "desc",
			Usage: "[Optional] Artifactory instance description.",
		},
		cli.StringFlag{
			Name:  "location",
			Usage: "[Optional] Artifactory instance location.",
		},
	}...)
}

func getAttachLicenseFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "bucket-id",
			Usage: "[Mandatory] license bucket ID",
		},
		cli.StringFlag{
			Name:  "node-id",
			Usage: "[Optional] Unique HA node identifier",
		},
		cli.StringFlag{
			Name:  "license-path",
			Usage: "[Optional] Full path to the license file",
		},
		cli.StringFlag{
			Name:  "override",
			Usage: "[Default: false] Set to true to override licence file.",
		},
		cli.StringFlag{
			Name:  "deploy",
			Usage: "[Default: false] Set to true to deploy licence to instace.",
		},
	}...)
}

func getDetachLicenseFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "bucket-id",
			Usage: "[Mandatory] license bucket ID",
		},
		cli.StringFlag{
			Name:  "node-id",
			Usage: "[Optional] Unique HA node identifier",
		},
	}...)
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "interactive",
			Usage: "[Default: true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.",
		},
	}
	return append(flags, getFlags()...)
}

func addInstance(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	addInstanceFlags, err := createAddInstanceFlag(c)
    if err != nil {
        cliutils.Exit(cliutils.ExitCodeError, err.Error())
    }
	rtinstances.AddInstance(c.Args()[0], addInstanceFlags)
}

func removeInstance(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	instanceName := c.Args()[0];
	if !c.Bool("quiet") {
		var confirm string
		fmt.Print("Remove Instance,  " + instanceName + "? (y/n): ")
		fmt.Scanln(&confirm)
		if !cliutils.ConfirmAnswer(confirm) {
			return
		}
	}
	flags, err := createRemoveInstanceFlags(c)
    if err != nil {
        cliutils.Exit(cliutils.ExitCodeError, err.Error())
    }
	rtinstances.Remove(instanceName, flags)
}

func attachLicense(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	flags, err := createAttachLicFlags(c)
    if err != nil {
        cliutils.Exit(cliutils.ExitCodeError, err.Error())
    }
	rtinstances.AttachLic(c.Args()[0], flags)
}

func detachLicense(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	flags, err := createDetachLicFlags(c)
    if err != nil {
        cliutils.Exit(cliutils.ExitCodeError, err.Error())
    }
	rtinstances.DetachLic(c.Args()[0], flags)
}

func offerConfig(c *cli.Context) (*config.MissionControlDetails, error) {
    exists, err := config.IsMissionControlConfExists()
	if err != nil {
	    return nil, err
	}
	if exists {
		return nil, nil
	}
	val, err := cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
	if err != nil {
	    return nil, err
	}
	if !val {
		config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	msg := "The CLI commands require the Mission Control URL and authentication details\n" +
	"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n" +
	"You can also configure these parameters later using the 'config' command.\n" +
	"Configure now? (y/n): "
	fmt.Print(msg)
	var confirm string
	fmt.Scanln(&confirm)
	if !cliutils.ConfirmAnswer(confirm) {
		config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	details, err := createMissionControlDetails(c, false)
	if err != nil {
	    return nil, err
    }
	return commands.Config(nil, details, true)
}

func configure(c *cli.Context) {
	if len(c.Args()) > 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	} else if len(c.Args()) == 1 {
		if c.Args()[0] == "show" {
			commands.ShowConfig()
		} else if c.Args()[0] == "clear" {
			commands.ClearConfig()
		} else {
			cliutils.Exit(cliutils.ExitCodeError, "Unknown argument '"+c.Args()[0]+"'. Available arguments are 'show' and 'clear'.")
		}
	} else {
		flags, err := createConfigFlags(c)
        cliutils.ExitOnErr(err)
		commands.Config(flags.MissionControlDetails, nil, flags.Interactive)
	}
}

func createDetachLicFlags(c *cli.Context) (flags *rtinstances.DetachLicFlags, err error) {
	flags = new(rtinstances.DetachLicFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true);
	if err != nil {
	    return
	}
	if flags.BucketId = c.String("bucket-id"); flags.BucketId == ""{
		cliutils.Exit(cliutils.ExitCodeError, "The --bucket-id option is mandatory")
	}
	flags.NodeId = c.String("node-id")
	return
}

func createAttachLicFlags(c *cli.Context) (flags *rtinstances.AttachLicFlags, err error) {
	flags = new(rtinstances.AttachLicFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
	    return
	}
	flags.LicensePath = c.String("license-path")
	if strings.HasSuffix(flags.LicensePath, ioutils.GetFileSeperator()) {
		cliutils.Exit(cliutils.ExitCodeError, "The --license-path option cannot be a directory")
	}
	if flags.BucketId = c.String("bucket-id"); flags.BucketId == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --bucket-id option is mandatory")
	}
	flags.Override = cliutils.GetBoolFlagValue(c, "override", false)
	flags.Deploy = cliutils.GetBoolFlagValue(c, "deploy", false)
	flags.NodeId = c.String("node-id")
	return
}

func createConfigFlags(c *cli.Context) (flags *commands.ConfigFlags, err error) {
	flags = new(commands.ConfigFlags)
	flags.Interactive = cliutils.GetBoolFlagValue(c, "interactive", true)
	flags.MissionControlDetails, err = createMissionControlDetails(c, false)
	if err != nil {
	    return
	}
	if !flags.Interactive && flags.MissionControlDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory when the --interactive option is set to false")
	}
	return
}

func createAddInstanceFlag(c *cli.Context) (flags *rtinstances.AddInstanceFlags, err error) {
	flags = new(rtinstances.AddInstanceFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
	    return
	}
	flags.ArtifactoryInstanceDetails = new(utils.ArtifactoryInstanceDetails)
	if flags.ArtifactoryInstanceDetails.Url = c.String("rt-url"); flags.ArtifactoryInstanceDetails.Url == ""{
		cliutils.Exit(cliutils.ExitCodeError, "The --rt-url option is mandatory")
	}
	if flags.ArtifactoryInstanceDetails.User = c.String("rt-user"); flags.ArtifactoryInstanceDetails.User == ""{
		cliutils.Exit(cliutils.ExitCodeError, "The --rt-user option is mandatory")
	}
	if flags.ArtifactoryInstanceDetails.Password = c.String("rt-password"); flags.ArtifactoryInstanceDetails.Password == ""{
		cliutils.Exit(cliutils.ExitCodeError, "The --rt-password option is mandatory test")
	}
	flags.Description = c.String("desc")
	flags.Location = c.String("location")
	return
}

func createRemoveInstanceFlags(c *cli.Context) (flags *rtinstances.RemoveFlags, err error) {
    details, err := createMissionControlDetails(c, true)
    if err != nil {
        return
    }
	flags = &rtinstances.RemoveFlags{
		MissionControlDetails: details,
		Interactive:		   cliutils.GetBoolFlagValue(c, "interactive", true)}

    return
}

func createMissionControlDetails(c *cli.Context, includeConfig bool) (*config.MissionControlDetails, error) {
	if includeConfig {
		details, err := offerConfig(c)
		if err != nil {
		    return nil, err
		}
		if details != nil {
			return details, nil
		}
	}
	details := new(config.MissionControlDetails)
	details.Url = c.String("url")
	details.User = c.String("user")
	details.Password = c.String("password")

	if includeConfig {
		if details.Url == "" || details.User == "" || details.Password == "" {
			confDetails, err := commands.GetConfig()
            if err != nil {
                return nil, err
            }
			if details.Url == "" {
				details.Url = confDetails.Url
			}
			if details.User == "" {
				details.SetUser(confDetails.User)
			}
			if details.Password == "" {
				details.SetPassword(confDetails.Password)
			}
		}
	}
	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	return details, nil
}

