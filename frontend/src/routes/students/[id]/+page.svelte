<script lang="ts">
	import { enhance } from '$app/forms';
	import EnrollForm from '#lib/features/enrollments/components/EnrollForm.svelte';
	import EnrollmentTable from '#lib/features/enrollments/components/EnrollmentTable.svelte';
	import StudentForm from '#lib/features/students/components/StudentForm.svelte';
	import { fullName } from '#lib/features/students/types.ts';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">{fullName(data.student)}</h1>
	<form method="POST" action="?/delete" use:enhance>
		<button class="btn-danger" onclick={(e) => confirm('Delete this student?') || e.preventDefault()}>
			Delete student
		</button>
	</form>
</div>

<FormError message={form?.error} />
{#if form?.saved}<p class="mb-4 text-sm text-green-700">✅ Saved.</p>{/if}

<StudentForm action="?/update" values={form?.values ?? data.student} submitLabel="Save changes" />

<h2 class="mt-8 mb-3 text-lg font-semibold">Courses</h2>
<EnrollForm
	fixed={{ name: 'studentId', value: data.student.id }}
	pick={{ name: 'courseId', label: 'Choose a course' }}
	options={data.availableCourses.map((c) => ({ value: c.id, label: `${c.code} — ${c.name}` }))}
/>
<EnrollmentTable enrollments={data.enrollments} show="course" />
