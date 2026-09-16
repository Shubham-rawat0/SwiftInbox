export interface Message {
    id: string;
    sender: string;
    subject: string;
    preview: string;
    createdAt: string;
}

export interface MessageListResponse {
    address: string;
    messages: Message[];
    messageCount: number;
}

export interface MessageAttachment {
    filename: string;
    contentType: string;
    size: number;
    contentId: string;
    index: number;
}

export interface ParsedMessageData {
    subject: string;
    from: string;
    text: string;
    html: string;
    attachments: MessageAttachment[];
}

export interface MessageDetail {
    id: string;
    from: string;
    subject: string;
    body: string;
    mailbox: string;
    createdAt: string;
    parsedData: ParsedMessageData;
}

export interface MailboxResponse {
    address: string;
    createdAt: string;
    expiresAt: string;
    createdBy: NullableUUID;
}

export interface MailboxDeleteResponse {
    message: string;
    address: string;
    id: string;
}

export interface CreateDeveloperResponse {
    ID: string;
    Name: string;
    Email: string;
}

export interface DeveloperSignInResponse {
    id: string;
    name: string;
    email: string;
}

export interface DeveloperDetailsResponse {
    Name: string;
    Email: string;
    ApiQuota: number;
    MailboxQuota: number;
    MessageQuota: number;
    CreatedAt: NullableTime;
    ApiRequests: number;
    MailboxRequests: number;
    MessagesRequests: number;
}

export interface CreateApiKeyResponse {
    id: string;
    name: string;
    api_key: string;
}

export interface RevokeApiKeyResponse {
    ID: string;
    LastUsedAt: NullableTime;
    RevokedAt: NullableTime;
}

export interface ApiKeyUsageResponse {
    ID: string;
    Name: string;
    LastUsedAt: NullableTime;
    RevokedAt: NullableTime;
    CreatedAt: NullableTime;
    ApiRequests: number;
    MailboxRequests: number;
    MessageRequests: number;
}

export interface WebhookResponse {
    id: string;
    developer_id: string;
    url: string;
    is_active: boolean;
    events: string[];
}

export interface CreateWebhookResponse extends WebhookResponse {
    secret: string;
    mailbox_ids: string[];
}

export interface WebhookMailboxResponse {
    message: string;
    mailbox_id: string;
}

export interface WebhookDeleteResponse {
    message: string;
    id: string;
}

export interface WebhookTestResponse {
    success: boolean;
    statusCode: number;
}

export interface WebhookDeadLetterResponse {
    id: string;
    webhook_id: string;
    mailbox_id: string | null;
    message_id: string | null;
    event: string;
    url: string;
    reason: string;
    attempts: number;
    created_at: string;
}

export interface ApiErrorResponse {
    error: string;
}

export type HealthResponse = "healthy";

export interface MessageResponse {
    message: string;
}

export interface NullableTime {
    Time: string;
    Valid: boolean;
}

export interface NullableUUID {
    UUID: string;
    Valid: boolean;
}