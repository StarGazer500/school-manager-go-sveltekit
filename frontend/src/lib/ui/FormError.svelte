<script lang="ts">
	let { message }: { message?: string | null } = $props();
</script>

{#if message}
	<p class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700" role="alert">{message}</p>
{/if}
