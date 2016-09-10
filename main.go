package main

import (
	"log/syslog"
	"os"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/digitalocean/go-metadata"
	"github.com/docker/go-plugins-helpers/volume"
	flag "github.com/ogier/pflag"
)

const DefaultBaseMountPath = "/mnt/dostorage"

type CommandLineArgs struct {
	accessToken *string
	mountPath   *string
}

func main() {
	syslogHook, herr := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, DriverName)
	if herr == nil {
		logrus.AddHook(syslogHook)
	} else {
		logrus.Warn("it was not possible to activate logging to the local syslog")
	}

	args := parseCommandLineArgs()

	doMetadataClient := metadata.NewClient()
	doAPIClient := NewDoAPIClient(*args.accessToken)
	doFacade := NewDoFacade(doMetadataClient, doAPIClient)

	mountUtil := NewMountUtil()

	driver, derr := NewDriver(doFacade, mountUtil, *args.mountPath)
	if derr != nil {
		logrus.Fatalf("failed to create the driver: %v", derr)
		os.Exit(1)
	}

	handler := volume.NewHandler(driver)

	serr := handler.ServeUnix("root", DriverName)
	if serr != nil {
		logrus.Fatalf("failed to bind to the Unix socket: %v", serr)
		os.Exit(1)
	}

	for {
		// block while requests are served in a separate routine
	}
}

func parseCommandLineArgs() *CommandLineArgs {
	args := &CommandLineArgs{}

	args.accessToken = flag.StringP("access-token", "t", "", "the DigitalOcean API access token")
	args.mountPath = flag.StringP("mount-path", "m", DefaultBaseMountPath, "the path under which to create the volume mount folders")
	flag.Parse()

	if *args.accessToken == "" {
		flag.Usage()
		os.Exit(1)
	}

	return args
}