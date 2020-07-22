import { cloudTrailRecord, uguisuRule, detection } from "../../models";

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "resource_lifeevent_ec2",
      title: "EC2 Life Event",
      description: "Monitoring events of EC2 creation and destruction.",
      severity: "low",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents = new Set(["RunInstances", "TerminateInstances"]);

    // Ignore events by autoscaling or batch
    if (
      record.sourceIPAddress === "autoscaling.amazonaws.com" ||
      record.sourceIPAddress === "batch.amazonaws.com"
    ) {
      return null;
    }

    if (
      record.eventSource === "ec2.amazonaws.com" &&
      targetEvents.has(record.eventName)
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
