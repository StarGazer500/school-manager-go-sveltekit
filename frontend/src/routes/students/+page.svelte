<script lang="ts">
	import StudentTable from '#lib/features/students/components/StudentTable.svelte';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	// `data` comes from load(), `form` from the last action's return value.
	let { data, form }: PageProps = $props();
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">Students</h1>
	<a class="btn" href="/students/new">+ New student</a>
</div>

<!-- A plain GET form: submitting changes the URL to ?q=..., which re-runs load(). -->
<form class="mb-4 flex max-w-md gap-2" data-sveltekit-keepfocus>
	<input class="input" name="q" type="search" placeholder="Search name or email" value={data.q} />
	<button class="btn-secondary">Search</button>
</form>

<FormError message={form?.error} />
<StudentTable students={data.students} />
