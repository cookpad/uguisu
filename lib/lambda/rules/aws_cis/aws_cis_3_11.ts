import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.11-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.11",
      title: "Network Access Control Lists (NACL)",
      description:
        "AWS CIS benchmark 3.11 recommend to ensure a log metric filter and alarm exist for Network Access Control Lists (NACL)",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      CreateNetworkAcl: true,
      CreateNetworkAclEntry: true,
      DeleteNetworkAcl: true,
      DeleteNetworkAclEntry: true,
      ReplaceNetworkAclEntry: true,
      ReplaceNetworkAclAssociation: true,
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
