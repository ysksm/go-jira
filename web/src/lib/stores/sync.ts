import { writable } from 'svelte/store';
import type { SyncProgress, SyncResult } from '$lib/types';

export const syncInProgress = writable(false);
export const syncProgress = writable<SyncProgress | null>(null);
export const syncResults = writable<SyncResult[]>([]);
export const syncError = writable<string | null>(null);
