<script lang="ts">
	import { enhance } from '$app/forms';
	import { fullName, type Teacher } from '#lib/features/teachers/types.ts';
	import type { Course } from '../types.ts';

	let { courses, teachers }: { courses: Course[]; teachers: Teacher[] } = $props();

	// $derived recomputes only when `teachers` changes.
	const teacherById = $derived(new Map(teachers.map((t) => [t.id, t])));
</script>

{#if courses.length === 0}
	<p class="text-sm text-gray-500 italic">No courses found.</p>
{:else}
	<div class="overflow-x-auto rounded-lg border border-gray-200 bg-white shadow-sm">
		<table class="data-table">
			<thead>
				<tr><th>Code</th><th>Name</th><th>Teacher</th><th>Capacity</th><th></th></tr>
			</thead>
			<tbody>
				{#each courses as course (course.id)}
					{@const teacher = course.teacherId ? teacherById.get(course.teacherId) : undefined}
					<tr class="hover:bg-gray-50">
						<td><a class="link font-mono" href="/courses/{course.id}">{course.code}</a></td>
						<td>{course.name}</td>
						<td>
							{#if teacher}
								<a class="link" href="/teachers/{teacher.id}">{fullName(teacher)}</a>
							{:else}
								<span class="text-gray-400 italic">Unassigned</span>
							{/if}
						</td>
						<td>{course.capacity}</td>
						<td class="text-right">
							<form method="POST" action="?/delete" use:enhance>
								<input type="hidden" name="id" value={course.id} />
								<button class="btn-danger" onclick={(e) => confirm(`Delete ${course.code}?`) || e.preventDefault()}>
									Delete
								</button>
							</form>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
