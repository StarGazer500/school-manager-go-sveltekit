<script lang="ts">
	import { page } from '$app/state';
	import FormError from '#lib/ui/FormError.svelte';
	import { showMessage } from '#lib/ui/flash.svelte.ts';
	import { deleteTeacher, readTeachers } from '#lib/services/teachers/index.ts';
	import type { Teacher } from '#lib/services/teachers/types.ts';
	import TeacherTable from './TeacherTable.svelte';

	let teachers = $state<Teacher[]>([]);
	let busy = $state(true);
	let error = $state('');

	const search = $derived(page.url.searchParams.get('q') ?? '');

	async function load(q: string) {
		const ret = await readTeachers(q);
		busy = false;
		if (!ret.success) {
			error = ret.message;
			return;
		}
		error = '';
		teachers = ret.data;
	}

	// Runs on mount and again whenever ?q= changes.
	$effect(() => {
		load(search);
	});

	async function remove(teacher: Teacher) {
		const ret = await deleteTeacher(teacher.id);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		showMessage(ret.message);
		await load(search);
	}
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">Teachers</h1>
	<a class="btn" href="/teachers/new">+ New teacher</a>
</div>

<form class="mb-4 flex max-w-md gap-2" data-sveltekit-keepfocus>
	<input class="input" name="q" type="search" placeholder="Search name or subject" value={search} />
	<button class="btn-secondary">Search</button>
</form>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else}
	<TeacherTable {teachers} ondelete={remove} />
{/if}
