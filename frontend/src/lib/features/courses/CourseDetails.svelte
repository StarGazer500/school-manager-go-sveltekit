<script lang="ts">
	import { goto } from '$app/navigation';
	import FormError from '#lib/ui/FormError.svelte';
	import { showMessage } from '#lib/ui/flash.svelte.ts';
	import { deleteCourse, readCourse, updateCourse } from '#lib/services/courses/index.ts';
	import type { Course, CourseInput } from '#lib/services/courses/types.ts';
	import { enroll, readEnrollments, withdraw } from '#lib/services/enrollments/index.ts';
	import type { Enrollment } from '#lib/services/enrollments/types.ts';
	import { readStudents } from '#lib/services/students/index.ts';
	import { fullName, type Student } from '#lib/services/students/types.ts';
	import { readTeachers } from '#lib/services/teachers/index.ts';
	import type { Teacher } from '#lib/services/teachers/types.ts';
	import EnrollForm from '../enrollments/EnrollForm.svelte';
	import EnrollmentTable from '../enrollments/EnrollmentTable.svelte';
	import CourseForm from './CourseForm.svelte';

	let { id }: { id: number } = $props();

	let course = $state<Course | null>(null);
	let enrollments = $state<Enrollment[]>([]);
	let teachers = $state<Teacher[]>([]);
	let students = $state<Student[]>([]);
	let busy = $state(true);
	let error = $state('');
	let saved = $state(false);

	const seatsLeft = $derived(course ? course.capacity - enrollments.length : 0);

	// Only offer students who aren't already on the roster.
	const availableStudents = $derived.by(() => {
		const enrolled = new Set(enrollments.map((e) => e.studentId));
		return students.filter((s) => !enrolled.has(s.id));
	});

	async function load(courseId: number) {
		const [c, e, t, s] = await Promise.all([
			readCourse(courseId),
			readEnrollments({ courseId }),
			readTeachers(),
			readStudents()
		]);
		busy = false;
		if (!c.success || !e.success || !t.success || !s.success) {
			error = (!c.success ? c : !e.success ? e : !t.success ? t : s).message;
			return;
		}
		course = c.data;
		enrollments = e.data;
		teachers = t.data;
		students = s.data;
	}

	// Runs on mount and again if the route moves to another course's id.
	$effect(() => {
		load(id);
	});

	/** Shows the message of a failed call, or reloads the page data after a successful one. */
	async function after(ret: { success: boolean; message: string }) {
		error = ret.success ? '' : ret.message;
		if (ret.success && ret.message) showMessage(ret.message);
		if (ret.success) await load(id);
		return ret.success;
	}

	async function save(input: CourseInput) {
		saved = false;
		saved = await after(await updateCourse(id, input));
	}

	async function remove() {
		if (!confirm('Delete this course and its enrollments?')) return;
		const ret = await deleteCourse(id);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		await goto('/courses');
		showMessage(ret.message); // after goto, so the list page shows it
	}
</script>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if course}
	<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
		<h1 class="text-2xl font-bold">
			<span class="font-mono">{course.code}</span>
			<span class="text-gray-500">·</span>
			{course.name}
		</h1>
		<button type="button" class="btn-danger" onclick={remove}>Delete course</button>
	</div>

	{#if saved}<p class="mb-4 text-sm text-green-700">✅ Saved.</p>{/if}

	<CourseForm values={course} {teachers} onsave={save} submitLabel="Save changes" />

	<h2 class="mt-8 mb-3 flex items-center gap-3 text-lg font-semibold">
		Roster
		<span
			class={[
				'rounded-full px-2.5 py-0.5 text-xs font-medium',
				seatsLeft > 0 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
			]}
		>
			{enrollments.length}/{course.capacity}
		</span>
	</h2>
	{#if seatsLeft > 0}
		<EnrollForm
			label="Choose a student"
			options={availableStudents.map((s) => ({ value: s.id, label: `${fullName(s)} (grade ${s.gradeLevel})` }))}
			onenroll={async (studentId) => after(await enroll({ studentId, courseId: id }))}
		/>
	{:else}
		<p class="mb-4 text-sm text-gray-500 italic">Course is full.</p>
	{/if}
	<EnrollmentTable {enrollments} show="student" onwithdraw={async (e) => after(await withdraw(e.id))} />
{:else}
	<a class="link" href="/courses">Back to courses</a>
{/if}
