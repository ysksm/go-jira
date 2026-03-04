<script lang="ts">
	import { onMount } from 'svelte';
	import { settings, loadSettings } from '$lib/stores/settings';
	import { projectsInitialize, projectsEnable, projectsDisable } from '$lib/api/client';

	let fetching = $state(false);
	let message = $state('');

	onMount(() => loadSettings());

	async function handleFetch() {
		fetching = true;
		message = '';
		try {
			const result = await projectsInitialize();
			message = `Fetched ${result.newCount} new projects (${result.projects.length} total)`;
			await loadSettings();
		} catch (e) {
			message = `Error: ${e instanceof Error ? e.message : e}`;
		} finally {
			fetching = false;
		}
	}

	async function toggleProject(key: string, enabled: boolean) {
		try {
			if (enabled) {
				await projectsDisable(key);
			} else {
				await projectsEnable(key);
			}
			await loadSettings();
		} catch (e) {
			message = `Error: ${e instanceof Error ? e.message : e}`;
		}
	}
</script>

<svelte:head>
	<title>Projects - go-jira</title>
</svelte:head>

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-bold">Projects</h1>
	<button
		class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
		disabled={fetching}
		onclick={handleFetch}
	>
		{fetching ? 'Fetching...' : 'Fetch from JIRA'}
	</button>
</div>

{#if message}
	<div class="mb-4 p-3 bg-blue-50 text-blue-700 rounded">{message}</div>
{/if}

{#if $settings?.projects?.length}
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<table class="w-full text-sm">
			<thead class="bg-gray-50 border-b">
				<tr>
					<th class="text-left px-4 py-2">Key</th>
					<th class="text-left px-4 py-2">Name</th>
					<th class="text-left px-4 py-2">Endpoint</th>
					<th class="text-left px-4 py-2">Sync</th>
					<th class="text-left px-4 py-2">Action</th>
				</tr>
			</thead>
			<tbody>
				{#each $settings.projects as project}
					<tr class="border-b hover:bg-gray-50">
						<td class="px-4 py-2 font-mono">{project.key}</td>
						<td class="px-4 py-2">{project.name}</td>
						<td class="px-4 py-2 text-gray-500">{project.endpoint || '-'}</td>
						<td class="px-4 py-2">
							{#if project.syncEnabled}
								<span class="text-green-600">Enabled</span>
							{:else}
								<span class="text-gray-400">Disabled</span>
							{/if}
						</td>
						<td class="px-4 py-2">
							<button
								class="text-sm px-3 py-1 rounded border hover:bg-gray-100"
								onclick={() => toggleProject(project.key, project.syncEnabled)}
							>
								{project.syncEnabled ? 'Disable' : 'Enable'}
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<p class="text-gray-500">No projects. Click "Fetch from JIRA" to load projects.</p>
{/if}
