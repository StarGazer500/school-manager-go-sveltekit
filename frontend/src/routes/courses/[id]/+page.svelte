<script lang="ts">
	import { page } from '$app/state';
	import CourseDetails from '#lib/features/courses/CourseDetails.svelte';
</script>

<CourseDetails id={Number(page.params.id)} />
