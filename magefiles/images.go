//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/client"
	"github.com/jacobbrewer1/utils"
	"github.com/magefile/mage/mg"
)

const (
	imageAppSeparator = "/"
)

var (
	dockerRegistry = os.Getenv("DOCKER_REGISTRY")
	envTags        = os.Getenv("TAGS")
	toPushEnv      = os.Getenv("DOCKER_PUSH")
	toPush         = false
)

type Images mg.Namespace

// A build step that requires additional params, or platform specific steps for example
func (i Images) BuildAll() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	toPush, _ = strconv.ParseBool(toPushEnv)
	fmt.Println("Push images set to: ", toPush)

	// Get all directory names in the ./cmd directory
	cmds, err := os.ReadDir("./cmd")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	wg := new(sync.WaitGroup)
	multiErr := utils.NewMultiError()

	// Iterate over each directory
	for _, cmd := range cmds {
		if cmd.IsDir() {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				if err := i.buildImage(cli, name); err != nil {
					multiErr.Add(fmt.Errorf("failed to build image for %s: %w", name, err))
				}

				if toPush {
					if err := i.pushImage(name); err != nil {
						multiErr.Add(fmt.Errorf("failed to push image for %s: %w", name, err))
					}
				}
			}(cmd.Name())
		}
	}

	wg.Wait()

	if multiErr.Err() != nil {
		for _, err := range multiErr.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	return nil
}

func (i Images) buildImage(cli *client.Client, appName string) error {
	applicationDockerRegistry := dockerRegistry + imageAppSeparator + appName
	fmt.Println(applicationDockerRegistry)

	tags := i.imageTags(applicationDockerRegistry)
	fmt.Println(tags)

	cmd := exec.Command("docker", "build")

	for _, tag := range tags {
		cmd.Args = append(cmd.Args, "-t", tag)
	}

	cmd.Args = append(cmd.Args, ".")
	cmd.Args = append(cmd.Args, "--build-arg", "APP_NAME="+appName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	return nil
}

func (i Images) pushImage(appName string) error {
	applicationDockerRegistry := dockerRegistry + imageAppSeparator + appName
	fmt.Println(applicationDockerRegistry)

	tags := i.imageTags(applicationDockerRegistry)
	fmt.Println(tags)
	for _, tag := range tags {
		cmd := exec.Command("docker", "push", tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to push image: %w", err)
		}
	}

	return nil
}

func (i Images) imageTags(registry string) []string {
	envSplit := strings.Split(envTags, ",")
	tags := make([]string, 0)
	for _, tag := range envSplit {
		tags = append(tags, registry+":"+tag)
	}
	return tags
}
