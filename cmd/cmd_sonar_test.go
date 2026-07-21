package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	if err != nil {
		fmt.Println(err)
	}

	return buf.String(), err
}

func Test_SonarDiscoverStaticCmd_illegal_type(t *testing.T) {
	t.Setenv("CONSTELLIX_API_KEY", "test-key")
	t.Setenv("CONSTELLIX_SECRET_KEY", "test-secret")
	rootCmd.AddCommand(sonarCmd)
	sonarCmd.AddCommand(sonarDiscoverCmd)
	sonarDiscoverCmd.AddCommand(sonarDiscoverStaticCmd)
	output, _ := executeCommand(rootCmd, "sonar", "discover", "static", "--type", "illegal")
	if !strings.Contains(output, "unsupported resource type: got \"illegal\", want one of") {
		t.Errorf("Expected error message, got %s", output)
	}
}
