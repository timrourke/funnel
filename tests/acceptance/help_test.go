package acceptance

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/exec"
	"testing"
)

var didCompileExecutable bool

func compileExecutable(t *testing.T) {
	if didCompileExecutable {
		return
	}

	err := os.Chdir("../..")
	if err != nil {
		t.Fatal("failed to change dirs", err)
	}

	build := exec.Command("go", "build", ".")
	buildOutput, err := build.CombinedOutput()
	if err != nil {
		t.Fatal("failed to build binary", string(buildOutput), err)
	}

	didCompileExecutable = true
}

func TestHelpMenu(t *testing.T) {
	compileExecutable(t)

	Convey("Should show help menu", t, func() {
		cmd := exec.Command("funnel", "-h")
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(string(cmdOutput))
		}

		So(string(cmdOutput), ShouldContainSubstring, "Usage:")
		So(string(cmdOutput), ShouldNotContainSubstring, "Error:")
	})
}
