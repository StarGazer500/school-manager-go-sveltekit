<script lang="ts">
	import TeacherForm from '#lib/features/teachers/components/TeacherForm.svelte';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { form }: PageProps = $props();
</script>

<h1 class="mb-6 text-2xl font-bold">New teacher</h1>
<FormError message={form?.error} />
<TeacherForm values={form?.values} submitLabel="Create teacher" />
