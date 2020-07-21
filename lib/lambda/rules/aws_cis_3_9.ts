import { cloudTrailRecord, uguisuRule, detection } from "../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.9-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.9",
      title: "AWS Config configuration changes",
      description:
        "AWS CIS benchmark 3.9 recommend to ensure a log metric filter and alarm exist for AWS Config configuration changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      StopConfigurationRecorder: true,
      DeleteDeliveryChannel: true,
      PutDeliveryChannel: true,
      PutConfigurationRecorder: true,
    };

    if (
      record.eventSource === "config.amazonaws.com" &&
      targetEvents[record.eventName] === true
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
