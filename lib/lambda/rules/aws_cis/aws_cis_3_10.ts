import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.10-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.10",
      title: "Security group changes",
      description:
        "AWS CIS benchmark 3.10 recommend to ensure a log metric filter and alarm exist for security group changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      AuthorizeSecurityGroupIngress: true,
      AuthorizeSecurityGroupEgress: true,
      RevokeSecurityGroupIngress: true,
      RevokeSecurityGroupEgress: true,
      CreateSecurityGroup: true,
      DeleteSecurityGroup: true,
    };

    if (targetEvents[record.eventName] === true) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
