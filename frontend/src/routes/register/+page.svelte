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
	let firstName = '';
	let lastName = '';
	let loading = false;
	let error = '';
	let success = '';

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
</script>

<svelte:head>
	<title>Register - Database Manager Pro</title>
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
			<h1 class="auth-title">Create Account</h1>
			<p class="auth-subtitle">Join thousands of developers managing their databases efficiently</p>
		</div>

		{#if error}
			<Alert type="error" title="Registration Error" message={error} />
		{/if}

		{#if success}
			<Alert type="success" title="Success!" message={success} />
		{/if}

		<form on:submit|preventDefault={handleSubmit} data-aos="fade-up" data-aos-delay="600">
			<div class="form-row">
				<div class="form-col">
					<Input
						label="First Name"
						type="text"
						icon="user"
						bind:value={firstName}
						placeholder="Enter your first name"
						required
						disabled={loading}
						floating={true}
					/>
				</div>
				<div class="form-col">
					<Input
						label="Last Name"
						type="text"
						icon="user"
						bind:value={lastName}
						placeholder="Enter your last name"
						required
						disabled={loading}
						floating={true}
					/>
				</div>
			</div>

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
				placeholder="Create a password (min 6 characters)"
				required
				disabled={loading}
				floating={true}
			/>

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

			<div class="terms-section">
				<label class="terms-checkbox">
					<input type="checkbox" required />
					<span class="checkmark"></span>
					<span class="terms-text">
						I agree to the <a href="/terms" target="_blank">Terms of Service</a> 
						and <a href="/privacy" target="_blank">Privacy Policy</a>
					</span>
				</label>
			</div>

			<div class="auth-actions" data-aos="fade-up" data-aos-delay="800">
				<Button
					type="submit"
					variant="primary"
					size="lg"
					{loading}
					disabled={loading}
					icon="user-plus"
					style="width: 100%; justify-content: center;"
				>
					Create Account
				</Button>
			</div>
		</form>

		<div class="auth-footer" data-aos="fade" data-aos-delay="900">
			<p class="footer-text">Already have an account?</p>
			<Button
				type="button"
				variant="ghost"
				size="sm"
				on:click={() => goto('/login')}
				disabled={loading}
			>
				Sign In
			</Button>
		</div>
	</div>
</div>

<style>
	/* Auth Page Modern Styling */
	.auth-container {
		min-height: 100vh;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 20px;
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

	.auth-background {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		overflow: hidden;
		pointer-events: none;
	}

	.auth-shape {
		position: absolute;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.05);
		animation: float 6s ease-in-out infinite;
	}

	.auth-shape-1 {
		width: 300px;
		height: 300px;
		top: -150px;
		left: -150px;
		animation-delay: 0s;
	}

	.auth-shape-2 {
		width: 200px;
		height: 200px;
		top: 20%;
		right: -100px;
		animation-delay: 2s;
	}

	.auth-shape-3 {
		width: 400px;
		height: 400px;
		bottom: -200px;
		left: 30%;
		animation-delay: 4s;
	}

	@keyframes float {
		0%, 100% { transform: translateY(0px) rotate(0deg); }
		50% { transform: translateY(-20px) rotate(180deg); }
	}

	.auth-card {
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 16px;
		padding: 48px;
		width: 100%;
		max-width: 500px;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
		position: relative;
		z-index: 1;
	}

	.auth-header {
		text-align: center;
		margin-bottom: 32px;
	}

	.auth-logo {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		margin-bottom: 24px;
	}

	.auth-logo-icon {
		width: 40px;
		height: 40px;
		color: #667eea;
	}

	.auth-logo-text {
		font-size: 1.5rem;
		font-weight: 700;
		color: #333;
	}

	.auth-title {
		font-size: 2.5rem;
		font-weight: 700;
		color: #333;
		margin-bottom: 8px;
		line-height: 1.2;
	}

	.auth-subtitle {
		font-size: 1rem;
		color: #666;
		line-height: 1.5;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
		margin-bottom: 20px;
	}

	.form-col {
		min-width: 0;
	}

	.terms-section {
		margin: 24px 0;
	}

	.terms-checkbox {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		cursor: pointer;
		line-height: 1.4;
	}

	.terms-checkbox input[type="checkbox"] {
		width: 18px;
		height: 18px;
		margin: 0;
		cursor: pointer;
	}

	.terms-text {
		font-size: 0.9rem;
		color: #666;
	}

	.terms-text a {
		color: #667eea;
		text-decoration: none;
		font-weight: 500;
	}

	.terms-text a:hover {
		text-decoration: underline;
	}

	.auth-actions {
		margin-top: 32px;
	}

	.auth-footer {
		text-align: center;
		margin-top: 32px;
		padding-top: 24px;
		border-top: 1px solid #eee;
	}

	.footer-text {
		color: #666;
		margin-bottom: 12px;
		font-size: 0.9rem;
	}

	/* Form Spacing */
	form :global(.input-group) {
		margin-bottom: 20px;
	}

	form :global(.alert) {
		margin-bottom: 24px;
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
		box-shadow: 0 30px 60px rgba(0, 0, 0, 0.2);
		transition: all 0.3s ease;
	}

	/* Responsive Design */
	@media (max-width: 640px) {
		.auth-container {
			padding: 16px;
		}

		.auth-card {
			padding: 32px 24px;
		}

		.auth-title {
			font-size: 2rem;
		}

		.auth-logo-text {
			font-size: 1.25rem;
		}

		.form-row {
			grid-template-columns: 1fr;
			gap: 0;
		}

		.form-col {
			margin-bottom: 20px;
		}
	}

	@media (max-width: 480px) {
		.auth-card {
			padding: 24px 20px;
		}
		
		.auth-title {
			font-size: 1.75rem;
		}
	}
</style>
