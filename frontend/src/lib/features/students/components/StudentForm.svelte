<script lang="ts">
	import { enhance } from '$app/forms';
	import type { StudentInput } from '../types.ts';

	let {
		values,
		action = '',
		submitLabel = 'Save'
	}: { values?: Partial<StudentInput>; action?: string; submitLabel?: string } = $props();
</script>

<!-- use:enhance submits with fetch instead of a full page reload; without JS it still works. -->
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
		Grade level
		<input
			class="input w-32"
			name="gradeLevel"
			type="number"
			min="1"
			max="12"
			required
			value={values?.gradeLevel ?? ''}
		/>
	</label>
	<div>
		<button class="btn" type="submit">{submitLabel}</button>
	</div>
</form>
