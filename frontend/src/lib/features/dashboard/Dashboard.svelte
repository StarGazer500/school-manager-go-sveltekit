<script lang="ts">
	import FormError from '#lib/ui/FormError.svelte';
	import { readCourses } from '#lib/services/courses/index.ts';
	import { readEnrollments } from '#lib/services/enrollments/index.ts';
	import type { Enrollment } from '#lib/services/enrollments/types.ts';
	import { readStudents } from '#lib/services/students/index.ts';
	import { readTeachers } from '#lib/services/teachers/index.ts';
	import { onMount } from 'svelte';

	let counts = $state({ students: 0, teachers: 0, courses: 0, enrollments: 0 });
	let recentEnrollments = $state<Enrollment[]>([]);
	let busy = $state(true);
	let error = $state('');

	onMount(async () => {
		// Fire all four requests at once rather than one after another.
		const [s, t, c, e] = await Promise.all([readStudents(), readTeachers(), readCourses(), readEnrollments()]);
		busy = false;
		if (!s.success || !t.success || !c.success || !e.success) {
			error = (!s.success ? s : !t.success ? t : !c.success ? c : e).message;
			return;
		}
		counts = { students: s.data.length, teachers: t.data.length, courses: c.data.length, enrollments: e.data.length };
		recentEnrollments = e.data.slice(0, 5);
	});

	const stats = $derived([
		{ label: 'Students', value: counts.students, href: '/students' },
		{ label: 'Teachers', value: counts.teachers, href: '/teachers' },
		{ label: 'Courses', value: counts.courses, href: '/courses' },
		{ label: 'Enrollments', value: counts.enrollments, href: '/courses' }
	]);
</script>

<h1 class="mb-6 text-2xl font-bold">Dashboard</h1>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else}
	<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
		{#each stats as stat (stat.label)}
			<a class="card grid gap-1 transition hover:border-indigo-300" href={stat.href}>
				<span class="text-3xl font-bold">{stat.value}</span>
				<span class="text-sm text-gray-500">{stat.label}</span>
			</a>
		{/each}
	</div>

	<h2 class="mt-8 mb-3 text-lg font-semibold">Recent enrollments</h2>
	{#if recentEnrollments.length === 0}
		<p class="text-sm text-gray-500 italic">Nothing yet. Add a teacher, a course and a student, then enroll them.</p>
	{:else}
		<ul class="list-disc space-y-1 pl-5 text-sm">
			{#each recentEnrollments as e (e.id)}
				<li>
					<a class="link" href="/students/{e.studentId}">{e.studentName}</a> joined
					<a class="link" href="/courses/{e.courseId}">{e.courseCode}</a> on {new Date(e.enrolledAt).toLocaleDateString()}
				</li>
			{/each}
		</ul>
	{/if}
{/if}
