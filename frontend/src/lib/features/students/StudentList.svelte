<script lang="ts">
	import { page } from '$app/state';
	import FormError from '#lib/ui/FormError.svelte';
	import { showMessage } from '#lib/ui/flash.svelte.ts';
	import { deleteStudent, readStudents } from '#lib/services/students/index.ts';
	import type { Student } from '#lib/services/students/types.ts';
	import StudentTable from './StudentTable.svelte';

	let students = $state<Student[]>([]);
	let busy = $state(true);
	let error = $state('');

	// The search box is a plain GET form, so searching changes ?q= in the URL.
	const search = $derived(page.url.searchParams.get('q') ?? '');

	async function load(q: string) {
		const ret = await readStudents(q);
		busy = false;
		if (!ret.success) {
			error = ret.message;
			return;
		}
		error = '';
		students = ret.data;
	}

	// Like MyApp's onMount fetch, but it also runs again when ?q= changes,
	// because the router keeps this component mounted for /students?q=...
	$effect(() => {
		load(search);
	});

	async function remove(student: Student) {
		const ret = await deleteStudent(student.id);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		showMessage(ret.message);
		await load(search);
	}
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">Students</h1>
	<a class="btn" href="/students/new">+ New student</a>
</div>

<form class="mb-4 flex max-w-md gap-2" data-sveltekit-keepfocus>
	<input class="input" name="q" type="search" placeholder="Search name or email" value={search} />
	<button class="btn-secondary">Search</button>
</form>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else}
	<StudentTable {students} ondelete={remove} />
{/if}
