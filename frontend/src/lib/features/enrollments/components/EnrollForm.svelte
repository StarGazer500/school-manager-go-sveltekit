<script lang="ts">
	import { enhance } from '$app/forms';

	// One side of the enrollment is fixed (the page we're on), the other is
	// picked from `options`. E.g. on a student page: fixed studentId, pick a course.
	let {
		fixed,
		pick,
		options
	}: {
		fixed: { name: 'studentId' | 'courseId'; value: number };
		pick: { name: 'studentId' | 'courseId'; label: string };
		options: { value: number; label: string }[];
	} = $props();
</script>

{#if options.length > 0}
	<form method="POST" action="?/enroll" class="mb-4 flex max-w-lg gap-2" use:enhance>
		<input type="hidden" name={fixed.name} value={fixed.value} />
		<select class="input" name={pick.name} required aria-label={pick.label}>
			<option value="">{pick.label}…</option>
			{#each options as option (option.value)}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
		<button class="btn shrink-0">Enroll</button>
	</form>
{/if}
