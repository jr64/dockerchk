package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jr64/dockerchk/cmdline"
	"github.com/jr64/dockerchk/dockerhub"
	"github.com/jr64/dockerchk/priv"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {

	flag.StringP("images", "i", "", "space separated list of images to check")
	flag.String("nobody-username", "nobody", "drop from root privileges to this user")
	flag.BoolP("verbose", "v", false, "print verbose output")
	flag.Bool("debug", false, "print debug output")

	cmdline.Setup("DOCKERCHK")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to connect to docker: %v", err)
	}

	// get only running containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		log.Fatalf("Failed to list docker containers: %v", err)
	}

	// close exiting connections before dropping privileges
	err = cli.Close()
	if err != nil {
		log.Fatalf("Failed to close docker connection: %v", err)
	}

	// drop privileges
	if dropped, err := priv.Drop(viper.GetString("nobody_username")); err != nil {
		log.Fatalf("Failed to drop privileges: %v", err)
	} else {
		if dropped {
			log.Debugf("Successfully dropped privileges to UID %d and GID %d", syscall.Getuid(), syscall.Getgid())
		} else {
			log.Debugf("Not running as root, keeping UID %d and GID %d", syscall.Getuid(), syscall.Getgid())
		}
	}

	images := strings.Split(viper.GetString("images"), " ")

	cache := make(map[string]string, len(containers))
	var digest string

	updatesRequired := false

	//check all running containers
	for _, container := range containers {

		found := false
		for idx := range images {
			if strings.EqualFold(images[idx], container.Image) {
				found = true
			}
		}
		//one of the images we are interested in?
		if found || (len(images) == 1 && images[0] == "") {

			//fetch current digest from cache or DockerHub
			if cache[container.Image] == "" {
				digest, err = dockerhub.GetContainerDigest(dockerhub.ParseContainerIdentifier(container.Image))
				if err != nil {
					log.Fatalf("Failed to fetch digest for %s: %v", container.Image, err)
				}
				cache[container.Image] = digest
			} else {
				digest = cache[container.Image]
			}

			if strings.EqualFold(strings.ToLower(digest), strings.ToLower(container.ImageID)) {
				log.Infof("Container %s (%s) with image %s (%s) is up to date", strings.Join(container.Names, " "), container.ID, container.Image, container.ID)
			} else {
				log.Infof("Container %s (%s) with image %s (%s) has newer version %s available", strings.Join(container.Names, " "), container.ID, container.Image, container.ID, digest)
				fmt.Println(container.ID)
				updatesRequired = true
			}
		}
	}

	if updatesRequired {
		os.Exit(2)
	} else {
		os.Exit(0)
	}
}
