<script lang="ts">
	import { goto } from '$app/navigation';
	import FormError from '#lib/ui/FormError.svelte';
	import { showMessage } from '#lib/ui/flash.svelte.ts';
	import { readCourses } from '#lib/services/courses/index.ts';
	import type { Course } from '#lib/services/courses/types.ts';
	import { enroll, readEnrollments, withdraw } from '#lib/services/enrollments/index.ts';
	import type { Enrollment } from '#lib/services/enrollments/types.ts';
	import { deleteStudent, readStudent, updateStudent } from '#lib/services/students/index.ts';
	import { fullName, type Student, type StudentInput } from '#lib/services/students/types.ts';
	import EnrollForm from '../enrollments/EnrollForm.svelte';
	import EnrollmentTable from '../enrollments/EnrollmentTable.svelte';
	import StudentForm from './StudentForm.svelte';

	let { id }: { id: number } = $props();

	let student = $state<Student | null>(null);
	let enrollments = $state<Enrollment[]>([]);
	let courses = $state<Course[]>([]);
	let busy = $state(true);
	let error = $state('');
	let saved = $state(false);

	// Only offer courses the student isn't already in.
	const availableCourses = $derived.by(() => {
		const enrolled = new Set(enrollments.map((e) => e.courseId));
		return courses.filter((c) => !enrolled.has(c.id));
	});

	async function load(studentId: number) {
		const [s, e, c] = await Promise.all([
			readStudent(studentId),
			readEnrollments({ studentId }),
			readCourses()
		]);
		busy = false;
		if (!s.success || !e.success || !c.success) {
			error = (!s.success ? s : !e.success ? e : c).message;
			return;
		}
		student = s.data;
		enrollments = e.data;
		courses = c.data;
	}

	// Runs on mount and again if the route moves to another student's id.
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

	async function save(input: StudentInput) {
		saved = false;
		saved = await after(await updateStudent(id, input));
	}

	async function remove() {
		if (!confirm('Delete this student?')) return;
		const ret = await deleteStudent(id);
		if (!ret.success) {
			error = ret.message;
			return;
		}
		await goto('/students');
		showMessage(ret.message); // after goto, so the list page shows it
	}
</script>

<FormError message={error} />
{#if busy}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if student}
	<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
		<h1 class="text-2xl font-bold">{fullName(student)}</h1>
		<button type="button" class="btn-danger" onclick={remove}>Delete student</button>
	</div>

	{#if saved}<p class="mb-4 text-sm text-green-700">✅ Saved.</p>{/if}

	<StudentForm values={student} onsave={save} submitLabel="Save changes" />

	<h2 class="mt-8 mb-3 text-lg font-semibold">Courses</h2>
	<EnrollForm
		label="Choose a course"
		options={availableCourses.map((c) => ({ value: c.id, label: `${c.code} — ${c.name}` }))}
		onenroll={async (courseId) => after(await enroll({ studentId: id, courseId }))}
	/>
	<EnrollmentTable {enrollments} show="course" onwithdraw={async (e) => after(await withdraw(e.id))} />
{:else}
	<a class="link" href="/students">Back to students</a>
{/if}
