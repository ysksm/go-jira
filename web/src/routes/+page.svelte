<script lang="ts">
	import { onMount } from 'svelte';
	import { settings, loadSettings } from '$lib/stores/settings';
	import { syncStatus } from '$lib/api/client';

	let syncState = $state<{ inProgress: boolean } | null>(null);

	onMount(async () => {
		await loadSettings();
		try {
			syncState = await syncStatus();
		} catch {
			// ignore
		}
	});
</script>

<svelte:head>
	<title>Dashboard - go-jira</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-6">Dashboard</h1>

{#if $settings}
	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<!-- Endpoint -->
		<div class="bg-white rounded-lg shadow p-4">
			<h2 class="text-sm font-medium text-gray-500 mb-1">Active Endpoint</h2>
			<p class="text-lg font-semibold">{$settings.activeEndpoint || 'Not configured'}</p>
		</div>

		<!-- Projects -->
		<div class="bg-white rounded-lg shadow p-4">
			<h2 class="text-sm font-medium text-gray-500 mb-1">Projects</h2>
			<p class="text-lg font-semibold">
				{$settings.projects?.filter((p) => p.syncEnabled).length || 0} enabled
				/ {$settings.projects?.length || 0} total
			</p>
		</div>

		<!-- Sync Status -->
		<div class="bg-white rounded-lg shadow p-4">
			<h2 class="text-sm font-medium text-gray-500 mb-1">Sync Status</h2>
			<p class="text-lg font-semibold">
				{#if syncState?.inProgress}
					<span class="text-blue-600">In Progress</span>
				{:else}
					<span class="text-green-600">Idle</span>
				{/if}
			</p>
		</div>
	</div>

	<!-- Projects List -->
	{#if $settings.projects?.length}
		<h2 class="text-xl font-bold mt-8 mb-4">Projects</h2>
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="text-left px-4 py-2">Key</th>
						<th class="text-left px-4 py-2">Name</th>
						<th class="text-left px-4 py-2">Status</th>
						<th class="text-left px-4 py-2">Last Synced</th>
					</tr>
				</thead>
				<tbody>
					{#each $settings.projects as project}
						<tr class="border-b hover:bg-gray-50">
							<td class="px-4 py-2 font-mono">{project.key}</td>
							<td class="px-4 py-2">{project.name}</td>
							<td class="px-4 py-2">
								{#if project.syncEnabled}
									<span class="text-green-600 text-xs font-medium px-2 py-0.5 bg-green-100 rounded">Enabled</span>
								{:else}
									<span class="text-gray-500 text-xs font-medium px-2 py-0.5 bg-gray-100 rounded">Disabled</span>
								{/if}
							</td>
							<td class="px-4 py-2 text-gray-500">
								{project.lastSynced ? new Date(project.lastSynced).toLocaleString() : 'Never'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
{:else}
	<p class="text-gray-500">Loading...</p>
{/if}
