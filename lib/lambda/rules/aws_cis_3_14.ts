import { cloudTrailRecord, uguisuRule, detection } from "../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.14-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.14",
      title: "VPC changes",
      description:
        "AWS CIS benchmark 3.14 recommend to ensure a log metric filter and alarm exist for VPC changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      CreateVpc: true,
      DeleteVpc: true,
      ModifyVpcAttribute: true,
      AcceptVpcPeeringConnection: true,
      CreateVpcPeeringConnection: true,
      DeleteVpcPeeringConnection: true,
      RejectVpcPeeringConnection: true,
      AttachClassicLinkVpc: true,
      DetachClassicLinkVpc: true,
      DisableVpcClassicLink: true,
      EnableVpcClassicLink: true,
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
