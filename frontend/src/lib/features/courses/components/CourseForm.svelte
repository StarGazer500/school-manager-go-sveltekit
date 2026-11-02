<script lang="ts">
	import { enhance } from '$app/forms';
	import { fullName, type Teacher } from '#lib/features/teachers/types.ts';
	import type { CourseInput } from '../types.ts';

	let {
		values,
		teachers,
		action = '',
		submitLabel = 'Save'
	}: {
		values?: Partial<CourseInput>;
		teachers: Teacher[];
		action?: string;
		submitLabel?: string;
	} = $props();
</script>

<form method="POST" {action} class="card grid max-w-lg gap-4" use:enhance>
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
