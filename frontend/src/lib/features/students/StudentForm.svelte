<script lang="ts">
	import { parseStudentForm } from './form.ts';
	import type { StudentInput } from '#lib/services/students/types.ts';

	let {
		values,
		onsave,
		submitLabel = 'Save'
	}: { values?: Partial<StudentInput>; onsave: (input: StudentInput) => unknown; submitLabel?: string } = $props();

	// The page decides what saving means (create or update); this component only
	// stops the browser's own submit and hands over the parsed values.
	function submit(e: SubmitEvent & { currentTarget: HTMLFormElement }) {
		e.preventDefault();
		onsave(parseStudentForm(new FormData(e.currentTarget)));
	}
</script>

<form class="card grid max-w-lg gap-4" onsubmit={submit}>
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
