<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	const navItems = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/projects', label: 'Projects' },
		{ href: '/sync', label: 'Sync' },
		{ href: '/issues', label: 'Issues' },
		{ href: '/sql', label: 'SQL' },
		{ href: '/settings', label: 'Settings' }
	];

	let { children } = $props();
</script>

<div class="min-h-screen flex">
	<!-- Sidebar -->
	<nav class="w-56 bg-gray-900 text-white flex flex-col">
		<div class="p-4 text-lg font-bold border-b border-gray-700">go-jira</div>
		<ul class="flex-1 py-2">
			{#each navItems as item}
				<li>
					<a
						href={item.href}
						class="block px-4 py-2 hover:bg-gray-800 transition-colors"
						class:bg-gray-800={$page.url.pathname === item.href}
					>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
	</nav>

	<!-- Main -->
	<main class="flex-1 p-6">
		{@render children()}
	</main>
</div>
