const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:3001';

import type { AttachmentResult, CreateDeveloperRequest, DeveloperDetailsResponse, DeveloperSignInRequest, DeveloperSignInResponse, MailboxDeleteResponse, MailboxResponse, Message, MessageDetail } from './types';
import { toast } from 'sonner';

/** Error thrown by the API layer that carries the HTTP status when available. */
export class ApiError extends Error {
    readonly status?: number;

    constructor(message: string, status?: number) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
    }
}

export function isApiError(error: unknown): error is ApiError {
    return error instanceof ApiError;
}

const requestCache = new Map<string, { data: unknown; timestamp: number }>();
const pendingRequests = new Map<string, Promise<unknown>>();
const CACHE_DURATION = 6000;

const getCachedData = <T>(key: string): T | null => {
    const cached = requestCache.get(key);
    if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
        console.log("cached hit for", key);
        return cached.data as T;
    }
    return null;
};

const setCachedData = <T>(key: string, data: T) => {
    requestCache.set(key, { data, timestamp: Date.now() });
};

export function clearCache() {
    requestCache.clear()
    pendingRequests.clear()
    console.log("deleted cached data")
}

export function clearCacheForAddress(address: string) {
    const username = address.split('@')[0];
    const keys = [`messages-${address}`, `mailbox-${username}`];
    keys.forEach(key => {
        requestCache.delete(key);
        pendingRequests.delete(key);
    });
    console.log(`Cache cleared for address: ${address}`);
}

function deduplicate<T>(key: string, request: () => Promise<T>): Promise<T> {
    const existing = pendingRequests.get(key);
    if (existing) {
        return existing as Promise<T>;
    }

    //finally return promise(api result) and does cleanup
    const promise = request().finally(() => {
        pendingRequests.delete(key);
    });

    pendingRequests.set(key, promise);
    return promise;
}

export async function createCustomMailbox(username: string, auth: boolean, expiresAt?: Date): Promise<{ address: string; createdAt: string; expiresAt: string | null; }> {
    const cacheKey = `mailbox-${username}`;

    const cached = getCachedData<{
        address: string;
        createdAt: string;
        expiresAt: string | null;
    }>(cacheKey);

    if (cached) {
        return cached;
    }

    return deduplicate(cacheKey, async () => {
        try {
            const cacheBuster = `?_=${Date.now()}`;

            const endpoint = auth
                ? `${API_BASE}/api/dev/mailboxes/custom${cacheBuster}`
                : `${API_BASE}/api/mailboxes/custom${cacheBuster}`;

            const headers: HeadersInit = {
                'Content-Type': 'application/json',
            };

            const options: RequestInit = {
                method: 'POST',
                headers,
                body: JSON.stringify({ username, expiresAt: expiresAt ? expiresAt.toISOString() : undefined }),
                cache: 'no-store',
            };

            if (auth) {
                options.credentials = 'include';
            }

            const response = await fetch(endpoint, options);
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));

                if (response.status === 429) {
                    console.log('Rate limit hit on mailbox creation');

                    toast.error('Rate limit exceeded', {
                        description:
                            errorData.error ||
                            'Too many mailboxes created. Please try again after 1 hour.',
                    });

                    throw new ApiError('Rate limit exceeded', 429);
                }

                throw new ApiError(
                    errorData.error || `Failed to create custom mailbox: ${response.status} ${response.statusText}`,
                    response.status,
                );
            }

            const data = await response.json();

            const result = {
                address: data.address,
                createdAt: data.createdAt,
                expiresAt: data.expiresAt,
            };

            setCachedData(cacheKey, result);

            return result;
        } catch (error) {
            console.error('Error creating custom mailbox:', error);
            throw error;
        }
    });
}

export async function fetchMessages(address: string, forceRefresh = false, auth = false): Promise<{ messages: Message[] }> {
    const cacheKey = `messages-${address}`;

    if (!forceRefresh) {
        const cached = getCachedData<{ messages: Message[] }>(cacheKey);
        if (cached) {
            return cached;
        }
    }

    return deduplicate(cacheKey, async () => {
        try {
            const cacheBuster = forceRefresh ? `?_=${Date.now()}` : '';
            const fetchOptions: RequestInit = {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                cache: 'no-store',
            };

            if (auth) {
                fetchOptions.credentials = 'include';
            }

            const response = await fetch(`${API_BASE}/api/mailboxes/${encodeURIComponent(address)}/message${cacheBuster}`, fetchOptions);

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));

                if (response.status === 429) {
                    console.log('Rate limit hit on message fetch');
                    toast.error('Rate limit exceeded', {
                        description: errorData.error || 'Too many requests. Please slow down.',
                    });

                    throw new ApiError('Rate limit exceeded', 429);
                }

                throw new ApiError(errorData.error || `Failed to fetch messages: ${response.status} ${response.statusText}`, response.status);
            }

            const result = await response.json();
            setCachedData(cacheKey, result);
            return result;
        } catch (error) {
            console.error('Error fetching messages:', error);
            throw error;
        }
    });
}

export async function fetchMessage(messageId: string, auth = false): Promise<MessageDetail> {
    const cacheKey = `message-${messageId}`;

    const cached = getCachedData<MessageDetail>(cacheKey);
    if (cached) {
        return cached;
    }

    return deduplicate(cacheKey, async () => {
        try {
            const fetchOptions: RequestInit = {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            };

            if (auth) {
                fetchOptions.credentials = 'include';
            }

            const response = await fetch(`${API_BASE}/api/message/${messageId}`, fetchOptions);

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));

                if (response.status === 429) {
                    console.log('Rate limit hit on individual message fetch');
                    toast.error('Rate limit exceeded', {
                        description: errorData.error || 'Too many requests. Please wait a moment.',
                    });

                    await new Promise(resolve => setTimeout(resolve, 2000));
                    const retryResponse = await fetch(`${API_BASE}/api/message/${messageId}`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        ...(auth ? { credentials: 'include' as RequestCredentials } : {}),
                    });

                    if (retryResponse.ok) {
                        const result = await retryResponse.json();
                        setCachedData(cacheKey, result);
                        return result;
                    }
                }

                throw new ApiError(errorData.error || `Failed to fetch message: ${response.status} ${response.statusText}`, response.status);
            }

            const result = await response.json();
            setCachedData(cacheKey, result);
            return result;
        } catch (error) {
            console.error('Error fetching message:', error);
            throw error;
        }
    });
}

export async function fetchAttachment(messageId: string, index: number, auth = false): Promise<AttachmentResult> {
    const fetchOptions: RequestInit = {
        method: 'GET',
        cache: 'no-store',
    };

    if (auth) {
        fetchOptions.credentials = 'include';
    }

    const response = await fetch(`${API_BASE}/api/message/${messageId}/attachment/${index}`, fetchOptions);

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));

        if (response.status === 429) {
            console.log('Rate limit hit on attachment fetch');
            toast.error('Rate limit exceeded', {
                description: errorData.error || 'Too many requests. Please wait a moment.',
            });

            throw new ApiError('Rate limit exceeded', 429);
        }

        throw new ApiError(errorData.error || `Failed to fetch attachment: ${response.status} ${response.statusText}`, response.status);
    }

    const contentType = response.headers.get('Content-Type') || 'application/octet-stream';
    const disposition = response.headers.get('Content-Disposition');
    const filename = disposition?.match(/filename="?([^";]+)"?/i)?.[1] ?? null;

    const blob = await response.blob();
    const byteLength = blob.size;
    return {
        blob,
        contentType,
        filename,
        size: Number(response.headers.get('Content-Length')) || byteLength,
    };
}

async function getApiError(response: Response, fallback: string): Promise<ApiError> {
    const data = await response.json().catch(() => ({}));
    return new ApiError(data.error || `${fallback}: ${response.status} ${response.statusText}`, response.status);
}

export async function createDeveloper(developer: CreateDeveloperRequest): Promise<void> {
    const response = await fetch(`${API_BASE}/api/dev/create`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(developer),
    });

    if (!response.ok) {
        throw await getApiError(response, 'Unable to create developer account');
    }
}

export async function signInDeveloper(credentials: DeveloperSignInRequest): Promise<DeveloperSignInResponse> {
    const response = await fetch(`${API_BASE}/api/dev/signin`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(credentials),
    });

    if (!response.ok) {
        throw await getApiError(response, 'Unable to sign in');
    }
    return response.json();
}

export async function signOutDeveloper(): Promise<void> {
    const response = await fetch(`${API_BASE}/api/dev/signout`, {
        method: 'POST',
        credentials: 'include',
    });

    if (!response.ok) {
        throw await getApiError(response, 'Unable to sign out');
    }
}

export async function getDeveloper(): Promise<DeveloperDetailsResponse> {
    const developerId = localStorage.getItem("developer_id");
    if (!developerId || developerId.trim().length === 0) {
        throw new ApiError("Please sign in again.", 401);
    }

    const response = await fetch(`${API_BASE}/api/dev/${encodeURIComponent(developerId)}`, {
        method: "GET",
        credentials: "include",
    });

    if (!response.ok) {
        throw await getApiError(response, 'Unable to get developer details');
    }

    return response.json();
}

export async function listDeveloperMailboxes(forceRefresh = false): Promise<MailboxResponse[]> {
    const cacheKey = 'dev-mailboxes';

    if (!forceRefresh) {
        const cached = getCachedData<MailboxResponse[]>(cacheKey);
        if (cached) {
            return cached;
        }
    }

    return deduplicate(cacheKey, async () => {
        try {
            const cacheBuster = forceRefresh ? `?_=${Date.now()}` : '';
            const response = await fetch(`${API_BASE}/api/dev/mailboxes${cacheBuster}`, {
                method: 'GET',
                credentials: 'include',
                cache: 'no-store',
            });

            if (!response.ok) {
                throw await getApiError(response, 'Unable to list mailboxes');
            }

            const result = await response.json();
            setCachedData(cacheKey, result);
            return result;
        } catch (error) {
            console.error('Error listing developer mailboxes:', error);
            throw error;
        }
    });
}

export async function deleteDeveloperMailbox(identifier: string): Promise<MailboxDeleteResponse> {
    const response = await fetch(`${API_BASE}/api/dev/mailboxes/${encodeURIComponent(identifier)}`, {
        method: 'DELETE',
        credentials: 'include',
        cache: 'no-store',
    });

    if (!response.ok) {
        throw await getApiError(response, 'Unable to delete mailbox');
    }

    clearCacheForAddress(identifier);
    requestCache.delete('dev-mailboxes');
    pendingRequests.delete('dev-mailboxes');

    return response.json();
}