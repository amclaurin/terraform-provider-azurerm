package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmServiceBusTopicAuthorizationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmServiceBusTopicAuthorizationRuleCreateUpdate,
		Read:   resourceArmServiceBusTopicAuthorizationRuleRead,
		Update: resourceArmServiceBusTopicAuthorizationRuleCreateUpdate,
		Delete: resourceArmServiceBusTopicAuthorizationRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateServiceBusAuthorizationRuleName(),
			},

			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateServiceBusNamespaceName(),
			},

			"topic_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateServiceBusTopicName(),
			},

			"resource_group_name": resourceGroupNameSchema(),

			"listen": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"send": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"manage": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"primary_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_connection_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_connection_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},

		CustomizeDiff: azure.ServiceBusAuthorizationRuleCustomizeDiff,
	}
}

func resourceArmServiceBusTopicAuthorizationRuleCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).serviceBusTopicsClient
	ctx := meta.(*ArmClient).StopContext
	log.Printf("[INFO] preparing arguments for AzureRM ServiceBus Topic Authorization Rule creation.")

	name := d.Get("name").(string)
	namespaceName := d.Get("namespace_name").(string)
	topicName := d.Get("topic_name").(string)
	resGroup := d.Get("resource_group_name").(string)

	parameters := servicebus.SBAuthorizationRule{
		Name: &name,
		SBAuthorizationRuleProperties: &servicebus.SBAuthorizationRuleProperties{
			Rights: azure.ExpandServiceBusAuthorizationRuleRights(d),
		},
	}

	_, err := client.CreateOrUpdateAuthorizationRule(ctx, resGroup, namespaceName, topicName, name, parameters)
	if err != nil {
		return err
	}

	read, err := client.GetAuthorizationRule(ctx, resGroup, namespaceName, topicName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read ServiceBus Topic Authorization Rule %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmServiceBusTopicAuthorizationRuleRead(d, meta)
}

func resourceArmServiceBusTopicAuthorizationRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).serviceBusTopicsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	namespaceName := id.Path["namespaces"]
	topicName := id.Path["topics"]
	name := id.Path["authorizationRules"]

	resp, err := client.GetAuthorizationRule(ctx, resGroup, namespaceName, topicName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure ServiceBus Topic Authorization Rule %s: %+v", name, err)
	}

	d.Set("name", name)
	d.Set("topic_name", topicName)
	d.Set("namespace_name", namespaceName)
	d.Set("resource_group_name", resGroup)

	if properties := resp.SBAuthorizationRuleProperties; properties != nil {
		listen, send, manage := azure.FlattenServiceBusAuthorizationRuleRights(properties.Rights)
		d.Set("listen", listen)
		d.Set("send", send)
		d.Set("manage", manage)
	}

	keysResp, err := client.ListKeys(ctx, resGroup, namespaceName, topicName, name)
	if err != nil {
		return fmt.Errorf("Error making Read request on Azure ServiceBus Topic Authorization Rule List Keys %s: %+v", name, err)
	}

	d.Set("primary_key", keysResp.PrimaryKey)
	d.Set("primary_connection_string", keysResp.PrimaryConnectionString)
	d.Set("secondary_key", keysResp.SecondaryKey)
	d.Set("secondary_connection_string", keysResp.SecondaryConnectionString)

	return nil
}

func resourceArmServiceBusTopicAuthorizationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).serviceBusTopicsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	namespaceName := id.Path["namespaces"]
	topicName := id.Path["topics"]
	name := id.Path["authorizationRules"]

	if _, err = client.DeleteAuthorizationRule(ctx, resGroup, namespaceName, topicName, name); err != nil {
		return fmt.Errorf("Error issuing Azure ARM delete request of ServiceBus Topic Authorization Rule %q (Resource Group %q): %+v", name, resGroup, err)
	}

	return nil
}
