<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { syncExecute, syncProgressSSE, syncStatus } from '$lib/api/client';
	import type { SyncProgress, SyncResult } from '$lib/types';

	let inProgress = $state(false);
	let progress = $state<SyncProgress | null>(null);
	let results = $state<SyncResult[]>([]);
	let error = $state('');
	let eventSource: EventSource | null = null;

	const phaseLabels: Record<string, string> = {
		fetch_issues: 'Fetching Issues',
		sync_metadata: 'Syncing Metadata',
		generate_snapshots: 'Generating Snapshots',
		verify_integrity: 'Verifying Integrity'
	};

	onMount(async () => {
		try {
			const status = await syncStatus();
			inProgress = status.inProgress;
			if (status.progress) progress = status.progress;
			if (inProgress) startSSE();
		} catch {
			// ignore
		}
	});

	onDestroy(() => {
		eventSource?.close();
	});

	function startSSE() {
		eventSource?.close();
		eventSource = syncProgressSSE(
			(data) => {
				inProgress = data.inProgress;
				progress = data.progress ?? null;
			},
			() => {
				inProgress = false;
			}
		);
	}

	async function handleSync(projectKey?: string) {
		error = '';
		inProgress = true;
		startSSE();
		try {
			results = await syncExecute({ projectKey });
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			inProgress = false;
		}
	}
</script>

<svelte:head>
	<title>Sync - go-jira</title>
</svelte:head>

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-bold">Sync</h1>
	<button
		class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
		disabled={inProgress}
		onclick={() => handleSync()}
	>
		{inProgress ? 'Syncing...' : 'Start Sync'}
	</button>
</div>

{#if error}
	<div class="mb-4 p-3 bg-red-50 text-red-700 rounded">{error}</div>
{/if}

<!-- Progress -->
{#if inProgress && progress}
	<div class="bg-white rounded-lg shadow p-4 mb-6">
		<div class="flex justify-between text-sm mb-2">
			<span class="font-medium">{progress.projectKey}</span>
			<span class="text-gray-500">{phaseLabels[progress.phase] || progress.phase}</span>
		</div>
		<div class="w-full bg-gray-200 rounded-full h-2.5">
			<div
				class="bg-blue-600 h-2.5 rounded-full transition-all"
				style="width: {progress.total > 0 ? (progress.current / progress.total) * 100 : 0}%"
			></div>
		</div>
		<p class="text-xs text-gray-500 mt-1">
			{progress.current} / {progress.total} — {progress.message}
		</p>
	</div>
{/if}

<!-- Results -->
{#if results.length}
	<h2 class="text-xl font-bold mb-4">Results</h2>
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<table class="w-full text-sm">
			<thead class="bg-gray-50 border-b">
				<tr>
					<th class="text-left px-4 py-2">Project</th>
					<th class="text-left px-4 py-2">Issues</th>
					<th class="text-left px-4 py-2">Duration</th>
					<th class="text-left px-4 py-2">Status</th>
				</tr>
			</thead>
			<tbody>
				{#each results as r}
					<tr class="border-b">
						<td class="px-4 py-2 font-mono">{r.projectKey}</td>
						<td class="px-4 py-2">{r.issueCount}</td>
						<td class="px-4 py-2">{r.duration.toFixed(1)}s</td>
						<td class="px-4 py-2">
							{#if r.success}
								<span class="text-green-600">Success</span>
							{:else}
								<span class="text-red-600">{r.error || 'Failed'}</span>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
