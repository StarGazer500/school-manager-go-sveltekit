<script lang="ts">
	import { page } from '$app/state';
	import StudentDetails from '#lib/features/students/StudentDetails.svelte';
</script>

<StudentDetails id={Number(page.params.id)} />
