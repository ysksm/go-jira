import { writable } from 'svelte/store';
import type { Project } from '$lib/types';
import { projectsList } from '$lib/api/client';

export const projects = writable<Project[]>([]);
export const projectsLoading = writable(false);

export async function loadProjects() {
	projectsLoading.set(true);
	try {
		const p = await projectsList();
		projects.set(p);
	} catch {
		// ignore
	} finally {
		projectsLoading.set(false);
	}
}
