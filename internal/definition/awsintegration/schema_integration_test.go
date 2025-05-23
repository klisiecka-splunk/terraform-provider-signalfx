// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsintegration

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/signalfx/signalfx-go/integration"
	"github.com/stretchr/testify/assert"
)

func TestNewIntegrationSchema(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, newIntegrationSchema(), "Must return a valid schema value")
}

func TestIntegrationDecode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		data   map[string]any
		expect *integration.AwsCloudWatchIntegration
		errVal string
	}{
		{
			name:   "empty data",
			data:   make(map[string]any),
			expect: nil,
			errVal: "requires either `external_id` or `token` and `key`",
		},
		{
			name: "external auth set",
			data: map[string]any{
				"external_id": "my-id",
				"role_arn":    "....",
			},
			expect: &integration.AwsCloudWatchIntegration{
				Type:                     "AWSCloudWatch",
				AuthMethod:               integration.EXTERNAL_ID,
				ExternalId:               "my-id",
				RoleArn:                  "....",
				PollRate:                 300000,
				CustomNamespaceSyncRules: []*integration.AwsCustomNameSpaceSyncRule{},
				NamespaceSyncRules:       []*integration.AwsNameSpaceSyncRule{},
				Services:                 []integration.AwsService{},
				Regions:                  []string{},
			},
			errVal: "",
		},
		{
			name: "token auth set",
			data: map[string]any{
				"token": "my-token",
				"key":   "my-key",
			},
			expect: &integration.AwsCloudWatchIntegration{
				Type:                     "AWSCloudWatch",
				AuthMethod:               integration.SECURITY_TOKEN,
				Token:                    "my-token",
				Key:                      "my-key",
				PollRate:                 300000,
				CustomNamespaceSyncRules: []*integration.AwsCustomNameSpaceSyncRule{},
				NamespaceSyncRules:       []*integration.AwsNameSpaceSyncRule{},
				Services:                 []integration.AwsService{},
				Regions:                  []string{},
			},
			errVal: "",
		},
		{
			name: "min required values",
			data: map[string]any{
				"token":   "my-token",
				"key":     "my-key",
				"regions": []any{"us-east-1"},
			},
			expect: &integration.AwsCloudWatchIntegration{
				Type:                     "AWSCloudWatch",
				AuthMethod:               integration.SECURITY_TOKEN,
				Token:                    "my-token",
				Key:                      "my-key",
				Regions:                  []string{"us-east-1"},
				PollRate:                 300000,
				CustomNamespaceSyncRules: []*integration.AwsCustomNameSpaceSyncRule{},
				NamespaceSyncRules:       []*integration.AwsNameSpaceSyncRule{},
				Services:                 []integration.AwsService{},
			},
			errVal: "",
		},
		{
			name: "syncing specific metrics",
			data: map[string]any{
				"token":   "my-token",
				"key":     "my-key",
				"regions": []any{"us-east-1"},
				"metric_stats_to_sync": []any{
					map[string]any{
						"namespace": "aws/kinesis",
						"metric":    "my-awesome-metric",
						"stats":     []any{"mean"},
					},
				},
			},
			expect: &integration.AwsCloudWatchIntegration{
				Type:       "AWSCloudWatch",
				AuthMethod: integration.SECURITY_TOKEN,
				Token:      "my-token",
				Key:        "my-key",
				Regions:    []string{"us-east-1"},
				PollRate:   300000,
				MetricStatsToSync: map[string]map[string][]string{
					"aws/kinesis": {
						"my-awesome-metric": {"mean"},
					},
				},
				CustomNamespaceSyncRules: []*integration.AwsCustomNameSpaceSyncRule{},
				NamespaceSyncRules:       []*integration.AwsNameSpaceSyncRule{},
				Services:                 []integration.AwsService{},
			},
			errVal: "",
		},
		{
			name: "all fields",
			data: map[string]any{
				"token":                   "my-token",
				"key":                     "my-key",
				"regions":                 []any{"us-east-1"},
				"use_metric_streams_sync": true,
				"enable_log_sync":         true,
				"poll_rate":               10_000,
				"named_token":             "my-awesome-token",
				"metric_stats_to_sync": []any{
					map[string]any{
						"namespace": "aws/kinesis",
						"metric":    "my-awesome-metric",
						"stats":     []any{"mean"},
					},
				},
			},
			expect: &integration.AwsCloudWatchIntegration{
				Type:                   "AWSCloudWatch",
				AuthMethod:             integration.SECURITY_TOKEN,
				Token:                  "my-token",
				Key:                    "my-key",
				Regions:                []string{"us-east-1"},
				PollRate:               10000000,
				NamedToken:             "my-awesome-token",
				MetricStreamsSyncState: "ENABLED",
				MetricStatsToSync: map[string]map[string][]string{
					"aws/kinesis": {
						"my-awesome-metric": {"mean"},
					},
				},
				CustomNamespaceSyncRules: []*integration.AwsCustomNameSpaceSyncRule{},
				NamespaceSyncRules:       []*integration.AwsNameSpaceSyncRule{},
				Services:                 []integration.AwsService{},
			},
			errVal: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := decodeTerraform(
				schema.TestResourceDataRaw(t, newIntegrationSchema(), tc.data),
			)
			assert.Equal(t, tc.expect, actual, "Must match the expected value")
			if tc.errVal != "" {
				assert.EqualError(t, err, tc.errVal, "Must match the expected error value")
			} else {
				assert.NoError(t, err, "Must not error creating type")
			}
		})
	}
}

func TestEncodeTerraformIntegration(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		input  *integration.AwsCloudWatchIntegration
		errVal string
	}{
		{
			name:   "empty integration",
			input:  &integration.AwsCloudWatchIntegration{},
			errVal: "",
		},
		{
			name: "token set",
			input: &integration.AwsCloudWatchIntegration{
				Token: "my-token",
			},
			errVal: "",
		},
		{
			name: "regions set",
			input: &integration.AwsCloudWatchIntegration{
				Regions: []string{
					"us-east-1",
					"us-east-2",
				},
			},
			errVal: "",
		},
		{
			name: "custom namespace sync rules",
			input: &integration.AwsCloudWatchIntegration{
				CustomNamespaceSyncRules: []*integration.AwsCustomNameSpaceSyncRule{
					{
						DefaultAction: integration.INCLUDE,
						Namespace:     "AWS/Kinesis",
						Filter: &integration.AwsSyncRuleFilter{
							Source: "hope",
							Action: integration.INCLUDE,
						},
					},
				},
			},
		},
		{
			name: "custom namespaces",
			input: &integration.AwsCloudWatchIntegration{
				CustomCloudWatchNamespaces: "namespace-1,namespace-2,namespace-3",
			},
			errVal: "",
		},
		{
			name: "services",
			input: &integration.AwsCloudWatchIntegration{
				Services: []integration.AwsService{
					"my-service-01",
					"my-service-02",
				},
			},
			errVal: "",
		},
		{
			name: "namespace rules",
			input: &integration.AwsCloudWatchIntegration{
				NamespaceSyncRules: []*integration.AwsNameSpaceSyncRule{
					{
						DefaultAction: integration.INCLUDE,
						Namespace:     "AWS/Kinesis",
						Filter: &integration.AwsSyncRuleFilter{
							Action: integration.INCLUDE,
							Source: "hope",
						},
					},
				},
			},
			errVal: "",
		},
		{
			name: "metric stats",
			input: &integration.AwsCloudWatchIntegration{
				MetricStatsToSync: map[string]map[string][]string{
					"AWS/SQS": {
						"ApproximateAgeOfOldestMessage": {"mean"},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := encodeTerraform(tc.input, schema.TestResourceDataRaw(t, newIntegrationSchema(), map[string]any{}))
			if tc.errVal != "" {
				assert.EqualError(t, err, tc.errVal, "Must match the expected value")
			} else {
				assert.NoError(t, err, "Must not return an error")
			}
		})
	}
}
