export interface cloudTrailEvent {
  Records: Array<cloudTrailRecord>;
}

export interface cloudTrailRecord {
  eventVersion: string;
  eventTime: string;
  eventSource: string;
  eventName: string;
  awsRegion: string;
  requestID: string;
  eventID: string;
  eventType: string;
  recipientAccountId: string;

  errorCode?: string;
  errorMessage?: string;

  sourceIPAddress?: string;
  userIdentity?: any;
  userAgent?: string;
  requestParameters?: any;
  responseElements?: any;
}

export interface detection {
  rule: uguisuRule;
  event: cloudTrailRecord;
}

export interface ruleParameters {
  // Duplicated id is not allowed for each other rules.
  id: string;
  // title is short description for human.
  title: string;
  // description is more detail (What is monitored, reason of detection, etc.)
  description: string;
  // severity is
  severity: "high" | "medium" | "low";
}

export abstract class uguisuRule {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly severity: "high" | "medium" | "low";

  constructor(params: ruleParameters) {
    this.id = params.id;
    this.title = params.title;
    this.description = params.description;
    this.severity = params.severity;
  }
  abstract detect(record: cloudTrailRecord): detection | null;
}
