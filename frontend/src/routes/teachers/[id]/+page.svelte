<script lang="ts">
	import { enhance } from '$app/forms';
	import TeacherForm from '#lib/features/teachers/components/TeacherForm.svelte';
	import { fullName } from '#lib/features/teachers/types.ts';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">{fullName(data.teacher)}</h1>
	<form method="POST" action="?/delete" use:enhance>
		<button
			class="btn-danger"
			onclick={(e) => confirm('Delete this teacher? Their courses will become unassigned.') || e.preventDefault()}
		>
			Delete teacher
		</button>
	</form>
</div>

<FormError message={form?.error} />
{#if form?.saved}<p class="mb-4 text-sm text-green-700">✅ Saved.</p>{/if}

<TeacherForm action="?/update" values={form?.values ?? data.teacher} submitLabel="Save changes" />

<h2 class="mt-8 mb-3 text-lg font-semibold">Courses taught</h2>
{#if data.courses.length === 0}
	<p class="text-sm text-gray-500 italic">Not teaching any courses. Assign them from a course's page.</p>
{:else}
	<ul class="list-disc space-y-1 pl-5 text-sm">
		{#each data.courses as course (course.id)}
			<li><a class="link font-mono" href="/courses/{course.id}">{course.code}</a> {course.name}</li>
		{/each}
	</ul>
{/if}
