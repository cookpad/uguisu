import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.5-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.5",
      title: "CloudTrail configuration changes",
      description:
        "AWS CIS benchmark 3.5 recommend to ensure a log metric filter and alarm exist for CloudTrail configuration changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      CreateTrail: true,
      UpdateTrail: true,
      DeleteTrail: true,
      StartLogging: true,
      StopLogging: true,
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
