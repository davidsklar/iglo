package iglo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	version "github.com/hashicorp/go-version"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

func ParseJSON(r io.Reader) (*API, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	api := new(API)
	err = json.Unmarshal(b, &api)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func ParseMarkdown(r io.Reader) ([]byte, error) {
	path, err := drafter()
	if err != nil {
		return nil, err
	}

	err = detectVersion()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	echo := exec.Command("echo", string(b))
	out, err := echo.StdoutPipe()
	if err != nil {
		return nil, err
	}

	echo.Start()

	cmd := exec.Command(path, "--format", "json")
	cmd.Stdin = out

	return cmd.Output()
}

func CheckVersion(v string) error {
	mv, _ := version.NewVersion("0.1.0")
	ov, err := version.NewVersion(v)
	if err != nil {
		return err
	}

	if ov.LessThan(mv) {
		return errors.New(fmt.Sprintf("You are using drafter version %s. Minimum version should be %s", ov, mv))
	}

	return nil
}

func detectVersion() error {
	v, err := drafterVersion()
	if err != nil {
		return err
	}

	err = CheckVersion(v)
	if err != nil {
		return err
	}

	return nil
}

func drafter() (string, error) {
	path, err := exec.LookPath("drafter")
	if err != nil {
		return "", errors.New("Couldn't find drafter. Please install it first https://github.com/apiaryio/drafter")
	}

	return path, nil
}

func drafterVersion() (string, error) {
	var cmd *exec.Cmd

	path, err := drafter()
	if err != nil {
		return "", err
	}

	var stderr bytes.Buffer
	cmd = exec.Command(path, "--help")
	cmd.Stderr = &stderr
	err = cmd.Run()

	// returns 0.0.0 if drafter doesn't return version
	if !strings.Contains(stderr.String(), "--version") {
		return "0.0.0", nil
	}

	cmd = exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.Replace(string(output), "v", "", 1)), nil
}
