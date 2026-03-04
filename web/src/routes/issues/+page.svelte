<script lang="ts">
	import { onMount } from 'svelte';
	import { issuesSearch, issuesGet, issuesHistory } from '$lib/api/client';
	import { settings, loadSettings } from '$lib/stores/settings';
	import type { Issue, ChangeHistoryItem } from '$lib/types';

	let project = $state('');
	let issues = $state<Issue[]>([]);
	let total = $state(0);
	let page = $state(0);
	let limit = 50;
	let loading = $state(false);
	let selectedIssue = $state<Issue | null>(null);
	let history = $state<ChangeHistoryItem[]>([]);

	onMount(() => loadSettings());

	async function search() {
		if (!project) return;
		loading = true;
		try {
			const res = await issuesSearch({ project, limit, offset: page * limit });
			issues = res.issues;
			total = res.total;
		} catch {
			issues = [];
		} finally {
			loading = false;
		}
	}

	async function viewIssue(key: string) {
		try {
			selectedIssue = await issuesGet(key);
			history = await issuesHistory(key);
		} catch {
			// ignore
		}
	}

	function closeDetail() {
		selectedIssue = null;
		history = [];
	}
</script>

<svelte:head>
	<title>Issues - go-jira</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-6">Issues</h1>

<!-- Search Form -->
<div class="flex gap-3 mb-6">
	<select class="border rounded px-3 py-2" bind:value={project} onchange={search}>
		<option value="">Select Project</option>
		{#if $settings?.projects}
			{#each $settings.projects.filter((p) => p.syncEnabled) as p}
				<option value={p.key}>{p.key} - {p.name}</option>
			{/each}
		{/if}
	</select>
	<button
		class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
		disabled={!project || loading}
		onclick={search}
	>
		Search
	</button>
	<span class="self-center text-sm text-gray-500">
		{#if total > 0}{total} issues{/if}
	</span>
</div>

<!-- Issue Detail Modal -->
{#if selectedIssue}
	<div class="fixed inset-0 bg-black/30 flex items-start justify-center pt-20 z-50" role="dialog">
		<div class="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[70vh] overflow-y-auto p-6">
			<div class="flex justify-between items-start mb-4">
				<div>
					<span class="text-sm text-gray-500 font-mono">{selectedIssue.key}</span>
					<h2 class="text-lg font-bold">{selectedIssue.summary}</h2>
				</div>
				<button class="text-gray-400 hover:text-gray-600 text-xl" onclick={closeDetail}>&times;</button>
			</div>
			<div class="grid grid-cols-2 gap-2 text-sm mb-4">
				<div><span class="text-gray-500">Status:</span> {selectedIssue.status}</div>
				<div><span class="text-gray-500">Priority:</span> {selectedIssue.priority}</div>
				<div><span class="text-gray-500">Type:</span> {selectedIssue.issueType}</div>
				<div><span class="text-gray-500">Assignee:</span> {selectedIssue.assignee || '-'}</div>
			</div>

			{#if history.length}
				<h3 class="font-bold text-sm mb-2">Change History</h3>
				<table class="w-full text-xs">
					<thead class="bg-gray-50">
						<tr>
							<th class="text-left px-2 py-1">Date</th>
							<th class="text-left px-2 py-1">Field</th>
							<th class="text-left px-2 py-1">From</th>
							<th class="text-left px-2 py-1">To</th>
						</tr>
					</thead>
					<tbody>
						{#each history as h}
							<tr class="border-b">
								<td class="px-2 py-1">{new Date(h.changedAt).toLocaleDateString()}</td>
								<td class="px-2 py-1">{h.field}</td>
								<td class="px-2 py-1 text-gray-500">{h.fromString || '-'}</td>
								<td class="px-2 py-1">{h.toString || '-'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
{/if}

<!-- Issues Table -->
{#if issues.length}
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<table class="w-full text-sm">
			<thead class="bg-gray-50 border-b">
				<tr>
					<th class="text-left px-4 py-2">Key</th>
					<th class="text-left px-4 py-2">Summary</th>
					<th class="text-left px-4 py-2">Status</th>
					<th class="text-left px-4 py-2">Priority</th>
					<th class="text-left px-4 py-2">Assignee</th>
				</tr>
			</thead>
			<tbody>
				{#each issues as issue}
					<tr class="border-b hover:bg-gray-50 cursor-pointer" onclick={() => viewIssue(issue.key)}>
						<td class="px-4 py-2 font-mono text-blue-600">{issue.key}</td>
						<td class="px-4 py-2">{issue.summary}</td>
						<td class="px-4 py-2">{issue.status || '-'}</td>
						<td class="px-4 py-2">{issue.priority || '-'}</td>
						<td class="px-4 py-2 text-gray-500">{issue.assignee || '-'}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<!-- Pagination -->
	<div class="flex justify-between items-center mt-4 text-sm">
		<button
			class="px-3 py-1 border rounded disabled:opacity-50"
			disabled={page === 0}
			onclick={() => { page--; search(); }}
		>
			Previous
		</button>
		<span class="text-gray-500">
			Page {page + 1} of {Math.ceil(total / limit)}
		</span>
		<button
			class="px-3 py-1 border rounded disabled:opacity-50"
			disabled={(page + 1) * limit >= total}
			onclick={() => { page++; search(); }}
		>
			Next
		</button>
	</div>
{:else if project && !loading}
	<p class="text-gray-500">No issues found.</p>
{/if}
