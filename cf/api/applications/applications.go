package applications

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ApplicationRepository interface {
	Create(params models.AppParams) (createdApp models.Application, apiErr error)
	Read(name string) (app models.Application, apiErr error)
	Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiErr error)
	Delete(appGuid string) (apiErr error)
	ReadEnv(guid string) (*models.Environment, error)
	CreateRestageRequest(guid string) (apiErr error)
}

type CloudControllerApplicationRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) Create(params models.AppParams) (createdApp models.Application, apiErr error) {
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiErr = errors.NewWithError(T("Failed to marshal JSON"), err)
		return
	}

	path := fmt.Sprintf("%s/v2/apps", repo.config.ApiEndpoint())
	resource := new(resources.ApplicationResource)
	apiErr = repo.gateway.CreateResource(path, strings.NewReader(data), resource)
	if apiErr != nil {
		return
	}

	createdApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Read(name string) (app models.Application, apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=%s&inline-relations-depth=1", repo.config.ApiEndpoint(), repo.config.SpaceFields().Guid, url.QueryEscape("name:"+name))
	appResources := new(resources.PaginatedApplicationResources)
	apiErr = repo.gateway.GetResource(path, appResources)
	if apiErr != nil {
		return
	}

	if len(appResources.Resources) == 0 {
		apiErr = errors.NewModelNotFoundError("App", name)
		return
	}

	res := appResources.Resources[0]
	app = res.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiErr error) {
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiErr = errors.NewWithError(T("Failed to marshal JSON"), err)
		return
	}

	path := fmt.Sprintf("%s/v2/apps/%s?inline-relations-depth=1", repo.config.ApiEndpoint(), appGuid)
	resource := new(resources.ApplicationResource)
	apiErr = repo.gateway.UpdateResource(path, strings.NewReader(data), resource)
	if apiErr != nil {
		return
	}

	updatedApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) formatAppJSON(input models.AppParams) (data string, err error) {
	appResource := resources.NewApplicationEntityFromAppParams(input)
	bytes, err := json.Marshal(appResource)
	data = string(bytes)
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.ApiEndpoint(), appGuid)
	return repo.gateway.DeleteResource(path)
}

func (repo CloudControllerApplicationRepository) ReadEnv(guid string) (*models.Environment, error) {
	var (
		err error
	)

	path := fmt.Sprintf("%s/v2/apps/%s/env", repo.config.ApiEndpoint(), guid)
	appResource := models.NewEnvironment()

	err = repo.gateway.GetResource(path, appResource)
	if err != nil {
		return &models.Environment{}, err
	}

	return appResource, err
}

func (repo CloudControllerApplicationRepository) CreateRestageRequest(guid string) error {
	path := fmt.Sprintf("%s/v2/apps/%s/restage", repo.config.ApiEndpoint(), guid)
	return repo.gateway.CreateResource(path, strings.NewReader(""), nil)
}
