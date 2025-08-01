package resourcemanager

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"
)

func DataSourceGoogleFolder() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFolderRead,
		Schema: map[string]*schema.Schema{
			"folder": {
				Type:     schema.TypeString,
				Required: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lifecycle_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lookup_organization": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"organization": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"configured_capabilities": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"management_project": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFolderRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*transport_tpg.Config)
	userAgent, err := tpgresource.GenerateUserAgentString(d, config.UserAgent)
	if err != nil {
		return err
	}

	id := canonicalFolderName(d.Get("folder").(string))
	d.SetId(id)
	if err := resourceGoogleFolderRead(d, meta); err != nil {
		return err
	}
	// If resource doesn't exist, read will not set ID and we should return error.
	if d.Id() == "" {
		return fmt.Errorf("%s not found", id)
	}

	if err := d.Set("deletion_protection", nil); err != nil {
		return fmt.Errorf("Error setting deletion_protection: %s", err)
	}

	if v, ok := d.GetOk("lookup_organization"); ok && v.(bool) {
		organization, err := lookupOrganizationName(d.Id(), userAgent, d, config)
		if err != nil {
			return err
		}

		if err := d.Set("organization", organization); err != nil {
			return fmt.Errorf("Error setting organization: %s", err)
		}
	}

	return nil
}

func canonicalFolderName(ba string) string {
	if strings.HasPrefix(ba, "folders/") {
		return ba
	}

	return "folders/" + ba
}

func lookupOrganizationName(parent, userAgent string, d *schema.ResourceData, config *transport_tpg.Config) (string, error) {
	if parent == "" || strings.HasPrefix(parent, "organizations/") {
		return parent, nil
	} else if strings.HasPrefix(parent, "folders/") {
		parentFolder, err := getGoogleFolder(parent, userAgent, d, config)
		if err != nil {
			return "", fmt.Errorf("Error getting parent folder '%s': %s", parent, err)
		}
		return lookupOrganizationName(parentFolder.Parent, userAgent, d, config)
	} else {
		return "", fmt.Errorf("Unknown parent type '%s' on folder '%s'", parent, d.Id())
	}
}
