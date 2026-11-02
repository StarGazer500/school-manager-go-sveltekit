<script lang="ts">
	import { enhance } from '$app/forms';
	import type { TeacherInput } from '../types.ts';

	let {
		values,
		action = '',
		submitLabel = 'Save'
	}: { values?: Partial<TeacherInput>; action?: string; submitLabel?: string } = $props();
</script>

<form method="POST" {action} class="card grid max-w-lg gap-4" use:enhance>
	<div class="grid gap-4 sm:grid-cols-2">
		<label class="grid gap-1 text-sm font-medium text-gray-700">
			First name
			<input class="input" name="firstName" required maxlength="100" value={values?.firstName ?? ''} />
		</label>
		<label class="grid gap-1 text-sm font-medium text-gray-700">
			Last name
			<input class="input" name="lastName" required maxlength="100" value={values?.lastName ?? ''} />
		</label>
	</div>
	<label class="grid gap-1 text-sm font-medium text-gray-700">
		Email
		<input class="input" name="email" type="email" required value={values?.email ?? ''} />
	</label>
	<label class="grid gap-1 text-sm font-medium text-gray-700">
		Subject
		<input class="input" name="subject" required maxlength="100" value={values?.subject ?? ''} />
	</label>
	<div>
		<button class="btn" type="submit">{submitLabel}</button>
	</div>
</form>
