import { writable } from 'svelte/store';
import type { Settings } from '$lib/types';
import { configGet } from '$lib/api/client';

export const settings = writable<Settings | null>(null);
export const settingsLoading = writable(false);
export const settingsError = writable<string | null>(null);

export async function loadSettings() {
	settingsLoading.set(true);
	settingsError.set(null);
	try {
		const s = await configGet();
		settings.set(s);
	} catch (e) {
		settingsError.set(e instanceof Error ? e.message : String(e));
	} finally {
		settingsLoading.set(false);
	}
}
