<script lang="ts">
	import { goto } from '$app/navigation';
	import FormError from '#lib/ui/FormError.svelte';
	import { createStudent } from '#lib/services/students/index.ts';
	import type { StudentInput } from '#lib/services/students/types.ts';
	import StudentForm from './StudentForm.svelte';

	let error = $state('');

	// On error the form keeps what was typed, because nothing re-renders it.
	async function create(input: StudentInput) {
		const ret = await createStudent(input);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		await goto(`/students/${ret.data.id}`);
	}
</script>

<h1 class="mb-6 text-2xl font-bold">New student</h1>
<FormError message={error} />
<StudentForm onsave={create} submitLabel="Create student" />
