import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.4-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.4",
      title: "IAM policy changes",
      description:
        "AWS CIS benchmark 3.4 recommend to ensure a log metric filter and alarm exist for IAM policy changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      DeleteGroupPolicy: true,
      DeleteRolePolicy: true,
      DeleteUserPolicy: true,
      PutGroupPolicy: true,
      PutRolePolicy: true,
      PutUserPolicy: true,
      CreatePolicy: true,
      DeletePolicy: true,
      CreatePolicyVersion: true,
      DeletePolicyVersion: true,
      AttachRolePolicy: true,
      DetachRolePolicy: true,
      AttachUserPolicy: true,
      DetachUserPolicy: true,
      AttachGroupPolicy: true,
      DetachGroupPolicy: true,
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
