<script lang="ts">
	import { fullName, type Teacher } from '#lib/services/teachers/types.ts';
	import { parseCourseForm } from './form.ts';
	import type { CourseInput } from '#lib/services/courses/types.ts';

	let {
		values,
		teachers,
		onsave,
		submitLabel = 'Save'
	}: {
		values?: Partial<CourseInput>;
		teachers: Teacher[];
		onsave: (input: CourseInput) => unknown;
		submitLabel?: string;
	} = $props();

	// The page decides what saving means (create or update); this component only
	// stops the browser's own submit and hands over the parsed values.
	function submit(e: SubmitEvent & { currentTarget: HTMLFormElement }) {
		e.preventDefault();
		onsave(parseCourseForm(new FormData(e.currentTarget)));
	}
</script>

<form class="card grid max-w-lg gap-4" onsubmit={submit}>
	<div class="grid gap-4 sm:grid-cols-[10rem_1fr]">
		<label class="grid gap-1 text-sm font-medium text-gray-700">
			Code
			<input class="input" name="code" required maxlength="20" placeholder="MATH101" value={values?.code ?? ''} />
		</label>
		<label class="grid gap-1 text-sm font-medium text-gray-700">
			Name
			<input class="input" name="name" required maxlength="150" value={values?.name ?? ''} />
		</label>
	</div>
	<label class="grid gap-1 text-sm font-medium text-gray-700">
		Description
		<textarea class="input" name="description" rows="3" maxlength="1000">{values?.description ?? ''}</textarea>
	</label>
	<div class="grid gap-4 sm:grid-cols-[10rem_1fr]">
		<label class="grid gap-1 text-sm font-medium text-gray-700">
			Capacity
			<input
				class="input"
				name="capacity"
				type="number"
				min="1"
				max="500"
				required
				value={values?.capacity ?? 30}
			/>
		</label>
		<label class="grid gap-1 text-sm font-medium text-gray-700">
			Teacher
			<select class="input" name="teacherId" value={values?.teacherId ?? ''}>
				<option value="">Unassigned</option>
				{#each teachers as teacher (teacher.id)}
					<option value={teacher.id}>{fullName(teacher)} ({teacher.subject})</option>
				{/each}
			</select>
		</label>
	</div>
	<div>
		<button class="btn" type="submit">{submitLabel}</button>
	</div>
</form>
