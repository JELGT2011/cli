package organization

import (
	"bytes"
	"fmt"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
	"text/tabwriter"
)

type ShowOrg struct {
	ui     terminal.UI
	config core_config.Reader
	orgReq requirements.OrganizationRequirement
}

func NewShowOrg(ui terminal.UI, config core_config.Reader) (cmd *ShowOrg) {
	cmd = new(ShowOrg)
	cmd.ui = ui
	cmd.config = config
	return
}

func (cmd *ShowOrg) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "org",
		Description: T("Show org info"),
		Usage:       T("CF_NAME org ORG"),
	}
}

func (cmd *ShowOrg) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return
}

func (cmd *ShowOrg) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()
	cmd.ui.Say(T("Getting info for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(org.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))
	cmd.ui.Ok()
	cmd.ui.Say("\n%s:", terminal.EntityNameColor(org.Name))

	domains := []string{}
	for _, domain := range org.Domains {
		domains = append(domains, domain.Name)
	}

	spaces := []string{}
	for _, space := range org.Spaces {
		spaces = append(spaces, space.Name)
	}

	quota := org.QuotaDefinition
	orgQuota := fmt.Sprintf(T("{{.QuotaName}} ({{.MemoryLimit}}M memory limit, {{.RoutesLimit}} routes, {{.ServicesLimit}} services, paid services {{.NonBasicServicesAllowed}})",
		map[string]interface{}{
			"QuotaName":               quota.Name,
			"MemoryLimit":             quota.MemoryLimit,
			"RoutesLimit":             quota.RoutesLimit,
			"ServicesLimit":           quota.ServicesLimit,
			"NonBasicServicesAllowed": formatters.Allowed(quota.NonBasicServicesAllowed)}))
	writer := new(tabwriter.Writer)
	output := new(bytes.Buffer)

	writer.Init(output, 0, 4, 2, ' ', 0)

	fmt.Fprintln(writer, T("\tdomains:\t{{.Domains}}", map[string]interface{}{"Domains": terminal.EntityNameColor(strings.Join(domains, ", "))}))
	fmt.Fprintln(writer, T("\tquota:\t{{.Quota}}", map[string]interface{}{"Quota": terminal.EntityNameColor(orgQuota)}))
	fmt.Fprintln(writer, T("\tspaces:\t{{.Spaces}}", map[string]interface{}{"Spaces": terminal.EntityNameColor(strings.Join(spaces, ", "))}))

	writer.Flush()
	cmd.ui.Say(output.String())
}
