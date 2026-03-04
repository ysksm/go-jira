<script lang="ts">
	import { onMount } from 'svelte';
	import { settings, loadSettings } from '$lib/stores/settings';
	import { configUpdate, configInitialize } from '$lib/api/client';

	let message = $state('');
	let error = $state('');

	// Init form
	let initEndpoint = $state('');
	let initUsername = $state('');
	let initApiKey = $state('');

	onMount(() => loadSettings());

	async function handleInit() {
		error = '';
		message = '';
		try {
			await configInitialize({
				endpoint: initEndpoint,
				username: initUsername,
				apiKey: initApiKey
			});
			message = 'Configuration initialized successfully';
			await loadSettings();
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}

	async function toggleSync(key: string, value: boolean) {
		try {
			await configUpdate({
				sync: {
					incrementalSyncEnabled: value,
					incrementalSyncMarginMinutes:
						$settings?.sync?.incrementalSyncMarginMinutes ?? 5
				}
			});
			await loadSettings();
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}
</script>

<svelte:head>
	<title>Settings - go-jira</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-6">Settings</h1>

{#if message}
	<div class="mb-4 p-3 bg-green-50 text-green-700 rounded">{message}</div>
{/if}
{#if error}
	<div class="mb-4 p-3 bg-red-50 text-red-700 rounded">{error}</div>
{/if}

{#if $settings}
	<!-- Endpoints -->
	<section class="bg-white rounded-lg shadow p-4 mb-6">
		<h2 class="font-bold mb-3">JIRA Endpoints</h2>
		{#if $settings.jiraEndpoints?.length}
			<table class="w-full text-sm mb-4">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="text-left px-3 py-2">Name</th>
						<th class="text-left px-3 py-2">Endpoint</th>
						<th class="text-left px-3 py-2">Username</th>
						<th class="text-left px-3 py-2">Active</th>
					</tr>
				</thead>
				<tbody>
					{#each $settings.jiraEndpoints as ep}
						<tr class="border-b">
							<td class="px-3 py-2">{ep.name}</td>
							<td class="px-3 py-2 font-mono text-xs">{ep.endpoint}</td>
							<td class="px-3 py-2">{ep.username}</td>
							<td class="px-3 py-2">
								{#if ep.name === $settings.activeEndpoint}
									<span class="text-green-600 text-xs font-medium">Active</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{:else if $settings.jira}
			<p class="text-sm text-gray-500 mb-2">
				Legacy config: {$settings.jira.endpoint} ({$settings.jira.username})
			</p>
		{:else}
			<p class="text-sm text-gray-500 mb-4">No endpoints configured.</p>
			<div class="grid gap-3 max-w-md">
				<input class="border rounded px-3 py-2 text-sm" placeholder="JIRA Endpoint URL" bind:value={initEndpoint} />
				<input class="border rounded px-3 py-2 text-sm" placeholder="Username" bind:value={initUsername} />
				<input class="border rounded px-3 py-2 text-sm" type="password" placeholder="API Key" bind:value={initApiKey} />
				<button
					class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm"
					onclick={handleInit}
				>
					Initialize
				</button>
			</div>
		{/if}
	</section>

	<!-- Sync Settings -->
	<section class="bg-white rounded-lg shadow p-4 mb-6">
		<h2 class="font-bold mb-3">Sync Settings</h2>
		<div class="flex items-center gap-3">
			<label class="text-sm">
				<input
					type="checkbox"
					checked={$settings.sync?.incrementalSyncEnabled ?? true}
					onchange={(e) => toggleSync('incrementalSync', (e.target as HTMLInputElement).checked)}
				/>
				Incremental Sync
			</label>
		</div>
	</section>

	<!-- Database -->
	<section class="bg-white rounded-lg shadow p-4">
		<h2 class="font-bold mb-3">Database</h2>
		<p class="text-sm text-gray-600">
			Path: <code class="bg-gray-100 px-1 rounded">{$settings.database.databaseDir || $settings.database.path || 'default'}</code>
		</p>
	</section>
{:else}
	<p class="text-gray-500">Loading settings...</p>
{/if}
