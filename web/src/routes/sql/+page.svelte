<script lang="ts">
	import { onMount } from 'svelte';
	import { sqlExecute, sqlGetSchema } from '$lib/api/client';
	import { settings, loadSettings } from '$lib/stores/settings';
	import type { SqlTable } from '$lib/types';

	let project = $state('');
	let query = $state('SELECT * FROM issues LIMIT 10');
	let columns = $state<string[]>([]);
	let rows = $state<unknown[][]>([]);
	let rowCount = $state(0);
	let executionTime = $state(0);
	let error = $state('');
	let loading = $state(false);
	let schema = $state<SqlTable[]>([]);
	let showSchema = $state(false);

	onMount(() => loadSettings());

	async function execute() {
		if (!project || !query.trim()) return;
		loading = true;
		error = '';
		try {
			const res = await sqlExecute({ projectKey: project, query, limit: 100 });
			columns = res.columns;
			rows = res.rows;
			rowCount = res.rowCount;
			executionTime = res.executionTimeMs;
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
			columns = [];
			rows = [];
		} finally {
			loading = false;
		}
	}

	async function loadSchema() {
		if (!project) return;
		try {
			schema = await sqlGetSchema(project);
			showSchema = true;
		} catch {
			schema = [];
		}
	}
</script>

<svelte:head>
	<title>SQL - go-jira</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-6">SQL Query</h1>

<div class="flex gap-3 mb-4">
	<select class="border rounded px-3 py-2" bind:value={project} onchange={loadSchema}>
		<option value="">Select Project</option>
		{#if $settings?.projects}
			{#each $settings.projects.filter((p) => p.syncEnabled) as p}
				<option value={p.key}>{p.key}</option>
			{/each}
		{/if}
	</select>
	<button
		class="px-3 py-2 text-sm border rounded hover:bg-gray-50"
		onclick={() => (showSchema = !showSchema)}
		disabled={!project}
	>
		{showSchema ? 'Hide' : 'Show'} Schema
	</button>
</div>

<!-- Schema Browser -->
{#if showSchema && schema.length}
	<div class="bg-gray-50 border rounded p-4 mb-4 text-xs max-h-48 overflow-y-auto">
		{#each schema as table}
			<div class="mb-2">
				<span class="font-bold">{table.name}</span>
				<span class="text-gray-400 ml-2">
					({table.columns.map((c) => `${c.name} ${c.type}`).join(', ')})
				</span>
			</div>
		{/each}
	</div>
{/if}

<!-- Query Editor -->
<div class="mb-4">
	<textarea
		class="w-full h-32 border rounded p-3 font-mono text-sm"
		bind:value={query}
		placeholder="Enter SQL query..."
	></textarea>
	<div class="flex justify-between mt-2">
		<button
			class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
			disabled={!project || !query.trim() || loading}
			onclick={execute}
		>
			{loading ? 'Running...' : 'Execute'}
		</button>
		{#if rowCount > 0}
			<span class="self-center text-sm text-gray-500">
				{rowCount} rows in {executionTime}ms
			</span>
		{/if}
	</div>
</div>

{#if error}
	<div class="p-3 bg-red-50 text-red-700 rounded mb-4 text-sm font-mono">{error}</div>
{/if}

<!-- Results -->
{#if columns.length}
	<div class="bg-white rounded-lg shadow overflow-x-auto">
		<table class="w-full text-xs">
			<thead class="bg-gray-50 border-b">
				<tr>
					{#each columns as col}
						<th class="text-left px-3 py-2 whitespace-nowrap">{col}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each rows as row}
					<tr class="border-b hover:bg-gray-50">
						{#each row as cell}
							<td class="px-3 py-1 whitespace-nowrap max-w-xs truncate">
								{cell ?? ''}
							</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
