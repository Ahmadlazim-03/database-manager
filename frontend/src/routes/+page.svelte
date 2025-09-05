<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores';

	onMount(() => {
		// Redirect to dashboard if authenticated, otherwise to login
		if ($isAuthenticated) {
			goto('/dashboard');
		} else {
			goto('/login');
		}
	});
</script>

<div class="loading">
	<div class="spinner"></div>
</div>
