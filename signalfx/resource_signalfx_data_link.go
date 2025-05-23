// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package signalfx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/signalfx/signalfx-go/datalink"
	"github.com/signalfx/signalfx-go/util"
)

func dataLinkResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"property_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name (key) of the metadata that's the trigger of a data link. If you specify `property_value`, you must specify `property_name`.",
			},
			"property_value": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value of the metadata that's the trigger of a data link. If you specify this property, you must also specify `property_name`.",
			},
			"context_dashboard_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The dashobard ID to which this data link will be applied",
			},
			"target_signalfx_dashboard": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Link to a Splunk Observability Cloud dashboard",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dashboard_group_id": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "SignalFx-assigned ID of the dashboard link target's dashboard group",
						},
						"dashboard_id": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "SignalFx-assigned ID of the dashboard link target",
						},
						"is_default": &schema.Schema{
							Type:        schema.TypeBool,
							Default:     true,
							Optional:    true,
							Description: "Flag that designates a target as the default for a data link object.",
						},
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "User-assigned target name. Use this value to differentiate between the link targets for a data link object.",
						},
					},
				},
			},
			"target_external_url": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Link to an external URL",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "User-assigned target name. Use this value to differentiate between the link targets for a data link object.",
						},
						"time_format": &schema.Schema{
							Type:        schema.TypeString,
							Default:     datalink.ISO8601,
							Optional:    true,
							Description: "Designates the format of minimumTimeWindow in the same data link target object.",
							ValidateFunc: validation.StringInSlice([]string{
								string(datalink.ISO8601), string(datalink.Epoch), string(datalink.EpochSeconds),
							}, false),
						},
						"url": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "URL string for a Splunk instance or external system data link target.",
						},
						"property_key_mapping": &schema.Schema{
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Describes the relationship between Splunk Observability Cloud metadata keys and external system properties when the key names are different",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"minimum_time_window": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "6000",
							Description: "The minimum time window for a search sent to an external site. Depends on the value set for `time_format`.",
						},
					},
				},
			},
			"target_splunk": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Link to a Splunk instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "User-assigned target name. Use this value to differentiate between the link targets for a data link object.",
						},
						"property_key_mapping": &schema.Schema{
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Describes the relationship between Splunk Observability Cloud metadata keys and external system properties when the key names are different",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"target_appd_url": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Link to AppDynamics URL",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "User-assigned target name. Use this value to differentiate between the link targets for a data link object.",
						},
						"url": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "URL string for an AppDyanmics data link target.",
						},
					},
				},
			},
		},

		Create: dataLinkCreate,
		Read:   dataLinkRead,
		Update: dataLinkUpdate,
		Delete: dataLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func getPayloadDataLink(d *schema.ResourceData) (*datalink.CreateUpdateDataLinkRequest, error) {
	dataLink := &datalink.CreateUpdateDataLinkRequest{}

	if name, ok := d.GetOk("property_name"); ok {
		dataLink.PropertyName = name.(string)
	}

	if val, ok := d.GetOk("property_value"); ok {
		if dataLink.PropertyName == "" {
			return dataLink, fmt.Errorf("Must supply a property_name when supplying a property_value")
		}
		dataLink.PropertyValue = val.(string)
	}

	if val, ok := d.GetOk("context_dashboard_id"); ok {
		dataLink.ContextId = val.(string)
	}

	if val, ok := d.GetOk("target_signalfx_dashboard"); ok {
		if dataLink.PropertyName == "" {
			return dataLink, fmt.Errorf("Must supply a property_name when using target_signalfx_dashboard")
		}

		sfxDashes := val.(*schema.Set).List()

		for _, tfLink := range sfxDashes {
			tfLink := tfLink.(map[string]interface{})
			dl := &datalink.Target{
				DashboardGroupId: tfLink["dashboard_group_id"].(string),
				DashboardId:      tfLink["dashboard_id"].(string),
				Name:             tfLink["name"].(string),
				IsDefault:        tfLink["is_default"].(bool),
				Type:             datalink.INTERNAL_LINK,
			}
			if val, ok := tfLink["dashboard_group_name"]; ok {
				dl.DashboardGroupName = val.(string)
			}
			if val, ok := tfLink["dashboard_name"]; ok {
				dl.DashboardName = val.(string)
			}
			dataLink.Targets = append(dataLink.Targets, dl)
		}
	}

	if val, ok := d.GetOk("target_splunk"); ok {
		splkDashes := val.(*schema.Set).List()

		for _, tfLink := range splkDashes {
			tfLink := tfLink.(map[string]interface{})
			dl := &datalink.Target{
				Name: tfLink["name"].(string),
				Type: datalink.SPLUNK_LINK,
			}

			if v, ok := tfLink["property_key_mapping"]; ok {
				pkMap := map[string]string{}
				for key, value := range v.(map[string]interface{}) {
					pkMap[key] = value.(string)
				}
				dl.PropertyKeyMapping = pkMap
			}
			dataLink.Targets = append(dataLink.Targets, dl)
		}
	}

	if val, ok := d.GetOk("target_appd_url"); ok {
		appdURLs := val.(*schema.Set).List()

		appdURLPatternRegex := "^https?:\\/\\/[a-zA-Z0-9-]+\\.saas\\.appdynamics\\.com\\/.*application=\\d+.*component=\\d+.*"
		re, err := regexp.Compile(appdURLPatternRegex)

		if err != nil {
			return dataLink, err
		}

		for _, tfLink := range appdURLs {
			tfLink := tfLink.(map[string]interface{})

			dl := &datalink.Target{
				Name: tfLink["name"].(string),
				URL:  tfLink["url"].(string),
				Type: datalink.APPD_LINK,
			}
			match := re.MatchString(dl.URL)
			if !match {
				return dataLink, fmt.Errorf("enter a valid AppD Link. The link needs to include the contoller URL, application ID, and Application component")
			}

			dataLink.Targets = append(dataLink.Targets, dl)
		}
	}

	if val, ok := d.GetOk("target_external_url"); ok {
		exURLs := val.(*schema.Set).List()

		for _, tfLink := range exURLs {
			tfLink := tfLink.(map[string]interface{})

			dl := &datalink.Target{
				Name:              tfLink["name"].(string),
				MinimumTimeWindow: util.StringOrInteger(tfLink["minimum_time_window"].(string)),
				URL:               tfLink["url"].(string),
				Type:              datalink.EXTERNAL_LINK,
			}

			// When changes are made to an existing target, the Terraform plugin SDK seems
			// to be creating an extraneous target with all empty values. Since name is a
			// required field on targets, skip when we encounter empty names. Ideally this
			// issue would be fixed at the Terraform SDK level - this is only a workaround.
			if dl.Name == "" {
				continue
			}

			switch tfLink["time_format"].(string) {
			case "Epoch":
				dl.TimeFormat = datalink.Epoch
			case "EpochSeconds":
				dl.TimeFormat = datalink.EpochSeconds
			default:
				dl.TimeFormat = datalink.ISO8601
			}

			if v, ok := tfLink["property_key_mapping"]; ok {
				pkMap := map[string]string{}
				for key, value := range v.(map[string]interface{}) {
					pkMap[key] = value.(string)
				}
				dl.PropertyKeyMapping = pkMap
			}

			dataLink.Targets = append(dataLink.Targets, dl)
		}
	}

	if len(dataLink.Targets) < 1 {
		return dataLink, fmt.Errorf("You must provide one or more of `target_signalfx_dashboard`, `target_external_url`, `target_appd_url` or `target_splunk`")
	}

	return dataLink, nil
}

func dataLinkCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload, err := getPayloadDataLink(d)
	if err != nil {
		return err
	}

	debugOutput, _ := json.Marshal(payload)
	log.Printf("[DEBUG] SignalFx: Create Data Link Payload: %s", string(debugOutput))

	dl, err := config.Client.CreateDataLink(context.TODO(), payload)
	if err != nil {
		return err
	}
	d.SetId(dl.Id)
	return dataLinkAPIToTF(d, dl)
}

func dataLinkAPIToTF(d *schema.ResourceData, dl *datalink.DataLink) error {
	debugOutput, _ := json.Marshal(dl)
	log.Printf("[DEBUG] SignalFx: Got Data Link to enState: %s", string(debugOutput))

	if err := d.Set("property_value", dl.PropertyValue); err != nil {
		return err
	}
	if err := d.Set("property_name", dl.PropertyName); err != nil {
		return err
	}
	if err := d.Set("context_dashboard_id", dl.ContextId); err != nil {
		return err
	}

	var internalLinks []map[string]interface{}
	var externalLinks []map[string]interface{}
	var splunkLinks []map[string]interface{}
	var appdLinks []map[string]interface{}

	for _, t := range dl.Targets {
		switch t.Type {
		case datalink.INTERNAL_LINK:
			tfTarget := map[string]interface{}{
				"name":               t.Name,
				"dashboard_group_id": t.DashboardGroupId,
				"dashboard_id":       t.DashboardId,
				"is_default":         t.IsDefault,
			}
			internalLinks = append(internalLinks, tfTarget)
		case datalink.EXTERNAL_LINK:
			tfTarget := map[string]interface{}{
				"name":                 t.Name,
				"minimum_time_window":  t.MinimumTimeWindow,
				"time_format":          t.TimeFormat,
				"url":                  t.URL,
				"property_key_mapping": t.PropertyKeyMapping,
			}
			externalLinks = append(externalLinks, tfTarget)
		case datalink.SPLUNK_LINK:
			tfTarget := map[string]interface{}{
				"name":                 t.Name,
				"property_key_mapping": t.PropertyKeyMapping,
			}
			splunkLinks = append(splunkLinks, tfTarget)
		case datalink.APPD_LINK:
			tfTarget := map[string]interface{}{
				"name": t.Name,
				"url":  t.URL,
			}
			appdLinks = append(appdLinks, tfTarget)
		default:
			return fmt.Errorf("Unknown link type: %s", t.Type)
		}
	}
	if internalLinks != nil && len(internalLinks) > 0 {
		if err := d.Set("target_signalfx_dashboard", internalLinks); err != nil {
			return err
		}
	}
	if externalLinks != nil && len(externalLinks) > 0 {
		if err := d.Set("target_external_url", externalLinks); err != nil {
			return err
		}
	}
	if splunkLinks != nil && len(splunkLinks) > 0 {
		if err := d.Set("target_splunk", splunkLinks); err != nil {
			return err
		}
	}
	if len(appdLinks) > 0 {
		if err := d.Set("target_appd_url", appdLinks); err != nil {
			return err
		}
	}

	return nil
}

func dataLinkRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	dl, err := config.Client.GetDataLink(context.TODO(), d.Id())
	if err != nil {
		return err
	}

	return dataLinkAPIToTF(d, dl)
}

func dataLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload, err := getPayloadDataLink(d)
	if err != nil {
		return err
	}
	debugOutput, _ := json.Marshal(payload)
	log.Printf("[DEBUG] SignalFx: Update Data Link Payload: %s", string(debugOutput))

	dl, err := config.Client.UpdateDataLink(context.TODO(), d.Id(), payload)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] SignalFx: Update Data Link Response: %v", dl)

	d.SetId(dl.Id)
	return dataLinkAPIToTF(d, dl)
}

func dataLinkDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)

	return config.Client.DeleteDataLink(context.TODO(), d.Id())
}
