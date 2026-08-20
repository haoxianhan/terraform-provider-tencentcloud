package cls

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestClsConfigExtraExtractRuleKeysPreserveOrder(t *testing.T) {
	resource := ResourceTencentCloudClsConfigExtra()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"extract_rule": []interface{}{
			map[string]interface{}{
				"keys": []interface{}{"first", "second"},
			},
		},
	})

	extractRules := data.Get("extract_rule").([]interface{})
	keysValue := extractRules[0].(map[string]interface{})["keys"]
	keys, ok := keysValue.([]interface{})
	if !ok {
		t.Fatalf("extract_rule.keys must preserve order as a list, got %T", keysValue)
	}

	want := []interface{}{"first", "second"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("extract_rule.keys = %#v, want %#v", keys, want)
	}
}

func TestClsConfigExtraStateUpgradeV0(t *testing.T) {
	resourceDef := ResourceTencentCloudClsConfigExtra()
	if resourceDef.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", resourceDef.SchemaVersion)
	}
	if len(resourceDef.StateUpgraders) != 1 {
		t.Fatalf("StateUpgraders length = %d, want 1", len(resourceDef.StateUpgraders))
	}

	upgrader := resourceDef.StateUpgraders[0]
	legacyKeysType := upgrader.Type.AttributeType("extract_rule").ElementType().AttributeType("keys")
	if !legacyKeysType.IsSetType() {
		t.Fatalf("legacy extract_rule.keys type = %s, want set", legacyKeysType.FriendlyName())
	}

	currentType := resourceDef.CoreConfigSchema().ImpliedType()
	currentKeysType := currentType.AttributeType("extract_rule").ElementType().AttributeType("keys")
	if !currentKeysType.IsListType() {
		t.Fatalf("current extract_rule.keys type = %s, want list", currentKeysType.FriendlyName())
	}

	rawState := map[string]interface{}{
		"id": "config-extra-id",
		"extract_rule": []interface{}{
			map[string]interface{}{
				"keys": []interface{}{"first", "second"},
			},
		},
	}
	upgradedState, err := upgrader.Upgrade(t.Context(), rawState, nil)
	if err != nil {
		t.Fatalf("state upgrade failed: %v", err)
	}

	data := schema.TestResourceDataRaw(t, resourceDef.Schema, upgradedState)
	extractRules := data.Get("extract_rule").([]interface{})
	got := extractRules[0].(map[string]interface{})["keys"].([]interface{})
	want := []interface{}{"first", "second"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("upgraded extract_rule.keys = %#v, want %#v", got, want)
	}

	if err := resourceDef.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource validation failed: %v", err)
	}
}
