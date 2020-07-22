import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.13-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.13",
      title: "Route table changes",
      description:
        "AWS CIS benchmark 3.13 recommend to ensure a log metric filter and alarm exist for route table changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      CreateRoute: true,
      CreateRouteTable: true,
      ReplaceRoute: true,
      ReplaceRouteTableAssociation: true,
      DeleteRouteTable: true,
      DeleteRoute: true,
      DisassociateRouteTable: true,
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
