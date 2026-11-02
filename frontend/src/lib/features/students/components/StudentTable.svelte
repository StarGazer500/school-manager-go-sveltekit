<script lang="ts">
	import { enhance } from '$app/forms';
	import { fullName, type Student } from '../types.ts';

	let { students }: { students: Student[] } = $props();
</script>

{#if students.length === 0}
	<p class="text-sm text-gray-500 italic">No students found.</p>
{:else}
	<div class="overflow-x-auto rounded-lg border border-gray-200 bg-white shadow-sm">
		<table class="data-table">
			<thead>
				<tr><th>Name</th><th>Email</th><th>Grade</th><th></th></tr>
			</thead>
			<tbody>
				{#each students as student (student.id)}
					<tr class="hover:bg-gray-50">
						<td><a class="link" href="/students/{student.id}">{fullName(student)}</a></td>
						<td class="text-gray-600">{student.email}</td>
						<td>{student.gradeLevel}</td>
						<td class="text-right">
							<form method="POST" action="?/delete" use:enhance>
								<input type="hidden" name="id" value={student.id} />
								<button
									class="btn-danger"
									onclick={(e) => confirm(`Delete ${fullName(student)}?`) || e.preventDefault()}
								>
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
