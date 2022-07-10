package main

import (
	"context"
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jr64/dockerchk/dockerhub"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func dropPrivilegesIfRoot() {
	nobodyUsername := viper.GetString("nobody_username")

	uid := syscall.Getuid()

	if uid == 0 {

		nobody, err := user.Lookup(nobodyUsername)
		if err != nil {
			log.Fatalf("Failed to look up user %s: %v", nobodyUsername, err)
		}

		nobodyGid, err := strconv.Atoi(nobody.Gid)
		if err != nil {
			log.Fatalf("failed to convert gid %s to int: %v", nobody.Gid, err)
		}
		nobodyUid, err := strconv.Atoi(nobody.Uid)
		if err != nil {
			log.Fatalf("failed to convert uid %s to int: %v", nobodyUsername, err)
		}
		if err := syscall.Setgid(nobodyGid); err != nil {
			log.Fatalf("Failed to setgid: %v", err)
		}

		if err := syscall.Setuid(nobodyUid); err != nil {
			log.Fatalf("Failed to setuid: %v", err)
		}

		log.Debugf("Successfully dropped privileges to UID %d and GID %d", syscall.Getuid(), syscall.Getgid())

	}
}
func main() {

	log.SetFormatter(&log.TextFormatter{TimestampFormat: "", FullTimestamp: true})
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	viper.SetEnvPrefix("DOCKERCHK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	flag.StringP("images", "i", "", "space separated list of images to check")
	flag.String("nobody-username", "nobody", "drop from root privileges to this user")
	flag.BoolP("verbose", "v", false, "verbose output")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	// replace - with _ in flags so we can use the snake_case version when accessing through viper
	normalizeFunc := flag.CommandLine.GetNormalizeFunc()
	flag.CommandLine.SetNormalizeFunc(func(fs *pflag.FlagSet, name string) pflag.NormalizedName {
		result := normalizeFunc(fs, name)
		name = strings.ReplaceAll(string(result), "-", "_")
		return pflag.NormalizedName(name)
	})

	viper.BindPFlags(flag.CommandLine)

	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to connect to docker: %v", err)
	}

	// get all running containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		log.Fatalf("failed to list docker containers: %v", err)
	}

	err = cli.Close()
	if err != nil {
		log.Fatalf("failed to close docker connection: %v", err)
	}

	dropPrivilegesIfRoot()

	images := strings.Split(viper.GetString("images"), " ")

	cache := make(map[string]string, len(containers))
	var digest string

	for _, container := range containers {

		found := false
		for idx := range images {
			if strings.EqualFold(images[idx], container.Image) {
				found = true
			}
		}
		if found || (len(images) == 1 && images[0] == "") {
			if cache[container.Image] == "" {
				digest, err = dockerhub.GetContainerDigest(dockerhub.ParseContainerIdentifier(container.Image))
				if err != nil {
					log.Fatalf("failed to fetch digest for %s: %v", container.Image, err)
				}
				cache[container.Image] = digest
			} else {
				digest = cache[container.Image]
			}

			if strings.EqualFold(strings.ToLower(digest), strings.ToLower(container.ImageID)) {
				log.Debugf("container %s (%s) with image %s (%s) is up to date", strings.Join(container.Names, " "), container.ID, container.Image, container.ID)
			} else {
				log.Debugf("container %s (%s) with image %s (%s) has newer version %s available", strings.Join(container.Names, " "), container.ID, container.Image, container.ID, digest)
				fmt.Println(container.ID)
			}
		}
	}
}
