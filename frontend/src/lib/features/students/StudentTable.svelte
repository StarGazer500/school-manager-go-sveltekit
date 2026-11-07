<script lang="ts">
	import { fullName, type Student } from '#lib/services/students/types.ts';

	let { students, ondelete }: { students: Student[]; ondelete: (student: Student) => unknown } = $props();
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
							<button
								type="button"
								class="btn-danger"
								onclick={() => confirm(`Delete ${fullName(student)}?`) && ondelete(student)}
							>
								Delete
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
