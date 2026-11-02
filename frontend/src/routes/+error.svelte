<script lang="ts">
	import { page } from '$app/state';
</script>

<h1 class="mb-4 text-2xl font-bold">{page.status}</h1>
<p class="mb-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">{page.error?.message}</p>
{#if page.status >= 500}
	<p class="mb-4 text-sm">
		Is the Go API running? Start it with <code class="rounded bg-gray-100 px-1">go run ./cmd/api</code> in
		<code class="rounded bg-gray-100 px-1">backend/</code>.
	</p>
{/if}
<a class="link" href="/">Back to dashboard</a>
