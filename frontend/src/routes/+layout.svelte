<script lang="ts">
	import '../app.css';
	import favicon from '#lib/assets/favicon.svg';
	import Flash from '#lib/ui/Flash.svelte';
	import { page } from '$app/state';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const links = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/students', label: 'Students' },
		{ href: '/teachers', label: 'Teachers' },
		{ href: '/courses', label: 'Courses' }
	];

	const isActive = (href: string) =>
		href === '/' ? page.url.pathname === '/' : page.url.pathname.startsWith(href);
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>School Manager</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 text-gray-900">
	<header class="border-b border-gray-200 bg-white">
		<nav class="mx-auto flex max-w-5xl flex-wrap items-center gap-5 px-4 py-4">
			<strong class="mr-auto text-lg">🏫 School Manager</strong>
			{#each links as link (link.href)}
				<a
					href={link.href}
					aria-current={isActive(link.href) ? 'page' : undefined}
					class={isActive(link.href) ? 'font-semibold text-indigo-600' : 'text-gray-500 hover:text-gray-900'}
				>
					{link.label}
				</a>
			{/each}
		</nav>
	</header>

	<main class="mx-auto max-w-5xl px-4 pt-6 pb-12">
		<Flash />
		{@render children()}
	</main>
</div>
