<script lang="ts">
	import CourseForm from '#lib/features/courses/components/CourseForm.svelte';
	import FormError from '#lib/ui/FormError.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<h1 class="mb-6 text-2xl font-bold">New course</h1>
<FormError message={form?.error} />
<CourseForm values={form?.values} teachers={data.teachers} submitLabel="Create course" />
