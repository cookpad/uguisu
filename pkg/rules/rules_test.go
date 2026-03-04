package rules

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/cookpad/uguisu/pkg/models"
	"github.com/stretchr/testify/assert"
)

func strp(s string) *string { return aws.String(s) }

// ── CIS 3.2 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_2(t *testing.T) {
	rule := newAwsCIS3_2()

	base := func() *models.CloudTrailRecord {
		return &models.CloudTrailRecord{
			EventName: "ConsoleLogin",
			AdditionalEventData: &models.CloudTrailAdditionalEventData{
				MFAUsed: "No",
			},
		}
	}

	t.Run("detects console login without MFA", func(t *testing.T) {
		assert.True(t, rule.Match(base()))
	})

	t.Run("no detection when MFA is used", func(t *testing.T) {
		r := base()
		r.AdditionalEventData.MFAUsed = "Yes"
		assert.False(t, rule.Match(r))
	})

	t.Run("no detection when SAML provider is set", func(t *testing.T) {
		r := base()
		r.AdditionalEventData.SamlProviderArn = strp("arn:aws:iam::123:saml-provider/okta")
		assert.False(t, rule.Match(r))
	})

	t.Run("no detection when session issuer type is Role", func(t *testing.T) {
		r := base()
		r.UserIdentity.SessionContext = &models.CloudTrailSessionContext{
			SessionIssuer: &models.CloudTrailSessionIssuer{Type: "Role"},
		}
		assert.False(t, rule.Match(r))
	})

	t.Run("detects when session context exists but issuer is not Role", func(t *testing.T) {
		r := base()
		r.UserIdentity.SessionContext = &models.CloudTrailSessionContext{
			SessionIssuer: &models.CloudTrailSessionIssuer{Type: "IAMUser"},
		}
		assert.True(t, rule.Match(r))
	})

	t.Run("detects when session context has nil issuer", func(t *testing.T) {
		r := base()
		r.UserIdentity.SessionContext = &models.CloudTrailSessionContext{}
		assert.True(t, rule.Match(r))
	})

	t.Run("no detection when event name is not ConsoleLogin", func(t *testing.T) {
		r := base()
		r.EventName = "AssumeRole"
		assert.False(t, rule.Match(r))
	})

	t.Run("no detection when AdditionalEventData is nil", func(t *testing.T) {
		r := base()
		r.AdditionalEventData = nil
		assert.False(t, rule.Match(r))
	})
}

// ── CIS 3.3 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_3(t *testing.T) {
	rule := newAwsCIS3_3()

	base := func() *models.CloudTrailRecord {
		return &models.CloudTrailRecord{
			UserIdentity: models.CloudTrailUserIdentity{Type: "Root"},
			EventType:    "AwsApiCall",
		}
	}

	t.Run("detects root account usage", func(t *testing.T) {
		assert.True(t, rule.Match(base()))
	})

	t.Run("no detection when identity type is not Root", func(t *testing.T) {
		r := base()
		r.UserIdentity.Type = "IAMUser"
		assert.False(t, rule.Match(r))
	})

	t.Run("no detection when InvokedBy is set", func(t *testing.T) {
		r := base()
		r.UserIdentity.InvokedBy = strp("cloudformation.amazonaws.com")
		assert.False(t, rule.Match(r))
	})

	t.Run("no detection when event type is AwsServiceEvent", func(t *testing.T) {
		r := base()
		r.EventType = "AwsServiceEvent"
		assert.False(t, rule.Match(r))
	})
}

// ── CIS 3.4 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_4(t *testing.T) {
	rule := newAwsCIS3_4()

	targetEvents := []string{
		"DeleteGroupPolicy", "DeleteRolePolicy", "DeleteUserPolicy",
		"PutGroupPolicy", "PutRolePolicy", "PutUserPolicy",
		"CreatePolicy", "DeletePolicy", "CreatePolicyVersion", "DeletePolicyVersion",
		"AttachRolePolicy", "DetachRolePolicy",
		"AttachUserPolicy", "DetachUserPolicy",
		"AttachGroupPolicy", "DetachGroupPolicy",
	}

	for _, event := range targetEvents {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeInstances"}))
	})
}

// ── CIS 3.5 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_5(t *testing.T) {
	rule := newAwsCIS3_5()

	for _, event := range []string{"CreateTrail", "UpdateTrail", "DeleteTrail", "StartLogging", "StopLogging"} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeTrails"}))
	})
}

// ── CIS 3.6 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_6(t *testing.T) {
	rule := newAwsCIS3_6()

	t.Run("detects failed console authentication", func(t *testing.T) {
		assert.True(t, rule.Match(&models.CloudTrailRecord{
			EventName:    "ConsoleLogin",
			ErrorMessage: strp("Failed authentication"),
		}))
	})

	t.Run("no detection when ErrorMessage is nil", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "ConsoleLogin"}))
	})

	t.Run("no detection when ErrorMessage differs", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventName:    "ConsoleLogin",
			ErrorMessage: strp("MFA required"),
		}))
	})

	t.Run("no detection when event name is not ConsoleLogin", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventName:    "AssumeRole",
			ErrorMessage: strp("Failed authentication"),
		}))
	})
}

// ── CIS 3.7 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_7(t *testing.T) {
	rule := newAwsCIS3_7()

	for _, event := range []string{"DisableKey", "ScheduleKeyDeletion"} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{
				EventSource: "kms.amazonaws.com",
				EventName:   event,
			}))
		})
	}

	t.Run("no detection for wrong event source", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource: "ec2.amazonaws.com",
			EventName:   "DisableKey",
		}))
	})

	t.Run("no detection for unrelated KMS event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource: "kms.amazonaws.com",
			EventName:   "DescribeKey",
		}))
	})
}

// ── CIS 3.8 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_8(t *testing.T) {
	rule := newAwsCIS3_8()

	for _, event := range []string{
		"PutBucketAcl", "PutBucketPolicy", "PutBucketCors",
		"PutBucketLifecycle", "PutBucketReplication",
		"DeleteBucketPolicy", "DeleteBucketCors",
		"DeleteBucketLifecycle", "DeleteBucketReplication",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "GetBucketPolicy"}))
	})
}

// ── CIS 3.9 ──────────────────────────────────────────────────────────────────

func TestAwsCIS3_9(t *testing.T) {
	rule := newAwsCIS3_9()

	for _, event := range []string{
		"StopConfigurationRecorder", "DeleteDeliveryChannel",
		"PutDeliveryChannel", "PutConfigurationRecorder",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeConfigurationRecorders"}))
	})
}

// ── CIS 3.10 ─────────────────────────────────────────────────────────────────

func TestAwsCIS3_10(t *testing.T) {
	rule := newAwsCIS3_10()

	for _, event := range []string{
		"AuthorizeSecurityGroupIngress", "AuthorizeSecurityGroupEgress",
		"RevokeSecurityGroupIngress", "RevokeSecurityGroupEgress",
		"CreateSecurityGroup", "DeleteSecurityGroup",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeSecurityGroups"}))
	})
}

// ── CIS 3.11 ─────────────────────────────────────────────────────────────────

func TestAwsCIS3_11(t *testing.T) {
	rule := newAwsCIS3_11()

	for _, event := range []string{
		"CreateNetworkAcl", "CreateNetworkAclEntry",
		"DeleteNetworkAcl", "DeleteNetworkAclEntry",
		"ReplaceNetworkAclEntry", "ReplaceNetworkAclAssociation",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeNetworkAcls"}))
	})
}

// ── CIS 3.12 ─────────────────────────────────────────────────────────────────

func TestAwsCIS3_12(t *testing.T) {
	rule := newAwsCIS3_12()

	for _, event := range []string{
		"CreateCustomerGateway", "DeleteCustomerGateway",
		"AttachInternetGateway", "CreateInternetGateway",
		"DeleteInternetGateway", "DetachInternetGateway",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeInternetGateways"}))
	})
}

// ── CIS 3.13 ─────────────────────────────────────────────────────────────────

func TestAwsCIS3_13(t *testing.T) {
	rule := newAwsCIS3_13()

	for _, event := range []string{
		"CreateRoute", "CreateRouteTable",
		"ReplaceRoute", "ReplaceRouteTableAssociation",
		"DeleteRouteTable", "DeleteRoute", "DisassociateRouteTable",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeRouteTables"}))
	})
}

// ── CIS 3.14 ─────────────────────────────────────────────────────────────────

func TestAwsCIS3_14(t *testing.T) {
	rule := newAwsCIS3_14()

	for _, event := range []string{
		"CreateVpc", "DeleteVpc", "ModifyVpcAttribute",
		"AcceptVpcPeeringConnection", "CreateVpcPeeringConnection",
		"DeleteVpcPeeringConnection", "RejectVpcPeeringConnection",
		"AttachClassicLinkVpc", "DetachClassicLinkVpc",
		"DisableVpcClassicLink", "EnableVpcClassicLink",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DescribeVpcs"}))
	})
}

// ── Life Events ───────────────────────────────────────────────────────────────

func TestLifeEventACM(t *testing.T) {
	rule := newLifeEventACM()

	for _, event := range []string{"ExportCertificate", "ImportCertificate", "RenewCertificate", "DeleteCertificate"} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "ListCertificates"}))
	})
}

func TestLifeEventEC2(t *testing.T) {
	rule := newLifeEventEC2()

	for _, event := range []string{"RunInstances", "TerminateInstances"} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{
				EventSource: "ec2.amazonaws.com",
				EventName:   event,
			}))
		})
	}

	t.Run("no detection for wrong event source", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource: "rds.amazonaws.com",
			EventName:   "RunInstances",
		}))
	})

	t.Run("no detection when source IP is autoscaling", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource:     "ec2.amazonaws.com",
			EventName:       "RunInstances",
			SourceIPAddress: "autoscaling.amazonaws.com",
		}))
	})

	t.Run("no detection when source IP is batch", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource:     "ec2.amazonaws.com",
			EventName:       "RunInstances",
			SourceIPAddress: "batch.amazonaws.com",
		}))
	})

	t.Run("no detection when source IP is ecs-compute", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource:     "ec2.amazonaws.com",
			EventName:       "TerminateInstances",
			SourceIPAddress: "ecs-compute.amazonaws.com",
		}))
	})

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource: "ec2.amazonaws.com",
			EventName:   "DescribeInstances",
		}))
	})
}

func TestLifeEventRDS(t *testing.T) {
	rule := newLifeEventRDS()

	for _, event := range []string{"CreateDBInstance", "DeleteDBInstance"} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{
				EventSource: "rds.amazonaws.com",
				EventName:   event,
			}))
		})
	}

	t.Run("no detection for wrong event source", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource: "ec2.amazonaws.com",
			EventName:   "CreateDBInstance",
		}))
	})

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{
			EventSource: "rds.amazonaws.com",
			EventName:   "DescribeDBInstances",
		}))
	})
}

func TestLifeEventVPC(t *testing.T) {
	rule := newLifeEventVPC()

	t.Run("detects CreateVpc", func(t *testing.T) {
		assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: "CreateVpc"}))
	})

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "DeleteVpc"}))
	})
}

func TestLifeEventOrg(t *testing.T) {
	rule := newLifeEventOrg()

	for _, event := range []string{
		"CreateAccount", "CreateOrganization", "DeleteOrganization",
		"AcceptHandshake", "LeaveOrganization",
	} {
		event := event
		t.Run("detects "+event, func(t *testing.T) {
			assert.True(t, rule.Match(&models.CloudTrailRecord{EventName: event}))
		})
	}

	t.Run("no detection for unrelated event", func(t *testing.T) {
		assert.False(t, rule.Match(&models.CloudTrailRecord{EventName: "ListAccounts"}))
	})
}
