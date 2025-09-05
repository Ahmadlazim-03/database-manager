<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api';
	import { setUser, isAuthenticated } from '$lib/stores';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let email = '';
	let password = '';
	let confirmPassword = '';
	let isLogin = true;
	let loading = false;
	let error = '';

	onMount(() => {
		if ($isAuthenticated) {
			goto('/dashboard');
		}
		
		// Initialize AOS
		if (typeof AOS !== 'undefined') {
			AOS.init({
				duration: 800,
				easing: 'ease-in-out',
				once: true
			});
		}
		
		// Initialize Lucide icons
		if (typeof lucide !== 'undefined') {
			lucide.createIcons();
		}
	});

	async function handleSubmit() {
		error = '';
		loading = true;

		try {
			if (!isLogin && password !== confirmPassword) {
				error = 'Passwords do not match';
				loading = false;
				return;
			}

			let response;
			if (isLogin) {
				response = await apiClient.login(email, password);
			} else {
				response = await apiClient.register(email, password);
			}

			setUser(response.user, response.token);
			goto('/dashboard');
		} catch (err) {
			error = err.response?.data?.error || 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function toggleMode() {
		isLogin = !isLogin;
		error = '';
		confirmPassword = '';
	}
</script>

<svelte:head>
	<title>{isLogin ? 'Login' : 'Register'} - Database Manager Pro</title>
</svelte:head>

<div class="auth-container">
	<div class="auth-background">
		<div class="auth-shape auth-shape-1"></div>
		<div class="auth-shape auth-shape-2"></div>
		<div class="auth-shape auth-shape-3"></div>
	</div>
	
	<div class="auth-card" data-aos="fade-up" data-aos-delay="200">
		<div class="auth-header" data-aos="fade-down" data-aos-delay="400">
			<div class="auth-logo">
				<i data-lucide="database" class="auth-logo-icon"></i>
				<span class="auth-logo-text">Database Manager Pro</span>
			</div>
			<h1 class="auth-title">
				{isLogin ? 'Welcome back!' : 'Get started today'}
			</h1>
			<p class="auth-subtitle">
				{isLogin ? 'Sign in to your account to continue' : 'Create a new account to get started'}
			</p>
		</div>

		{#if error}
			<Alert type="error" title="Authentication Error" message={error} />
		{/if}

		<form on:submit|preventDefault={handleSubmit} data-aos="fade-up" data-aos-delay="600">
			<Input
				label="Email Address"
				type="email"
				icon="mail"
				bind:value={email}
				placeholder="Enter your email"
				required
				disabled={loading}
				floating={true}
			/>

			<Input
				label="Password"
				type="password"
				icon="lock"
				bind:value={password}
				placeholder="Enter your password"
				required
				disabled={loading}
				floating={true}
			/>

			{#if !isLogin}
				<Input
					label="Confirm Password"
					type="password"
					icon="lock"
					bind:value={confirmPassword}
					placeholder="Confirm your password"
					required
					disabled={loading}
					floating={true}
				/>
			{/if}

			<div class="auth-actions" data-aos="fade-up" data-aos-delay="800">
				<Button
					type="submit"
					variant="primary"
					size="lg"
					{loading}
					disabled={loading}
					icon={isLogin ? "log-in" : "user-plus"}
					style="width: 100%; justify-content: center;"
				>
					{isLogin ? 'Sign In' : 'Create Account'}
				</Button>
			</div>
		</form>

		<div class="auth-footer" data-aos="fade" data-aos-delay="900">
			<p class="footer-text">
				{isLogin ? "Don't have an account?" : 'Already have an account?'}
			</p>
			<Button
				type="button"
				variant="ghost"
				size="sm"
				on:click={toggleMode}
				disabled={loading}
			>
				{isLogin ? 'Create Account' : 'Sign In'}
			</Button>
		</div>
	</div>
</div>

<style>
	.auth-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		padding: 20px;
	}

	.auth-card {
		background: white;
		border-radius: 12px;
		padding: 48px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
		width: 100%;
		max-width: 400px;
	}

	.auth-header {
		text-align: center;
		margin-bottom: 32px;
	}

	.auth-title {
		font-size: 2rem;
		font-weight: 700;
		color: #333;
		margin-bottom: 8px;
	}

	.auth-subtitle {
		color: #666;
		font-size: 1rem;
	}

	.btn-primary {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		border: none;
		color: white;
		padding: 16px;
		border-radius: 8px;
		font-size: 16px;
		font-weight: 600;
		cursor: pointer;
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
	}

	.btn-primary:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.spinner-small {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top: 2px solid white;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.auth-footer {
		text-align: center;
		margin-top: 32px;
		padding-top: 24px;
		border-top: 1px solid #eee;
	}

	.link-button {
		background: none;
		border: none;
		color: #667eea;
		cursor: pointer;
		font-weight: 600;
		text-decoration: underline;
	}

	.link-button:hover {
		color: var(--color-primary-dark);
	}

	/* Auth Page Modern Styling */
	.auth-container {
		min-height: 100vh;
		background: linear-gradient(135deg, var(--gradient-primary), var(--gradient-secondary));
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-4);
		position: relative;
		overflow: hidden;
	}

	.auth-container::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: 
			radial-gradient(circle at 20% 80%, rgba(120, 119, 198, 0.3) 0%, transparent 50%),
			radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.3) 0%, transparent 50%);
		pointer-events: none;
	}

	.auth-card {
		background: var(--surface-glass);
		backdrop-filter: blur(20px);
		border: 1px solid var(--border-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-8);
		width: 100%;
		max-width: 480px;
		box-shadow: var(--shadow-lg);
		position: relative;
		z-index: 1;
	}

	.auth-header {
		text-align: center;
		margin-bottom: var(--space-8);
	}

	.auth-logo {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-3);
		margin-bottom: var(--space-6);
	}

	.auth-logo-icon {
		width: 40px;
		height: 40px;
		color: var(--color-primary);
	}

	.auth-logo-text {
		font-size: var(--text-xl);
		font-weight: var(--font-bold);
		color: var(--text-primary);
	}

	.auth-title {
		font-size: var(--text-3xl);
		font-weight: var(--font-bold);
		color: var(--text-primary);
		margin-bottom: var(--space-2);
		line-height: var(--leading-tight);
	}

	.auth-subtitle {
		font-size: var(--text-base);
		color: var(--text-secondary);
		line-height: var(--leading-relaxed);
	}

	.auth-actions {
		margin-top: var(--space-6);
	}

	.auth-footer {
		text-align: center;
		margin-top: var(--space-6);
		padding-top: var(--space-6);
		border-top: 1px solid var(--border-subtle);
	}

	.footer-text {
		color: var(--text-secondary);
		margin-bottom: var(--space-3);
		font-size: var(--text-sm);
	}

	/* Form Spacing */
	form :global(.input-group) {
		margin-bottom: var(--space-5);
	}

	form :global(.alert) {
		margin-bottom: var(--space-6);
	}

	/* Animation Enhancements */
	.auth-card {
		animation: slideIn 0.6s ease-out;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(30px) scale(0.95);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	/* Focus and Hover States */
	.auth-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-xl);
		transition: all 0.3s ease;
	}

	/* Responsive Design */
	@media (max-width: 640px) {
		.auth-container {
			padding: var(--space-3);
		}

		.auth-card {
			padding: var(--space-6);
		}

		.auth-title {
			font-size: var(--text-2xl);
		}

		.auth-logo-text {
			font-size: var(--text-lg);
		}
	}
</style>
