package adapter

import (
	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/gofiber/fiber/v2"
)

type PhysicalDatabaseInfo struct {
	PhysicalDatabaseID string            `json:"physicalDatabaseId"`
	Type               string            `json:"type"`
	Labels             map[string]string `json:"labels"`
	APIVersions        APIVersionsInfo   `json:"apiVersions"`
	Features           map[string]bool   `json:"features"`
	SupportedRoles     []string          `json:"supportedRoles"`
	ReadOnlyHost       string            `json:"readOnlyHost"`
}

type APIVersionsInfo struct {
	Specs []APIVersionSpec `json:"specs"`
}

type APIVersionSpec struct {
	SpecRootURL     string `json:"specRootUrl"`
	Major           int    `json:"major"`
	Minor           int    `json:"minor"`
	SupportedMajors []int  `json:"supportedMajors"`
}

// GetPhysicalDatabaseHandler builds the response once at startup since all
// values are immutable for the lifetime of the adapter process.
func GetPhysicalDatabaseHandler(physicalDatabaseID string, labels map[string]string, features map[string]bool, roles []string, roHost string) func(c *fiber.Ctx) error {
	response := PhysicalDatabaseInfo{
		PhysicalDatabaseID: physicalDatabaseID,
		Type:               "redis",
		Labels:             labels,
		APIVersions: APIVersionsInfo{
			Specs: []APIVersionSpec{
				{
					SpecRootURL:     dao.RootUrl,
					Major:           dao.MajorAPIVersion,
					Minor:           dao.MinorAPIVersion,
					SupportedMajors: dao.SupportedMajorsVersions,
				},
			},
		},
		Features:       features,
		SupportedRoles: roles,
		ReadOnlyHost:   roHost,
	}

	return func(c *fiber.Ctx) error {
		return c.JSON(response)
	}
}
