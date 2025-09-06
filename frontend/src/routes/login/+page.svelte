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
	let loading = false;
	let error = '';
	let showPassword = false;

	onMount(() => {
		if ($isAuthenticated) {
			goto('/dashboard');
		}
	});

	async function handleSubmit() {
		error = '';
		loading = true;

		try {
			if (!email.trim()) {
				error = 'Email is required';
				loading = false;
				return;
			}

			if (!password.trim()) {
				error = 'Password is required';
				loading = false;
				return;
			}

			const response = await apiClient.login(email, password);
			setUser(response.user, response.token);
			goto('/dashboard');
		} catch (err) {
			error = err.response?.data?.error || 'Login failed. Please check your credentials.';
		} finally {
			loading = false;
		}
	}

	function togglePasswordVisibility() {
		showPassword = !showPassword;
	}

	function goToRegister() {
		goto('/register');
	}
</script>

<svelte:head>
	<title>Login - Database Manager Pro</title>
</svelte:head>

<div class="login-container">
	<!-- Background Elements -->
	<div class="background-shapes">
		<div class="shape shape-1"></div>
		<div class="shape shape-2"></div>
		<div class="shape shape-3"></div>
		<div class="shape shape-4"></div>
	</div>

	<!-- Left Side - Branding -->
	<div class="branding-section">
		<div class="branding-content">
			<div class="logo-section">
				<div class="logo-icon">
					<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 3c7.2 0 9 1.8 9 9s-1.8 9-9 9-9-1.8-9-9 1.8-9 9-9z"/>
						<path d="M8 11h8"/>
						<path d="M8 15h8"/>
						<path d="M8 7h8"/>
					</svg>
				</div>
				<h1 class="brand-name">Database Manager Pro</h1>
			</div>
			
			<div class="hero-content">
				<h2 class="hero-title">Manage Your Databases with Confidence</h2>
				<p class="hero-description">
					Powerful database management tools that make complex operations simple. 
					Connect, explore, and manage your data with our intuitive interface.
				</p>
				
				<div class="features">
					<div class="feature">
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20,6 9,17 4,12"/>
						</svg>
						<span>Multi-database support</span>
					</div>
					<div class="feature">
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20,6 9,17 4,12"/>
						</svg>
						<span>Real-time collaboration</span>
					</div>
					<div class="feature">
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20,6 9,17 4,12"/>
						</svg>
						<span>Advanced security</span>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Right Side - Login Form -->
	<div class="form-section">
		<div class="form-container">
			<div class="form-header">
				<h2 class="form-title">Welcome Back</h2>
				<p class="form-subtitle">Sign in to your account to continue</p>
			</div>

			{#if error}
				<Alert type="error" message={error} />
			{/if}

			<form on:submit|preventDefault={handleSubmit} class="login-form">
				<div class="input-group">
					<label for="email" class="input-label">Email Address</label>
					<div class="input-wrapper">
						<svg class="input-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
							<polyline points="22,6 12,13 2,6"/>
						</svg>
						<input
							id="email"
							type="email"
							bind:value={email}
							placeholder="Enter your email"
							required
							disabled={loading}
							class="form-input"
						/>
					</div>
				</div>

				<div class="input-group">
					<label for="password" class="input-label">Password</label>
					<div class="input-wrapper">
						<svg class="input-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
							<circle cx="12" cy="16" r="1"/>
							<path d="M7 11V7a5 5 0 0 1 10 0v4"/>
						</svg>
						{#if showPassword}
							<input
								id="password"
								type="text"
								bind:value={password}
								placeholder="Enter your password"
								required
								disabled={loading}
								class="form-input"
							/>
						{:else}
							<input
								id="password"
								type="password"
								bind:value={password}
								placeholder="Enter your password"
								required
								disabled={loading}
								class="form-input"
							/>
						{/if}
						<button
							type="button"
							class="password-toggle"
							on:click={togglePasswordVisibility}
							disabled={loading}
						>
							{#if showPassword}
								<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
									<path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
									<path d="M14.12 14.12a3 3 0 1 1-4.24-4.24"/>
									<line x1="1" y1="1" x2="23" y2="23"/>
								</svg>
							{:else}
								<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
									<circle cx="12" cy="12" r="3"/>
								</svg>
							{/if}
						</button>
					</div>
				</div>

				<div class="form-actions">
					<button
						type="submit"
						disabled={loading}
						class="submit-btn"
					>
						{#if loading}
							<svg class="spinner" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 12a9 9 0 11-6.219-8.56"/>
							</svg>
							Signing In...
						{:else}
							<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/>
								<polyline points="10,17 15,12 10,7"/>
								<line x1="15" y1="12" x2="3" y2="12"/>
							</svg>
							Sign In
						{/if}
					</button>
				</div>
			</form>

			<div class="form-footer">
				<p class="footer-text">
					Don't have an account?
					<button type="button" class="link-btn" on:click={goToRegister} disabled={loading}>
						Create Account
					</button>
				</p>
			</div>
		</div>
	</div>
</div>

<style>
	* {
		box-sizing: border-box;
	}

	.login-container {
		min-height: 100vh;
		display: grid;
		grid-template-columns: 1fr 1fr;
		position: relative;
		overflow: hidden;
	}

	/* Background Shapes */
	.background-shapes {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		pointer-events: none;
		z-index: 1;
	}

	.shape {
		position: absolute;
		border-radius: 50%;
		background: linear-gradient(135deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));
		animation: float 8s ease-in-out infinite;
	}

	.shape-1 {
		width: 300px;
		height: 300px;
		top: 10%;
		left: 5%;
		animation-delay: 0s;
	}

	.shape-2 {
		width: 200px;
		height: 200px;
		top: 60%;
		right: 10%;
		animation-delay: 2s;
	}

	.shape-3 {
		width: 150px;
		height: 150px;
		bottom: 20%;
		left: 70%;
		animation-delay: 4s;
	}

	.shape-4 {
		width: 100px;
		height: 100px;
		top: 30%;
		left: 80%;
		animation-delay: 6s;
	}

	@keyframes float {
		0%, 100% { transform: translateY(0px) rotate(0deg) scale(1); }
		33% { transform: translateY(-20px) rotate(120deg) scale(1.1); }
		66% { transform: translateY(10px) rotate(240deg) scale(0.9); }
	}

	/* Branding Section */
	.branding-section {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		position: relative;
		z-index: 2;
	}

	.branding-content {
		max-width: 500px;
		color: white;
		text-align: center;
	}

	.logo-section {
		margin-bottom: 3rem;
	}

	.logo-icon {
		background: rgba(255, 255, 255, 0.2);
		backdrop-filter: blur(10px);
		border-radius: 16px;
		padding: 1rem;
		display: inline-block;
		margin-bottom: 1rem;
	}

	.logo-icon svg {
		color: white;
	}

	.brand-name {
		font-size: 2rem;
		font-weight: 700;
		margin: 0;
		background: linear-gradient(45deg, #ffffff, #e0e7ff);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.hero-title {
		font-size: 2.5rem;
		font-weight: 700;
		margin-bottom: 1rem;
		line-height: 1.2;
	}

	.hero-description {
		font-size: 1.125rem;
		line-height: 1.6;
		margin-bottom: 2rem;
		opacity: 0.9;
	}

	.features {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		align-items: flex-start;
		text-align: left;
	}

	.feature {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 1rem;
	}

	.feature svg {
		color: #a7f3d0;
		flex-shrink: 0;
	}

	/* Form Section */
	.form-section {
		background: #ffffff;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		position: relative;
		z-index: 2;
	}

	.form-container {
		width: 100%;
		max-width: 400px;
	}

	.form-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.form-title {
		font-size: 2rem;
		font-weight: 700;
		color: #1f2937;
		margin-bottom: 0.5rem;
	}

	.form-subtitle {
		color: #6b7280;
		font-size: 1rem;
		line-height: 1.5;
	}

	.login-form {
		margin-bottom: 2rem;
	}

	.input-group {
		margin-bottom: 1.5rem;
	}

	.input-label {
		display: block;
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.input-wrapper {
		position: relative;
	}

	.input-icon {
		position: absolute;
		left: 1rem;
		top: 50%;
		transform: translateY(-50%);
		color: #9ca3af;
		pointer-events: none;
	}

	.form-input {
		width: 100%;
		padding: 0.875rem 1rem 0.875rem 3rem;
		border: 2px solid #e5e7eb;
		border-radius: 12px;
		font-size: 1rem;
		transition: all 0.2s ease;
		background: #ffffff;
	}

	.form-input:focus {
		outline: none;
		border-color: #667eea;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
		background: #ffffff;
	}

	.form-input:disabled {
		background: #f9fafb;
		cursor: not-allowed;
	}

	.password-toggle {
		position: absolute;
		right: 1rem;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: #9ca3af;
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 6px;
		transition: all 0.2s ease;
	}

	.password-toggle:hover {
		color: #667eea;
		background: rgba(102, 126, 234, 0.1);
	}

	.password-toggle:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.form-actions {
		margin-bottom: 1.5rem;
	}

	.submit-btn {
		width: 100%;
		padding: 1rem 1.5rem;
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border: none;
		border-radius: 12px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		box-shadow: 0 4px 14px rgba(102, 126, 234, 0.3);
	}

	.submit-btn:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
	}

	.submit-btn:disabled {
		opacity: 0.7;
		cursor: not-allowed;
		transform: none;
	}

	.spinner {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.form-footer {
		text-align: center;
	}

	.footer-text {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0;
	}

	.link-btn {
		background: none;
		border: none;
		color: #667eea;
		font-weight: 600;
		cursor: pointer;
		padding: 0;
		margin-left: 0.25rem;
		text-decoration: none;
		transition: all 0.2s ease;
	}

	.link-btn:hover {
		color: #5a67d8;
		text-decoration: underline;
	}

	.link-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Alert Styling */
	:global(.alert) {
		margin-bottom: 1.5rem;
	}

	/* Responsive Design */
	@media (max-width: 1024px) {
		.login-container {
			grid-template-columns: 1fr;
		}

		.branding-section {
			display: none;
		}

		.form-section {
			padding: 2rem 1rem;
		}
	}

	@media (max-width: 640px) {
		.form-container {
			max-width: none;
		}

		.hero-title {
			font-size: 2rem;
		}

		.form-title {
			font-size: 1.75rem;
		}

		.brand-name {
			font-size: 1.5rem;
		}

		.branding-content {
			padding: 2rem 1rem;
		}
	}

	@media (max-width: 480px) {
		.form-section {
			padding: 1rem;
		}

		.form-input {
			padding: 0.75rem 0.875rem 0.75rem 2.75rem;
		}

		.submit-btn {
			padding: 0.875rem 1.25rem;
		}
	}
</style>
