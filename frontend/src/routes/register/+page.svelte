<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api';
	import { isAuthenticated } from '$lib/stores';
	import Alert from '$lib/components/Alert.svelte';

	let email = '';
	let password = '';
	let confirmPassword = '';
	let firstName = '';
	let lastName = '';
	let loading = false;
	let error = '';
	let success = '';
	let showPassword = false;
	let showConfirmPassword = false;

	onMount(() => {
		if ($isAuthenticated) {
			goto('/dashboard');
		}
	});

	async function handleSubmit() {
		error = '';
		success = '';
		loading = true;

		try {
			// Validation
			if (!firstName.trim()) {
				error = 'First name is required';
				loading = false;
				return;
			}

			if (!lastName.trim()) {
				error = 'Last name is required';
				loading = false;
				return;
			}

			if (!email.trim()) {
				error = 'Email is required';
				loading = false;
				return;
			}

			if (password.length < 6) {
				error = 'Password must be at least 6 characters';
				loading = false;
				return;
			}

			if (password !== confirmPassword) {
				error = 'Passwords do not match';
				loading = false;
				return;
			}

			const response = await apiClient.register({
				first_name: firstName,
				last_name: lastName,
				email: email,
				password: password
			});

			success = 'Account created successfully! Redirecting to login...';
			
			// Redirect to login after success
			setTimeout(() => {
				goto('/login');
			}, 2000);

		} catch (err) {
			error = err.response?.data?.error || 'Failed to create account';
		} finally {
			loading = false;
		}
	}

	function togglePasswordVisibility() {
		showPassword = !showPassword;
	}

	function toggleConfirmPasswordVisibility() {
		showConfirmPassword = !showConfirmPassword;
	}

	function goToLogin() {
		goto('/login');
	}
</script>

<svelte:head>
	<title>Register - Database Manager Pro</title>
</svelte:head>

<div class="register-container">
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
				<h2 class="hero-title">Join Thousands of Developers</h2>
				<p class="hero-description">
					Start your journey with the most powerful database management platform. 
					Create, manage, and scale your databases with confidence.
				</p>
				
				<div class="benefits">
					<div class="benefit">
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4"/>
							<circle cx="12" cy="12" r="10"/>
						</svg>
						<span>Free to start</span>
					</div>
					<div class="benefit">
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4"/>
							<circle cx="12" cy="12" r="10"/>
						</svg>
						<span>No credit card required</span>
					</div>
					<div class="benefit">
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4"/>
							<circle cx="12" cy="12" r="10"/>
						</svg>
						<span>24/7 support</span>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Right Side - Register Form -->
	<div class="form-section">
		<div class="form-container">
			<div class="form-header">
				<h2 class="form-title">Create Account</h2>
				<p class="form-subtitle">Get started with your free account</p>
			</div>

			{#if error}
				<Alert type="error" message={error} />
			{/if}

			{#if success}
				<Alert type="success" message={success} />
			{/if}

			<form on:submit|preventDefault={handleSubmit} class="register-form">
				<div class="name-row">
					<div class="input-group">
						<label for="firstName" class="input-label">First Name</label>
						<div class="input-wrapper">
							<svg class="input-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
								<circle cx="12" cy="7" r="4"/>
							</svg>
							<input
								id="firstName"
								type="text"
								bind:value={firstName}
								placeholder="First name"
								required
								disabled={loading}
								class="form-input"
							/>
						</div>
					</div>

					<div class="input-group">
						<label for="lastName" class="input-label">Last Name</label>
						<div class="input-wrapper">
							<svg class="input-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
								<circle cx="12" cy="7" r="4"/>
							</svg>
							<input
								id="lastName"
								type="text"
								bind:value={lastName}
								placeholder="Last name"
								required
								disabled={loading}
								class="form-input"
							/>
						</div>
					</div>
				</div>

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
								placeholder="Create a password (min 6 characters)"
								required
								disabled={loading}
								class="form-input"
							/>
						{:else}
							<input
								id="password"
								type="password"
								bind:value={password}
								placeholder="Create a password (min 6 characters)"
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

				<div class="input-group">
					<label for="confirmPassword" class="input-label">Confirm Password</label>
					<div class="input-wrapper">
						<svg class="input-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
							<circle cx="12" cy="16" r="1"/>
							<path d="M7 11V7a5 5 0 0 1 10 0v4"/>
						</svg>
						{#if showConfirmPassword}
							<input
								id="confirmPassword"
								type="text"
								bind:value={confirmPassword}
								placeholder="Confirm your password"
								required
								disabled={loading}
								class="form-input"
							/>
						{:else}
							<input
								id="confirmPassword"
								type="password"
								bind:value={confirmPassword}
								placeholder="Confirm your password"
								required
								disabled={loading}
								class="form-input"
							/>
						{/if}
						<button
							type="button"
							class="password-toggle"
							on:click={toggleConfirmPasswordVisibility}
							disabled={loading}
						>
							{#if showConfirmPassword}
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

				<div class="terms-section">
					<label class="terms-checkbox">
						<input type="checkbox" required disabled={loading} />
						<span class="checkmark"></span>
						<span class="terms-text">
							I agree to the <a href="/terms" target="_blank">Terms of Service</a> 
							and <a href="/privacy" target="_blank">Privacy Policy</a>
						</span>
					</label>
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
							Creating Account...
						{:else}
							<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
								<circle cx="8.5" cy="7" r="4"/>
								<line x1="20" y1="8" x2="20" y2="14"/>
								<line x1="23" y1="11" x2="17" y2="11"/>
							</svg>
							Create Account
						{/if}
					</button>
				</div>
			</form>

			<div class="form-footer">
				<p class="footer-text">
					Already have an account?
					<button type="button" class="link-btn" on:click={goToLogin} disabled={loading}>
						Sign In
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

	.register-container {
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
		width: 350px;
		height: 350px;
		top: 5%;
		left: 2%;
		animation-delay: 0s;
	}

	.shape-2 {
		width: 180px;
		height: 180px;
		top: 70%;
		right: 5%;
		animation-delay: 2s;
	}

	.shape-3 {
		width: 120px;
		height: 120px;
		bottom: 10%;
		left: 75%;
		animation-delay: 4s;
	}

	.shape-4 {
		width: 80px;
		height: 80px;
		top: 25%;
		left: 85%;
		animation-delay: 6s;
	}

	@keyframes float {
		0%, 100% { transform: translateY(0px) rotate(0deg) scale(1); }
		33% { transform: translateY(-25px) rotate(120deg) scale(1.1); }
		66% { transform: translateY(15px) rotate(240deg) scale(0.9); }
	}

	/* Branding Section */
	.branding-section {
		background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
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

	.benefits {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		align-items: flex-start;
		text-align: left;
	}

	.benefit {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 1rem;
	}

	.benefit svg {
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
		overflow-y: auto;
	}

	.form-container {
		width: 100%;
		max-width: 450px;
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

	.register-form {
		margin-bottom: 2rem;
	}

	.name-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-bottom: 1.5rem;
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

	.terms-section {
		margin: 1.5rem 0;
	}

	.terms-checkbox {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		cursor: pointer;
		line-height: 1.5;
	}

	.terms-checkbox input[type="checkbox"] {
		width: 18px;
		height: 18px;
		margin: 2px 0 0 0;
		cursor: pointer;
		accent-color: #667eea;
	}

	.terms-text {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.terms-text a {
		color: #667eea;
		text-decoration: none;
		font-weight: 500;
	}

	.terms-text a:hover {
		text-decoration: underline;
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
		.register-container {
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

		.name-row {
			grid-template-columns: 1fr;
			gap: 0;
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
