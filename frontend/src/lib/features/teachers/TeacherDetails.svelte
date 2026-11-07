<script lang="ts">
	import { goto } from '$app/navigation';
	import FormError from '#lib/ui/FormError.svelte';
	import { showMessage } from '#lib/ui/flash.svelte.ts';
	import { readCourses } from '#lib/services/courses/index.ts';
	import type { Course } from '#lib/services/courses/types.ts';
	import { deleteTeacher, readTeacher, updateTeacher } from '#lib/services/teachers/index.ts';
	import { fullName, type Teacher, type TeacherInput } from '#lib/services/teachers/types.ts';
	import TeacherForm from './TeacherForm.svelte';

	let { id }: { id: number } = $props();

	let teacher = $state<Teacher | null>(null);
	let courses = $state<Course[]>([]);
	let busy = $state(true);
	let error = $state('');
	let saved = $state(false);

	async function load(teacherId: number) {
		const [t, c] = await Promise.all([readTeacher(teacherId), readCourses({ teacherId })]);
		busy = false;
		if (!t.success || !c.success) {
			error = (!t.success ? t : c).message;
			return;
		}
		teacher = t.data;
		courses = c.data;
	}

	// Runs on mount and again if the route moves to another teacher's id.
	$effect(() => {
		load(id);
	});

	async function save(input: TeacherInput) {
		saved = false;
		const ret = await updateTeacher(id, input);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		error = '';
		saved = true;
		await load(id);
	}

	async function remove() {
		if (!confirm('Delete this teacher? Their courses will become unassigned.')) return;
		const ret = await deleteTeacher(id);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		await goto('/teachers');
		showMessage(ret.message); // after goto, so the list page shows it
	}
</script>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if teacher}
	<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
		<h1 class="text-2xl font-bold">{fullName(teacher)}</h1>
		<button type="button" class="btn-danger" onclick={remove}>Delete teacher</button>
	</div>

	{#if saved}<p class="mb-4 text-sm text-green-700">✅ Saved.</p>{/if}

	<TeacherForm values={teacher} onsave={save} submitLabel="Save changes" />

	<h2 class="mt-8 mb-3 text-lg font-semibold">Courses taught</h2>
	{#if courses.length === 0}
		<p class="text-sm text-gray-500 italic">Not teaching any courses. Assign them from a course's page.</p>
	{:else}
		<ul class="list-disc space-y-1 pl-5 text-sm">
			{#each courses as course (course.id)}
				<li><a class="link font-mono" href="/courses/{course.id}">{course.code}</a> {course.name}</li>
			{/each}
		</ul>
	{/if}
{:else}
	<a class="link" href="/teachers">Back to teachers</a>
{/if}
