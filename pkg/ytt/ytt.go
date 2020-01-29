package ytt

import (
	"bytes"
	"fmt"
	"io"
	goexec "os/exec"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
)

type Ytt struct {
	data *schema.ResourceData
}

func (t *Ytt) Template() (string, string, error) {
	args, stdin, err := t.addArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building args: %s", err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("ytt", args...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs

	err = cmd.Run()
	if err != nil {
		stderrStr := stderrBs.String()
		return "", stderrStr, fmt.Errorf("Executing ytt: %s (stderr: %s)", err, stderrStr)
	}

	return stdoutBs.String(), "", nil
}

func (t *Ytt) addArgs() ([]string, io.Reader, error) {
	args := []string{}
	var stdin io.Reader

	files := t.data.Get(schemaFilesKey).([]interface{})
	if len(files) > 0 {
		for _, file := range files {
			args = append(args, "--file="+file.(string))
		}
	}

	if t.data.Get(schemaIgnoreUnknownCommentsKey).(bool) {
		args = append(args, "--ignore-unknown-comments")
	}

	values := t.data.Get(schemaValuesYAMLKey).(string)
	if len(values) > 0 {
		args = append(args, "-f-")

		values, err := schemamisc.Heredoc{values}.StripIndent()
		if err != nil {
			return nil, nil, fmt.Errorf("Formatting %s: %s", schemaValuesYAMLKey, err)
		}

		stdin = bytes.NewReader([]byte(values))
	}

	for k, v := range t.data.Get(schemaValuesKey).(map[string]interface{}) {
		args = append(args, []string{"--data-value", k + "=" + v.(string)}...)
	}

	return args, stdin, nil
}
