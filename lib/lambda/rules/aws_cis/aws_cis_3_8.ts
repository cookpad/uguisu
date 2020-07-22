import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.8-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.8",
      title: "S3 bucket policy changes",
      description:
        "AWS CIS benchmark 3.8 recommend to ensure a log metric filter and alarm exist for S3 bucket policy changes",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents: { [key: string]: boolean } = {
      PutBucketAcl: true,
      PutBucketPolicy: true,
      PutBucketCors: true,
      PutBucketLifecycle: true,
      PutBucketReplication: true,
      DeleteBucketPolicy: true,
      DeleteBucketCors: true,
      DeleteBucketLifecycle: true,
      DeleteBucketReplication: true,
    };

    if (
      record.eventSource === "s3.amazonaws.com" &&
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
