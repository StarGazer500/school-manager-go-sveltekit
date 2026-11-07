<script lang="ts">
	import { goto } from '$app/navigation';
	import FormError from '#lib/ui/FormError.svelte';
	import { createCourse } from '#lib/services/courses/index.ts';
	import type { CourseInput } from '#lib/services/courses/types.ts';
	import { readTeachers } from '#lib/services/teachers/index.ts';
	import type { Teacher } from '#lib/services/teachers/types.ts';
	import { onMount } from 'svelte';
	import CourseForm from './CourseForm.svelte';

	let teachers = $state<Teacher[]>([]);
	let busy = $state(true);
	let error = $state('');

	// The teacher dropdown needs the list of teachers first.
	onMount(async () => {
		const ret = await readTeachers();
		busy = false;
		if (!ret.success) {
			error = ret.message;
			return;
		}
		teachers = ret.data;
	});

	// On error the form keeps what was typed, because nothing re-renders it.
	async function create(input: CourseInput) {
		const ret = await createCourse(input);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		await goto(`/courses/${ret.data.id}`);
	}
</script>

<h1 class="mb-6 text-2xl font-bold">New course</h1>
<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else}
	<CourseForm {teachers} onsave={create} submitLabel="Create course" />
{/if}
