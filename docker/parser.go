package docker

import (
	"io/ioutil"
	"log"
	"os/user"
	"sort"
	"strconv"
	"strings"

	"github.com/cloudcredo/cloudfocker/config"
)

func ParseRunCommand(config *config.ContainerConfig) (runCmd []string) {
	runCmd = append(runCmd, userString())
	runCmd = append(runCmd, parseContainerName(config.ContainerName)...)
	runCmd = append(runCmd, parseDaemon(config.Daemon)...)
	runCmd = append(runCmd, parseMounts(config.Mounts)...)
	runCmd = append(runCmd, parsePublishedPorts(config.PublishedPorts)...)
	runCmd = append(runCmd, parseEnvVars(config.EnvVars)...)
	runCmd = append(runCmd, parseImageTag(config.ImageTag)...)
	runCmd = append(runCmd, parseCommand(config.Command)...)
	return
}

func WriteRuntimeDockerfile(config *config.ContainerConfig) {
	var dockerfile string

	dockerfile = bootstrapDockerfileString()
	dockerfile = dockerfile + envVarDockerfileString(config.EnvVars)
	dockerfile = dockerfile + commandDockerfileString(config.Command)

	ioutil.WriteFile(config.DropletDir+"/Dockerfile", []byte(dockerfile), 0644)
}

func userString() string {
	var thisUser *user.User
	var err error
	if thisUser, err = user.Current(); err != nil {
		log.Fatalf(" %s", err)
	}
	return "-u=" + thisUser.Uid
}

func parseContainerName(containerName string) (parsedContainerName []string) {
	parsedContainerName = append(parsedContainerName, "--name="+containerName)
	return
}

func parseDaemon(daemon bool) (parsedDaemon []string) {
	var daemonString string
	if daemon {
		daemonString = "-d"
		parsedDaemon = append(parsedDaemon, daemonString)
	}
	return
}

func parseMounts(mounts map[string]string) (parsedMounts []string) {
	for src, dst := range mounts {
		parsedMounts = append(parsedMounts,
			"--volume="+src+":"+dst)
	}
	sort.Strings(parsedMounts)
	return
}

func parsePublishedPorts(publishedPorts map[int]int) (parsedPublishedPorts []string) {
	for host, cont := range publishedPorts {
		parsedPublishedPorts = append(parsedPublishedPorts,
			"--publish="+strconv.Itoa(host)+":"+strconv.Itoa(cont))
	}
	sort.Strings(parsedPublishedPorts)
	return
}

func parseEnvVars(envVars map[string]string) (parsedEnvVars []string) {
	for key, val := range envVars {
		if val != "" {
			parsedEnvVars = append(parsedEnvVars,
				"--env=\""+key+"="+val+"\"")
		}
	}
	sort.Strings(parsedEnvVars)
	return
}

func parseImageTag(imageTag string) (parsedImageTag []string) {
	parsedImageTag = append(parsedImageTag, imageTag)
	return
}

func parseCommand(command []string) (parsedCommand []string) {
	parsedCommand = append(parsedCommand, command...)
	return
}

func bootstrapDockerfileString() string {
	return `FROM cloudfocker-base:latest
RUN /usr/sbin/useradd -mU -u 10000 -s /bin/bash vcap
COPY droplet.tgz /app/
RUN chown vcap:vcap /app && cd /app && su vcap -c "tar zxf droplet.tgz" && rm droplet.tgz
EXPOSE 8080
USER vcap
WORKDIR /app
`
}

func envVarDockerfileString(envVars map[string]string) string {
	var envVarStrings []string
	for envVarKey, envVarVal := range envVars {
		if envVarVal != "" {
			envVarStrings = append(envVarStrings, "ENV "+envVarKey+" "+envVarVal+"\n")
		}
	}
	sort.Strings(envVarStrings)
	return strings.Join(envVarStrings, "")
}

func commandDockerfileString(command []string) string {
	commandString := `CMD ["`
	commandString = commandString + strings.Join(command, `", "`)
	commandString = commandString + "\"]\n"
	return commandString
}
