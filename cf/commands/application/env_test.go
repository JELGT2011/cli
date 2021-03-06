package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("env command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *testApplication.FakeApplicationRepository
		configRepo          core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		app = models.Application{}
		app.Name = "my-app"
		appRepo = &testApplication.FakeApplicationRepository{}
		appRepo.ReadReturns.App = app

		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func(args ...string) {
		cmd := NewEnv(ui, configRepo, appRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("Requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("my-app")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("fails with usage when no app name is given", func() {
		runCommand()

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("fails with usage when the app cannot be found", func() {
		appRepo.ReadReturns.Error = errors.NewModelNotFoundError("app", "hocus-pocus")
		runCommand("hocus-pocus")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"not found"},
		))
	})

	Context("when the app has at least one env var", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "the-app-guid"

			appRepo.ReadReturns.App = app
			appRepo.ReadEnvReturns(&models.Environment{
				Environment: map[string]string{
					"my-key":    "my-value",
					"my-key2":   "my-value2",
					"first-key": "Zer0",
				},
				System: map[string]interface{}{
					"VCAP_SERVICES": map[string]interface{}{
						"pump-yer-brakes": "drive-slow",
					},
				},
			}, nil)
		})

		It("lists those environment variables, in sorted order for provided services", func() {
			runCommand("my-app")
			Expect(appRepo.ReadEnvArgsForCall(0)).To(Equal("the-app-guid"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"System-Provided:"},
				[]string{"VCAP_SERVICES", ":", "{"},
				[]string{"pump-yer-brakes", ":", "drive-slow"},
				[]string{"}"},
				[]string{"User-Provided:"},
				[]string{"first-key", "Zer0"},
				[]string{"my-key", "my-value"},
				[]string{"my-key2", "my-value2"},
			))
		})
	})

	Context("when the app has no user-defined environment variables", func() {
		It("shows an empty message", func() {
			appRepo.ReadEnvReturns(&models.Environment{}, nil)
			runCommand("my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app"},
				[]string{"OK"},
				[]string{"No", "system-provided", "env variables", "have been set"},
				[]string{"No", "env variables", "have been set"},
			))
		})
	})

	Context("when the app has no environment variables", func() {
		It("informs the user that each group is empty", func() {
			appRepo.ReadEnvReturns(&models.Environment{}, nil)

			runCommand("my-app")
			Expect(ui.Outputs).To(ContainSubstrings([]string{"No system-provided env variables have been set"}))
			Expect(ui.Outputs).To(ContainSubstrings([]string{"No user-defined env variables have been set"}))
			Expect(ui.Outputs).To(ContainSubstrings([]string{"No running env variables have been set"}))
			Expect(ui.Outputs).To(ContainSubstrings([]string{"No staging env variables have been set"}))
		})
	})

	Context("when the app has at least one running and staging environment variable", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "the-app-guid"

			appRepo.ReadReturns.App = app
			appRepo.ReadEnvReturns(&models.Environment{
				Running: map[string]string{
					"running-key-1": "running-value-1",
					"running-key-2": "running-value-2",
				},
				Staging: map[string]string{
					"staging-key-1": "staging-value-1",
					"staging-key-2": "staging-value-2",
				},
			}, nil)
		})

		It("lists the environment variables", func() {
			runCommand("my-app")
			Expect(appRepo.ReadEnvArgsForCall(0)).To(Equal("the-app-guid"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Running Environment Variable Groups:"},
				[]string{"running-key-1", ":", "running-value-1"},
				[]string{"running-key-2", ":", "running-value-2"},
				[]string{"Staging Environment Variable Groups:"},
				[]string{"staging-key-1", ":", "staging-value-1"},
				[]string{"staging-key-2", ":", "staging-value-2"},
			))
		})
	})

	Context("when reading the environment variables returns an error", func() {
		It("tells you about that error", func() {
			appRepo.ReadEnvReturns(nil, errors.New("BOO YOU CANT DO THAT; GO HOME; you're drunk"))
			runCommand("whatever")
			Expect(ui.Outputs).To(ContainSubstrings([]string{"you're drunk"}))
		})
	})
})
