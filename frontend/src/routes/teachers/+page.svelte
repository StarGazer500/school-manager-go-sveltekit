<script lang="ts">
	import TeacherTable from '#lib/features/teachers/components/TeacherTable.svelte';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
	<h1 class="text-2xl font-bold">Teachers</h1>
	<a class="btn" href="/teachers/new">+ New teacher</a>
</div>

<form class="mb-4 flex max-w-md gap-2" data-sveltekit-keepfocus>
	<input class="input" name="q" type="search" placeholder="Search name or subject" value={data.q} />
	<button class="btn-secondary">Search</button>
</form>

<FormError message={form?.error} />
<TeacherTable teachers={data.teachers} />
