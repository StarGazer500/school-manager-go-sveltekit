<script lang="ts">
	import { fullName, type Teacher } from '#lib/services/teachers/types.ts';

	let { teachers, ondelete }: { teachers: Teacher[]; ondelete: (teacher: Teacher) => unknown } = $props();
</script>

{#if teachers.length === 0}
	<p class="text-sm text-gray-500 italic">No teachers found.</p>
{:else}
	<div class="overflow-x-auto rounded-lg border border-gray-200 bg-white shadow-sm">
		<table class="data-table">
			<thead>
				<tr><th>Name</th><th>Email</th><th>Subject</th><th></th></tr>
			</thead>
			<tbody>
				{#each teachers as teacher (teacher.id)}
					<tr class="hover:bg-gray-50">
						<td><a class="link" href="/teachers/{teacher.id}">{fullName(teacher)}</a></td>
						<td class="text-gray-600">{teacher.email}</td>
						<td>{teacher.subject}</td>
						<td class="text-right">
							<button
								type="button"
								class="btn-danger"
								onclick={() => confirm(`Delete ${fullName(teacher)}?`) && ondelete(teacher)}
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
