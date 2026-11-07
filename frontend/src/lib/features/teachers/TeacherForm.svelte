<script lang="ts">
	import { parseTeacherForm } from './form.ts';
	import type { TeacherInput } from '#lib/services/teachers/types.ts';

	let {
		values,
		onsave,
		submitLabel = 'Save'
	}: { values?: Partial<TeacherInput>; onsave: (input: TeacherInput) => unknown; submitLabel?: string } = $props();

	// The page decides what saving means (create or update); this component only
	// stops the browser's own submit and hands over the parsed values.
	function submit(e: SubmitEvent & { currentTarget: HTMLFormElement }) {
		e.preventDefault();
		onsave(parseTeacherForm(new FormData(e.currentTarget)));
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
		Subject
		<input class="input" name="subject" required maxlength="100" value={values?.subject ?? ''} />
	</label>
	<div>
		<button class="btn" type="submit">{submitLabel}</button>
	</div>
</form>
