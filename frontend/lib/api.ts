const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:3001';

import type { CreateDeveloperRequest, DeveloperSignInRequest, DeveloperSignInResponse, Message, MessageDetail } from './types';
import { toast } from 'sonner';

const requestCache = new Map<string, { data: any, timestamps: number }>();
const pendingRequests = new Map<string, Promise<any>>()
const CACHE_DURATION = 6000

const getCachedData = (key: string) => {
    const cached = requestCache.get(key)
    if (cached && Date.now() - cached.timestamps < CACHE_DURATION) {
        console.log("cached hit for", key)
        return cached.data
    }
    return null
}

const setCachedData = (key: string, data: any) => {
    requestCache.set(key, { data: data, timestamps: Date.now() })
}

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
    if (pendingRequests.has(key)) {
        return pendingRequests.get(key)!
    }

    //finally return promise(api result) and does cleanup
    const promise = request().finally(() => {
        pendingRequests.delete(key)
    })

    pendingRequests.set(key, promise)
    return promise
}

export async function createCustomMailbox(username: string): Promise<{ address: string; createdAt: string; expiresAt: string | null }> {
    const cacheKey = `mailbox-${username}`;

    const cached = getCachedData(cacheKey);
    if (cached) {
        return cached;
    }

    return deduplicate(cacheKey, async () => {
        try {
            const cacheBuster = `?_=${Date.now()}`//changes url so browser won't cache

            const response = await fetch(`${API_BASE}/api/mailboxes/custom${cacheBuster}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username }),
                cache: 'no-store' // Ensure we don't use browser cache and make netwrok request
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));

                if (response.status === 429) {
                    console.log('Rate limit hit on mailbox creation');
                    toast.error('Rate limit exceeded', {
                        description: errorData.error || 'Too many mailboxes created. Please try again after 1 hour.',
                    });

                    throw new Error('Rate limit exceeded');
                }

                throw new Error(`Failed to create custom mailbox: ${response.status} ${errorData.error || response.statusText}`);
            }
            const data = await response.json();
            const result = {
                address: data.address,
                createdAt: data.createdAt,
                expiresAt: data.expiresAt
            };

            setCachedData(cacheKey, result);
            return result;
        } catch (error) {
            console.error('Error creating custom mailbox:', error);
            throw error;
        }
    });
}

export async function fetchMessages(address: string, forceRefresh = false): Promise<{ messages: Message[] }> {
    const cacheKey = `messages-${address}`;

    if (!forceRefresh) {
        const cached = getCachedData(cacheKey);
        if (cached) {
            return cached;
        }
    }

    return deduplicate(cacheKey, async () => {
        try {
            const cacheBuster = forceRefresh ? `?_=${Date.now()}` : '';
            const response = await fetch(`${API_BASE}/api/mailboxes/${encodeURIComponent(address)}/message${cacheBuster}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                cache: 'no-store'
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));

                if (response.status === 429) {
                    console.log('Rate limit hit on message fetch');
                    toast.error('Rate limit exceeded', {
                        description: errorData.error || 'Too many requests. Please slow down.',
                    });

                    throw new Error('Rate limit exceeded');
                }

                throw new Error(`Failed to fetch messages: ${response.status} ${errorData.error || response.statusText}`);
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

export async function fetchMessage(messageId: string): Promise<MessageDetail> {
    const cacheKey = `message-${messageId}`;

    const cached = getCachedData(cacheKey);
    if (cached) {
        return cached;
    }

    return deduplicate(cacheKey, async () => {
        try {
            const response = await fetch(`${API_BASE}/api/message/${messageId}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

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
                        }
                    });

                    if (retryResponse.ok) {
                        const result = await retryResponse.json();
                        setCachedData(cacheKey, result);
                        return result;
                    }
                }

                throw new Error(`Failed to fetch message: ${response.status} ${errorData.error || response.statusText}`);
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

async function getApiError(response: Response, fallback: string): Promise<Error> {
    const data = await response.json().catch(() => ({}));
    return new Error(data.error || `${fallback}: ${response.status} ${response.statusText}`);
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