package main_test

import (
	"io/ioutil"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/google/uuid"
)

var _ = Describe("Integration", func() {
	buildpack := "https://github.com/dgodd/staticfile-buildpack"
	var appName, appDir string
	var session *gexec.Session
	JustBeforeEach(func() {
		appName = uuid.New().String()
		cmd := exec.Command("cf", "push", appName, "-b", buildpack, "-p", appDir, "--no-start")
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())

		cmd = exec.Command("cf", "logs", appName)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		cmd = exec.Command("cf", "start", appName)
		err = cmd.Start()
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("Connected, tailing logs for app"))
	})

	Context("staticfile_app", func() {
		BeforeEach(func() {
			appDir = "cf_spec/fixtures/staticfile_app"
		})
		It("Has the buildpack correct version", func() {
			bpVersion, err := ioutil.ReadFile("./VERSION")
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out, 2*time.Minute).Should(gbytes.Say("Buildpack version " + string(bpVersion)))
		})
	})
})
