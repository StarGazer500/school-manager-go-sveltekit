<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { dismiss, flash } from './flash.svelte.ts';

	// Each new page starts without the previous page's message.
	afterNavigate(dismiss);
</script>

{#if flash.message}
	<div
		class="mb-4 flex items-center justify-between gap-3 rounded-md border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800"
		role="status"
	>
		<span>✅ {flash.message}</span>
		<button type="button" class="cursor-pointer text-green-700 hover:text-green-900" aria-label="Dismiss" onclick={dismiss}>
			✕
		</button>
	</div>
{/if}
