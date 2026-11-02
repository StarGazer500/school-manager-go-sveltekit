<script lang="ts">
	import { enhance } from '$app/forms';
	import CourseForm from '#lib/features/courses/components/CourseForm.svelte';
	import EnrollForm from '#lib/features/enrollments/components/EnrollForm.svelte';
	import EnrollmentTable from '#lib/features/enrollments/components/EnrollmentTable.svelte';
	import { fullName } from '#lib/features/students/types.ts';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const seatsLeft = $derived(data.course.capacity - data.enrollments.length);
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">
		<span class="font-mono">{data.course.code}</span>
		<span class="text-gray-500">·</span>
		{data.course.name}
	</h1>
	<form method="POST" action="?/delete" use:enhance>
		<button class="btn-danger" onclick={(e) => confirm('Delete this course and its enrollments?') || e.preventDefault()}>
			Delete course
		</button>
	</form>
</div>

<FormError message={form?.error} />
{#if form?.saved}<p class="mb-4 text-sm text-green-700">✅ Saved.</p>{/if}

<CourseForm action="?/update" values={form?.values ?? data.course} teachers={data.teachers} submitLabel="Save changes" />

<h2 class="mt-8 mb-3 flex items-center gap-3 text-lg font-semibold">
	Roster
	<span
		class={[
			'rounded-full px-2.5 py-0.5 text-xs font-medium',
			seatsLeft > 0 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
		]}
	>
		{data.enrollments.length}/{data.course.capacity}
	</span>
</h2>
{#if seatsLeft > 0}
	<EnrollForm
		fixed={{ name: 'courseId', value: data.course.id }}
		pick={{ name: 'studentId', label: 'Choose a student' }}
		options={data.availableStudents.map((s) => ({ value: s.id, label: `${fullName(s)} (grade ${s.gradeLevel})` }))}
	/>
{:else}
	<p class="mb-4 text-sm text-gray-500 italic">Course is full.</p>
{/if}
<EnrollmentTable enrollments={data.enrollments} show="student" />
