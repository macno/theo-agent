package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"io/ioutil"
)

type SshConfig struct {
	key string
	value string
}

var sshconfigs = []SshConfig{
	SshConfig{"PasswordAuthentication", "no"},
	SshConfig{"AuthorizedKeysFile", "/var/cache/theo-agent/%%u"},
	SshConfig{"AuthorizedKeysCommand", "/usr/sbin/theo-agent"},
	SshConfig{"AuthorizedKeysCommandUser", "theo-agent"},
}

func Install() {
	prepareInstall()
	checkConfig()
	mkdirs()
	writeConfigYaml()
	if *editSshdConfig {
		doEditSshdConfig()
	} else {
		fmt.Fprintf(os.Stderr, "You didn't specify -sshd-config so you have to edit manually /etc/ssh/sshd_config:\n\n")
		i := 0
		for i < len(sshconfigs) {
			line := fmt.Sprintf("%s %s\n", sshconfigs[i].key, sshconfigs[i].value) // I have to go through fmt.Sprintf because of %%u in sshconfigs[i].value
			fmt.Fprintf(os.Stderr, line)
			i++
		}
	}
}

func prepareInstall() {

	askOnce("Theo server URL", theoUrl)
	if *theoUrl == "" {
		fmt.Fprintf(os.Stderr, "Missing required Theo URL\n")
		os.Exit(2)
	}

	askOnce("Theo access token", theoAccessToken)
	if *theoAccessToken == "" {
		fmt.Fprintf(os.Stderr, "Missing required Theo access token\n")
		os.Exit(2)
	}
}

func askOnce(prompt string, result *string) {
	if *noInteractive {
		return
	}

	fmt.Println(prompt)

	if *result != "" {
		fmt.Printf("[%s]: ", *result)
	}

	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}

	data, _, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}

	newResult := string(data)
	newResult = strings.TrimSpace(newResult)

	if newResult != "" {
		*result = newResult
	}
}

func mkdirs() {
	dirs := [2]string{ path.Dir(*configFilePath), *cacheDirPath}
	for i := 0; i < len(dirs); i++ {
		ensureDir(dirs[i])
    }
}

func ensureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			panic(fmt.Sprintf("Unable to create dir (%s): %s", path, err))
		}
	}
}

func checkConfig() {
	ret := Query("test", theoUrl, theoAccessToken)
	if ret > 0 {
		panic(fmt.Sprintf("Check failed, unable to retrieve keys from %s", *theoUrl))
	}
}

func writeConfigYaml() {
	config := fmt.Sprintf("url: %s\ntoken: %s\n", *theoUrl, *theoAccessToken)
	f, err := os.Create(*configFilePath)
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "Unable to write config file (%s): %s", *configFilePath, err)
		}
		os.Exit(21)
	}
	defer f.Close()
    
	_, err = f.WriteString(config)
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "Unable to write config file (%s): %s", *configFilePath, err)
		}
		os.Exit(21)
	}
}

func doEditSshdConfig() bool {

	data, err := ioutil.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "Unable to read %s, %s", "/etc/ssh/sshd_config", err)
		}
		return false
	}
	lines := strings.Split(string(data), "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		ii := 0
		for ii < len(sshconfigs) {
			p := strings.Index(line, sshconfigs[ii].key)
			if p >= 0 {
				lines[i] = fmt.Sprintf("%s %s", sshconfigs[ii].key, sshconfigs[ii].value)
				sshconfigs = remove(sshconfigs, ii)
				break
			}
			ii++
		}
		i++
	}
	ii := 0
	for ii < len(sshconfigs) {
		lines = append(lines, fmt.Sprintf("%s %s", sshconfigs[ii].key, sshconfigs[ii].value))
		ii++
	}

	f, err := os.Create("/etc/ssh/sshd_config")
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "Unable to write config file (%s): %s", "/etc/ssh/sshd_config", err)
		}
		os.Exit(21)
	}
	defer f.Close()
    
	_, err = f.WriteString(strings.Join(lines, "\n"))
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "Unable to write config file (%s): %s", "/etc/ssh/sshd_config", err)
		}
		os.Exit(21)
	}

	return true
}

func remove(s []SshConfig, i int) []SshConfig {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}