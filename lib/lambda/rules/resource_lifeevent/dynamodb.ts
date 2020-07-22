import { cloudTrailRecord, uguisuRule, detection } from "../../models";

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "resource_lifeevent_dynamodb",
      title: "DynamoDB Life Event",
      description: "Monitoring events of DynamoDB creation and destruction.",
      severity: "low",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents = new Set(["CreateTable", "DeleteTable"]);

    if (
      record.eventSource === "dynamodb.amazonaws.com" &&
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
