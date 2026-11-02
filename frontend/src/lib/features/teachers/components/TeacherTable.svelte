<script lang="ts">
	import { enhance } from '$app/forms';
	import { fullName, type Teacher } from '../types.ts';

	let { teachers }: { teachers: Teacher[] } = $props();
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
							<form method="POST" action="?/delete" use:enhance>
								<input type="hidden" name="id" value={teacher.id} />
								<button
									class="btn-danger"
									onclick={(e) => confirm(`Delete ${fullName(teacher)}?`) || e.preventDefault()}
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
