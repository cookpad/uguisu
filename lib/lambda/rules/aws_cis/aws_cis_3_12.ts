import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.12-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.12",
      title: "Changes to network gateways",
      description:
        "AWS CIS benchmark 3.12 recommend to ensure a log metric filter and alarm exist for Changes to network gateways",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      CreateCustomerGateway: true,
      DeleteCustomerGateway: true,
      AttachInternetGateway: true,
      CreateInternetGateway: true,
      DeleteInternetGateway: true,
      DetachInternetGateway: true,
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
