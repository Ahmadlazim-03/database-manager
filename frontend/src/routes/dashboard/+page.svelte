<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api';
	import { user, logout, isAuthenticated, connections } from '$lib/stores';
	import Navbar from '$lib/components/Navbar.svelte';

	let stats = {
		totalConnections: 0,
		activeConnections: 0,
		totalEndpoints: 0,
		activeEndpoints: 0,
		totalAPIKeys: 0,
		activeAPIKeys: 0,
		totalLogs: 0,
		sharedDatabases: 0
	};
	let recentLogs = [];
	let sharedDatabases = [];
	let pendingInvitations = [];
	let loading = true;

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		await loadDashboardData();
	});

	async function loadDashboardData() {
		try {
			// Load connections
			const connectionsData = await apiClient.getConnections();
			connections.set(connectionsData);
			stats.totalConnections = connectionsData.length;
			stats.activeConnections = connectionsData.filter(c => c.status === 'active').length;

			// Load API keys
			const apiKeys = await apiClient.getAPIKeys();
			stats.totalAPIKeys = apiKeys.length;
			stats.activeAPIKeys = apiKeys.filter(k => k.is_active).length;

			// Load endpoints
			const endpoints = await apiClient.getEndpoints();
			stats.totalEndpoints = endpoints.length;
			stats.activeEndpoints = endpoints.filter(e => e.is_active).length;

			// Load recent logs
			const logs = await apiClient.getLogs();
			recentLogs = logs.slice(0, 10);
			stats.totalLogs = logs.length;

			// Load shared databases
			const sharedData = await apiClient.getSharedDatabases();
			sharedDatabases = sharedData;
			stats.sharedDatabases = sharedData.length;

			// Load pending invitations
			const pendingData = await apiClient.getPendingInvitations();
			pendingInvitations = pendingData;

		} catch (error) {
			console.error('Error loading dashboard data:', error);
		} finally {
			loading = false;
		}
	}

	function manageSharedDatabase(shared) {
		// Store the selected database ID and permission in localStorage for database management page
		localStorage.setItem('selectedDatabaseId', shared.database.id);
		localStorage.setItem('selectedDatabasePermission', shared.permission_level);
		localStorage.setItem('isSharedDatabase', 'true');
		
		// Redirect to database management with query parameters
		goto(`/database-management?db=${shared.database.id}&shared=true&permission=${shared.permission_level}`);
	}

	async function acceptInvitation(token) {
		try {
			await apiClient.acceptInvitation(token);
			// Reload dashboard data to update lists
			await loadDashboardData();
		} catch (error) {
			console.error('Error accepting invitation:', error);
			alert('Failed to accept invitation: ' + (error.response?.data?.error || error.message));
		}
	}
</script>

<svelte:head>
	<title>Dashboard - Database Manager</title>
</svelte:head>

<Navbar />

<div class="container">
	<div class="page-header">
		<h1>Dashboard</h1>
		<p>Welcome back, {$user?.email}</p>
	</div>

	{#if pendingInvitations.length > 0}
		<div class="pending-invitations-alert">
			<div class="alert-content">
				<div class="alert-icon">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
					</svg>
				</div>
				<div class="alert-text">
					<h3>You have {pendingInvitations.length} pending database invitation{pendingInvitations.length > 1 ? 's' : ''}!</h3>
					<p>Someone has shared a database with you.</p>
				</div>
				<div class="alert-actions">
					{#each pendingInvitations as invitation}
						<button 
							class="btn btn-primary btn-sm"
							on:click={() => acceptInvitation(invitation.invitation_token)}
						>
							Accept from {invitation.inviter.email}
						</button>
					{/each}
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="loading">
			<div class="spinner"></div>
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-4">
			<div class="stat-card">
				<div class="stat-icon stat-icon-blue">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
						<path d="M4 6h16v2H4zm0 5h16v2H4zm0 5h16v2H4z"/>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-number">{stats.totalConnections}</div>
					<div class="stat-label">Total Connections</div>
					<div class="stat-detail">{stats.activeConnections} active</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-icon stat-icon-green">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-number">{stats.totalAPIKeys}</div>
					<div class="stat-label">API Keys</div>
					<div class="stat-detail">{stats.activeAPIKeys} active</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-icon stat-icon-purple">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
						<path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-number">{stats.totalEndpoints}</div>
					<div class="stat-label">Endpoints</div>
					<div class="stat-detail">{stats.activeEndpoints} active</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-icon stat-icon-orange">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
						<path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 3c1.93 0 3.5 1.57 3.5 3.5S13.93 13 12 13s-3.5-1.57-3.5-3.5S10.07 6 12 6zm7 13H5v-.5c0-.67.25-1.3.71-1.77C6.73 15.81 7.85 15 12 15s5.27.81 6.29 1.73c.46.47.71 1.1.71 1.77V19z"/>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-number">{stats.totalLogs}</div>
					<div class="stat-label">API Calls</div>
					<div class="stat-detail">Total requests</div>
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="grid grid-3" style="margin-bottom: 2rem;">
			<div class="action-card">
				<div class="action-icon">
					🔗
				</div>
				<h3>Database Connections</h3>
				<p>Connect to MongoDB, MySQL, PostgreSQL, and more</p>
				<a href="/connections" class="btn btn-primary">Manage Connections</a>
			</div>
			<div class="action-card">
				<div class="action-icon">
					⚡
				</div>
				<h3>API Management</h3>
				<p>Create and manage REST API endpoints</p>
				<a href="/api-management" class="btn btn-primary">Manage APIs</a>
			</div>
			<div class="action-card">
				<div class="action-icon">
					📊
				</div>
				<h3>Database Management</h3>
				<p>View and manage your database collections</p>
				<a href="/database-management" class="btn btn-primary">Manage Data</a>
			</div>
		</div>

		<!-- Recent Activity -->
		<div class="grid grid-2">
			<div class="card">
				<h2>Recent Connections</h2>
				{#if $connections.length === 0}
					<div class="empty-state">
						<p>No database connections yet</p>
						<a href="/connections" class="btn">Add Connection</a>
					</div>
				{:else}
					<div class="connection-list">
						{#each $connections.slice(0, 5) as connection}
							<div class="connection-item">
								<div class="connection-info">
									<div class="connection-name">{connection.name}</div>
									<div class="connection-details">
										{connection.type} • {connection.host}:{connection.port}
									</div>
								</div>
								<div class="connection-status status-{connection.status}">
									{connection.status}
								</div>
							</div>
						{/each}
					</div>
					<a href="/connections" class="btn btn-secondary">View All</a>
				{/if}
			</div>

			<div class="card">
				<h2>Shared Databases</h2>
				{#if sharedDatabases.length === 0}
					<div class="empty-state">
						<p>No shared databases</p>
						<small>Databases shared with you will appear here</small>
					</div>
				{:else}
					<div class="shared-list">
						{#each sharedDatabases.slice(0, 5) as shared}
							<div class="shared-item">
								<div class="shared-info">
									<div class="shared-name">{shared.database.name}</div>
									<div class="shared-details">
										Shared by {shared.grantor.email} • {shared.permission_level} access
									</div>
								</div>
								<div class="shared-actions">
									<div class="shared-type">{shared.database.type}</div>
									<button 
										class="btn btn-sm btn-primary" 
										on:click={() => manageSharedDatabase(shared)}
									>
										Manage
									</button>
								</div>
							</div>
						{/each}
					</div>
					{#if sharedDatabases.length > 5}
						<button class="btn btn-secondary" on:click={() => goto('/database-management')}>
							View All Shared Databases
						</button>
					{/if}
				{/if}
			</div>
		</div>

		<!-- Additional Activity Section -->
		<div class="grid grid-2">
			<div class="card">
				<h2>Recent API Activity</h2>
				{#if recentLogs.length === 0}
					<div class="empty-state">
						<p>No API activity yet</p>
					</div>
				{:else}
					<div class="log-list">
						{#each recentLogs as log}
							<div class="log-item">
								<div class="log-info">
									<div class="log-path">{log.method} {log.path}</div>
									<div class="log-time">
										{new Date(log.created_at).toLocaleString()}
									</div>
								</div>
								<div class="log-status status-code-{Math.floor(log.status_code / 100)}">
									{log.status_code}
								</div>
							</div>
						{/each}
					</div>
					<a href="/api-management?tab=logs" class="btn btn-secondary">View All Logs</a>
				{/if}
			</div>

			<div class="card">
				<h2>Quick Stats</h2>
				<div class="quick-stats">
					<div class="quick-stat">
						<span class="stat-value">{stats.sharedDatabases}</span>
						<span class="stat-label">Shared Databases</span>
					</div>
					<div class="quick-stat">
						<span class="stat-value">{stats.totalLogs}</span>
						<span class="stat-label">Total API Calls</span>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.page-header {
		margin-bottom: 32px;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: #333;
		margin-bottom: 8px;
	}

	.page-header p {
		color: #666;
		font-size: 1.1rem;
	}

	.grid-4 {
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		margin-bottom: 32px;
	}

	.stat-card {
		background: white;
		border-radius: 12px;
		padding: 24px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.stat-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
	}

	.stat-icon-blue { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
	.stat-icon-green { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
	.stat-icon-purple { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
	.stat-icon-orange { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }

	.stat-content {
		flex: 1;
	}

	.stat-number {
		font-size: 2rem;
		font-weight: 700;
		color: #333;
		line-height: 1;
	}

	.stat-label {
		font-size: 0.9rem;
		color: #666;
		margin-top: 4px;
	}

	.stat-detail {
		font-size: 0.8rem;
		color: #999;
		margin-top: 2px;
	}

	.connection-list, .log-list {
		margin: 20px 0;
	}

	.connection-item, .log-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 0;
		border-bottom: 1px solid #eee;
	}

	.connection-item:last-child, .log-item:last-child {
		border-bottom: none;
	}

	.connection-name, .log-path {
		font-weight: 600;
		color: #333;
	}

	.connection-details, .log-time {
		font-size: 0.9rem;
		color: #666;
		margin-top: 2px;
	}

	.connection-status, .log-status {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.status-active {
		background: #d4edda;
		color: #155724;
	}

	.status-inactive {
		background: #f8d7da;
		color: #721c24;
	}

	.status-code-2 {
		background: #d4edda;
		color: #155724;
	}

	.status-code-4, .status-code-5 {
		background: #f8d7da;
		color: #721c24;
	}

	.empty-state {
		text-align: center;
		padding: 40px 20px;
		color: #666;
	}

	.empty-state p {
		margin-bottom: 16px;
	}

	/* Action Cards */
	.action-card {
		background: white;
		border-radius: 16px;
		padding: 2rem;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
		border: 2px solid #f1f5f9;
		text-align: center;
		transition: all 0.3s;
	}

	.action-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
		border-color: #667eea;
	}

	.action-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.action-card h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #333;
		margin: 0 0 0.5rem 0;
	}

	.action-card p {
		color: #666;
		margin: 0 0 1.5rem 0;
		font-size: 0.9rem;
	}

	.grid-3 {
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
	}

	.btn-primary {
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 500;
		text-decoration: none;
		display: inline-block;
		transition: all 0.2s;
	}

	.btn-primary:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
	}

	/* Shared Databases Styles */
	.shared-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-bottom: 16px;
	}

	.shared-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px;
		background: #f8fafc;
		border-radius: 8px;
		border: 1px solid #e2e8f0;
	}

	.shared-info {
		flex: 1;
	}

	.shared-actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.shared-name {
		font-weight: 600;
		color: #1e293b;
		margin-bottom: 4px;
	}

	.shared-details {
		font-size: 0.875rem;
		color: #64748b;
	}

	.shared-type {
		background: #e0f2fe;
		color: #0369a1;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
	}

	/* Quick Stats */
	.quick-stats {
		display: flex;
		gap: 2rem;
		justify-content: space-around;
		margin-top: 1rem;
	}

	.quick-stat {
		text-align: center;
	}

	.quick-stat .stat-value {
		display: block;
		font-size: 2rem;
		font-weight: 700;
		color: #667eea;
		margin-bottom: 0.25rem;
	}

	.quick-stat .stat-label {
		font-size: 0.875rem;
		color: #64748b;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	.btn-secondary {
		background: #f1f5f9;
		color: #475569;
		border: 1px solid #e2e8f0;
		text-decoration: none;
		display: inline-block;
		text-align: center;
		margin-top: 1rem;
	}

	.btn-secondary:hover {
		background: #e2e8f0;
		color: #334155;
	}

	/* Pending Invitations Alert */
	.pending-invitations-alert {
		background: linear-gradient(135deg, #fef3c7, #f59e0b);
		border: 1px solid #f59e0b;
		border-radius: 12px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		box-shadow: 0 4px 6px rgba(245, 158, 11, 0.1);
	}

	.alert-content {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.alert-icon {
		background: white;
		border-radius: 50%;
		width: 48px;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #f59e0b;
		flex-shrink: 0;
	}

	.alert-text {
		flex: 1;
	}

	.alert-text h3 {
		margin: 0 0 0.25rem 0;
		font-size: 1.125rem;
		font-weight: 600;
		color: #92400e;
	}

	.alert-text p {
		margin: 0;
		color: #92400e;
		opacity: 0.8;
	}

	.alert-actions {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}
</style>
