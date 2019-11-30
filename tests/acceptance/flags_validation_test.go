package acceptance

import (
	. "github.com/smartystreets/goconvey/convey"
	"os/exec"
	"testing"
)

func TestFlagsValidation(t *testing.T) {
	compileExecutable(t)

	Convey("Should fail if bucket not provided", t, func() {
		cmd := exec.Command("./funnel", "--region=us-east-1")
		cmdOutput, err := cmd.CombinedOutput()

		So(err, ShouldNotBeNil)
		So(string(cmdOutput), ShouldContainSubstring, "must specify an AWS S3 bucket to save files in")
	})

	Convey("Should fail if region not provided as flag or env var", t, func() {
		cmd := exec.Command("./funnel", "--bucket=some-cool-bucket")
		cmdOutput, err := cmd.CombinedOutput()

		So(err, ShouldNotBeNil)
		So(string(cmdOutput), ShouldContainSubstring, "must provide an AWS region where your S3 bucket exists")
	})
}
