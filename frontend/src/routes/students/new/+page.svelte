<script lang="ts">
	import StudentForm from '#lib/features/students/components/StudentForm.svelte';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { form }: PageProps = $props();
</script>

<h1 class="mb-6 text-2xl font-bold">New student</h1>
<FormError message={form?.error} />
<StudentForm values={form?.values} submitLabel="Create student" />
