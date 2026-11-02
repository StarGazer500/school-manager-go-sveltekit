<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const stats = $derived([
		{ label: 'Students', value: data.counts.students, href: '/students' },
		{ label: 'Teachers', value: data.counts.teachers, href: '/teachers' },
		{ label: 'Courses', value: data.counts.courses, href: '/courses' },
		{ label: 'Enrollments', value: data.counts.enrollments, href: '/courses' }
	]);
</script>

<h1 class="mb-6 text-2xl font-bold">Dashboard</h1>

<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
	{#each stats as stat (stat.label)}
		<a class="card grid gap-1 transition hover:border-indigo-300" href={stat.href}>
			<span class="text-3xl font-bold">{stat.value}</span>
			<span class="text-sm text-gray-500">{stat.label}</span>
		</a>
	{/each}
</div>

<h2 class="mt-8 mb-3 text-lg font-semibold">Recent enrollments</h2>
{#if data.recentEnrollments.length === 0}
	<p class="text-sm text-gray-500 italic">Nothing yet. Add a teacher, a course and a student, then enroll them.</p>
{:else}
	<ul class="list-disc space-y-1 pl-5 text-sm">
		{#each data.recentEnrollments as e (e.id)}
			<li>
				<a class="link" href="/students/{e.studentId}">{e.studentName}</a> joined
				<a class="link" href="/courses/{e.courseId}">{e.courseCode}</a> on {new Date(e.enrolledAt).toLocaleDateString()}
			</li>
		{/each}
	</ul>
{/if}
