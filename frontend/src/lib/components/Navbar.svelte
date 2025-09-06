<script>
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isAuthenticated, user } from '$lib/stores';

	export let currentPage = '';

	function handleLogout() {
		localStorage.removeItem('token');
		$isAuthenticated = false;
		goto('/login');
	}

	// Get current route for active menu
	$: currentRoute = $page.url.pathname;
</script>

<nav class="navbar">
	<div class="container">
		<a href="/dashboard" class="navbar-brand">Database Manager</a>
		<div class="navbar-nav">
			<a href="/dashboard" class:active={currentRoute === '/dashboard'}>Dashboard</a>
			<a href="/connections" class:active={currentRoute === '/connections'}>Connections</a>
			<a href="/api-management" class:active={currentRoute === '/api-management'}>API Management</a>
			<a href="/database-management" class:active={currentRoute.startsWith('/database-management')}>Database Management</a>
			<button class="btn btn-secondary" on:click={handleLogout}>Logout</button>
		</div>
	</div>
</nav>

<style>
	.navbar {
		background: linear-gradient(135deg, #667eea, #764ba2);
		padding: 1rem 0;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.navbar-brand {
		color: white;
		font-size: 1.5rem;
		font-weight: 700;
		text-decoration: none;
		transition: opacity 0.2s;
	}

	.navbar-brand:hover {
		opacity: 0.9;
	}

	.navbar-nav {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}

	.navbar-nav a {
		color: rgba(255, 255, 255, 0.9);
		text-decoration: none;
		font-weight: 500;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		transition: all 0.2s;
		position: relative;
	}

	.navbar-nav a:hover {
		color: white;
		background: rgba(255, 255, 255, 0.1);
	}

	.navbar-nav a.active {
		color: white;
		background: rgba(255, 255, 255, 0.2);
		font-weight: 600;
	}

	.btn {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 6px;
		font-size: 0.9rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-secondary {
		background: rgba(255, 255, 255, 0.2);
		color: white;
		border: 1px solid rgba(255, 255, 255, 0.3);
	}

	.btn-secondary:hover {
		background: rgba(255, 255, 255, 0.3);
		border-color: rgba(255, 255, 255, 0.5);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.container {
			flex-direction: column;
			gap: 1rem;
		}

		.navbar-nav {
			flex-wrap: wrap;
			gap: 0.75rem;
			justify-content: center;
		}

		.navbar-nav a {
			padding: 0.4rem 0.8rem;
			font-size: 0.9rem;
		}
	}
</style>
