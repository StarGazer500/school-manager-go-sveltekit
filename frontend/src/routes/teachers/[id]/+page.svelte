<script lang="ts">
	import { page } from '$app/state';
	import TeacherDetails from '#lib/features/teachers/TeacherDetails.svelte';
</script>

<TeacherDetails id={Number(page.params.id)} />
