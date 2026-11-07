<script lang="ts">
	// One side of the enrollment is fixed (the page we're on), the other is
	// picked from `options`. E.g. on a student page you pick a course, and the
	// page combines the picked id with the student's id.
	let {
		label,
		options,
		onenroll
	}: {
		label: string;
		options: { value: number; label: string }[];
		onenroll: (picked: number) => Promise<unknown>;
	} = $props();

	async function submit(e: SubmitEvent & { currentTarget: HTMLFormElement }) {
		e.preventDefault();
		const form = e.currentTarget;
		// Clear the picker only when it worked; `false` means the page showed an error.
		if ((await onenroll(Number(new FormData(form).get('picked')))) !== false) form.reset();
	}
</script>

{#if options.length > 0}
	<form class="mb-4 flex max-w-lg gap-2" onsubmit={submit}>
		<select class="input" name="picked" required aria-label={label}>
			<option value="">{label}…</option>
			{#each options as option (option.value)}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
		<button class="btn shrink-0">Enroll</button>
	</form>
{/if}
