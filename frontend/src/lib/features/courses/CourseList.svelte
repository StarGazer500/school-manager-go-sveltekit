<script lang="ts">
	import { page } from '$app/state';
	import FormError from '#lib/ui/FormError.svelte';
	import { showMessage } from '#lib/ui/flash.svelte.ts';
	import { deleteCourse, readCourses } from '#lib/services/courses/index.ts';
	import type { Course } from '#lib/services/courses/types.ts';
	import { readTeachers } from '#lib/services/teachers/index.ts';
	import type { Teacher } from '#lib/services/teachers/types.ts';
	import CourseTable from './CourseTable.svelte';

	let courses = $state<Course[]>([]);
	let teachers = $state<Teacher[]>([]);
	let busy = $state(true);
	let error = $state('');

	const search = $derived(page.url.searchParams.get('q') ?? '');

	async function load(q: string) {
		// Teachers are loaded too so the table can show names, not just IDs.
		const [c, t] = await Promise.all([readCourses({ search: q }), readTeachers()]);
		busy = false;
		if (!c.success || !t.success) {
			error = (!c.success ? c : t).message;
			return;
		}
		error = '';
		courses = c.data;
		teachers = t.data;
	}

	// Runs on mount and again whenever ?q= changes.
	$effect(() => {
		load(search);
	});

	async function remove(course: Course) {
		const ret = await deleteCourse(course.id);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		showMessage(ret.message);
		await load(search);
	}
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">Courses</h1>
	<a class="btn" href="/courses/new">+ New course</a>
</div>

<form class="mb-4 flex max-w-md gap-2" data-sveltekit-keepfocus>
	<input class="input" name="q" type="search" placeholder="Search code or name" value={search} />
	<button class="btn-secondary">Search</button>
</form>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else}
	<CourseTable {courses} {teachers} ondelete={remove} />
{/if}
