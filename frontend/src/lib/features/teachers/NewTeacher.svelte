<script lang="ts">
	import { goto } from '$app/navigation';
	import FormError from '#lib/ui/FormError.svelte';
	import { createTeacher } from '#lib/services/teachers/index.ts';
	import type { TeacherInput } from '#lib/services/teachers/types.ts';
	import TeacherForm from './TeacherForm.svelte';

	let error = $state('');

	// On error the form keeps what was typed, because nothing re-renders it.
	async function create(input: TeacherInput) {
		const ret = await createTeacher(input);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		await goto(`/teachers/${ret.data.id}`);
	}
</script>

<h1 class="mb-6 text-2xl font-bold">New teacher</h1>
<FormError message={error} />
<TeacherForm onsave={create} submitLabel="Create teacher" />
