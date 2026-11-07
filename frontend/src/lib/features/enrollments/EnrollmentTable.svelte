<script lang="ts">
	import type { Enrollment } from '#lib/services/enrollments/types.ts';

	// `show` picks which side of the relationship to list: on a student's page
	// we show their courses, on a course's page we show its students.
	let {
		enrollments,
		show,
		onwithdraw
	}: { enrollments: Enrollment[]; show: 'course' | 'student'; onwithdraw: (e: Enrollment) => unknown } = $props();
</script>

{#if enrollments.length === 0}
	<p class="text-sm text-gray-500 italic">No enrollments yet.</p>
{:else}
	<div class="overflow-x-auto rounded-lg border border-gray-200 bg-white shadow-sm">
		<table class="data-table">
			<thead>
				<tr><th>{show === 'course' ? 'Course' : 'Student'}</th><th>Enrolled</th><th></th></tr>
			</thead>
			<tbody>
				{#each enrollments as e (e.id)}
					<tr class="hover:bg-gray-50">
						<td>
							{#if show === 'course'}
								<a class="link font-mono" href="/courses/{e.courseId}">{e.courseCode}</a>
								<span class="text-gray-600">{e.courseName}</span>
							{:else}
								<a class="link" href="/students/{e.studentId}">{e.studentName}</a>
							{/if}
						</td>
						<td class="text-gray-600">{new Date(e.enrolledAt).toLocaleDateString()}</td>
						<td class="text-right">
							<button type="button" class="btn-secondary" onclick={() => onwithdraw(e)}>Withdraw</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
