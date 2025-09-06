<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api';
	import { isAuthenticated, connections, logout, user } from '$lib/stores';
	import Navbar from '$lib/components/Navbar.svelte';

	let showModal = false;
	let loading = false;
	let testingConnection = false;
	let error = '';
	let success = '';
	let selectedConnection = null;
	let databaseInfo = null;

	let newConnection = {
		name: '',
		type: 'mysql',
		host: '',
		port: 3306,
		database: '',
		username: '',
		password: ''
	};

	// Drag functionality variables
	let isDragging = false;
	let modalElement;
	let dragStartX = 0;
	let dragStartY = 0;
	let modalX = 0;
	let modalY = 0;
	let currentX = 0;
	let currentY = 0;

	// Drag functionality for manage modal
	let isDraggingManage = false;
	let manageModalElement;
	let manageDragStartX = 0;
	let manageDragStartY = 0;
	let manageCurrentX = 0;
	let manageCurrentY = 0;

	// Database configurations
	const databaseConfigs = {
		mysql: { defaultPort: 3306, requiresDatabase: true },
		postgres: { defaultPort: 5432, requiresDatabase: true },
		mongodb: { defaultPort: 27017, requiresDatabase: true },
		sqlite: { defaultPort: null, requiresDatabase: false },
		redis: { defaultPort: 6379, requiresDatabase: false },
		mariadb: { defaultPort: 3306, requiresDatabase: true },
		oracle: { defaultPort: 1521, requiresDatabase: true },
		sqlserver: { defaultPort: 1433, requiresDatabase: true },
		cassandra: { defaultPort: 9042, requiresDatabase: true },
		elasticsearch: { defaultPort: 9200, requiresDatabase: false },
		influxdb: { defaultPort: 8086, requiresDatabase: true },
		cockroachdb: { defaultPort: 26257, requiresDatabase: true }
	};

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}
		await loadConnections();
	});

	async function loadConnections() {
		loading = true;
		try {
			const data = await apiClient.getConnections();
			connections.set(data);
		} catch (err) {
			error = 'Failed to load connections';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		showModal = true;
		error = '';
		success = '';
		newConnection = {
			name: '',
			type: 'mysql',
			host: '',
			port: 3306,
			database: '',
			username: '',
			password: ''
		};
	}

	function closeModal() {
		showModal = false;
		resetModalPosition();
	}

	function updatePort() {
		const config = databaseConfigs[newConnection.type];
		if (config && config.defaultPort) {
			newConnection.port = config.defaultPort;
		}
	}

	// Drag functionality
	function startDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDragging = true;
			
			const rect = modalElement.getBoundingClientRect();
			dragStartX = e.clientX - rect.left;
			dragStartY = e.clientY - rect.top;
			
			// Store initial position for dragging
			currentX = rect.left;
			currentY = rect.top;
			
			// Apply initial position
			modalElement.style.left = currentX + 'px';
			modalElement.style.top = currentY + 'px';
			modalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', drag);
			document.addEventListener('mouseup', stopDrag);
			
			// Prevent text selection while dragging
			e.preventDefault();
			modalElement.style.userSelect = 'none';
		}
	}

	function drag(e) {
		if (!isDragging) return;
		
		currentX = e.clientX - dragStartX;
		currentY = e.clientY - dragStartY;
		
		// Keep modal within viewport bounds
		const modalRect = modalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		// Prevent modal from going off-screen (with some margin)
		const margin = 50;
		currentX = Math.max(-modalRect.width + margin, Math.min(currentX, viewportWidth - margin));
		currentY = Math.max(0, Math.min(currentY, viewportHeight - modalRect.height));
		
		modalElement.style.left = currentX + 'px';
		modalElement.style.top = currentY + 'px';
	}

	function stopDrag() {
		isDragging = false;
		modalElement.style.userSelect = '';
		document.removeEventListener('mousemove', drag);
		document.removeEventListener('mouseup', stopDrag);
	}

	function resetModalPosition() {
		if (modalElement) {
			modalElement.style.left = '';
			modalElement.style.top = '';
			modalElement.style.transform = '';
			currentX = 0;
			currentY = 0;
		}
	}

	// Drag functionality for manage modal
	function startManageDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingManage = true;
			
			const rect = manageModalElement.getBoundingClientRect();
			manageDragStartX = e.clientX - rect.left;
			manageDragStartY = e.clientY - rect.top;
			
			// Store initial position for dragging
			manageCurrentX = rect.left;
			manageCurrentY = rect.top;
			
			// Apply initial position
			manageModalElement.style.left = manageCurrentX + 'px';
			manageModalElement.style.top = manageCurrentY + 'px';
			manageModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragManage);
			document.addEventListener('mouseup', stopManageDrag);
			
			// Prevent text selection while dragging
			e.preventDefault();
			manageModalElement.style.userSelect = 'none';
		}
	}

	function dragManage(e) {
		if (!isDraggingManage) return;
		
		manageCurrentX = e.clientX - manageDragStartX;
		manageCurrentY = e.clientY - manageDragStartY;
		
		// Keep modal within viewport bounds
		const modalRect = manageModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		// Prevent modal from going off-screen (with some margin)
		const margin = 50;
		manageCurrentX = Math.max(-modalRect.width + margin, Math.min(manageCurrentX, viewportWidth - margin));
		manageCurrentY = Math.max(0, Math.min(manageCurrentY, viewportHeight - modalRect.height));
		
		manageModalElement.style.left = manageCurrentX + 'px';
		manageModalElement.style.top = manageCurrentY + 'px';
	}

	function stopManageDrag() {
		isDraggingManage = false;
		manageModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragManage);
		document.removeEventListener('mouseup', stopManageDrag);
	}

	function resetManageModalPosition() {
		if (manageModalElement) {
			manageModalElement.style.left = '';
			manageModalElement.style.top = '';
			manageModalElement.style.transform = '';
			manageCurrentX = 0;
			manageCurrentY = 0;
		}
	}

	async function testConnection() {
		testingConnection = true;
		error = '';
		success = '';

		try {
			await apiClient.testConnection(newConnection);
			success = 'Connection successful!';
		} catch (err) {
			error = err.response?.data?.error || 'Connection failed';
		} finally {
			testingConnection = false;
		}
	}

	async function saveConnection() {
		if (!success) {
			error = 'Please test the connection first';
			return;
		}

		loading = true;
		try {
			await apiClient.createConnection(newConnection);
			await loadConnections();
			closeModal();
			success = 'Connection saved successfully!';
		} catch (err) {
			error = err.response?.data?.error || 'Failed to save connection';
		} finally {
			loading = false;
		}
	}

	async function deleteConnection(id) {
		if (!confirm('Are you sure you want to delete this connection?')) {
			return;
		}

		try {
			await apiClient.deleteConnection(id);
			await loadConnections();
			success = 'Connection deleted successfully!';
		} catch (err) {
			error = 'Failed to delete connection';
		}
	}

	async function manageConnection(connection) {
		selectedConnection = connection;
		try {
			databaseInfo = await apiClient.getDatabaseInfo(connection.id);
		} catch (err) {
			error = 'Failed to load database information';
		}
	}

	function handleLogout() {
		logout();
		goto('/login');
	}
</script>

<svelte:head>
	<title>Database Connections - Database Manager</title>
</svelte:head>

<Navbar {user} {handleLogout} />

<div class="container">
	<div class="page-header">
		<h1>Database Connections</h1>
		<button class="btn" on:click={openModal}>Add Connection</button>
	</div>

	{#if error}
		<div class="alert alert-error">
			{error}
		</div>
	{/if}

	{#if success}
		<div class="alert alert-success">
			{success}
		</div>
	{/if}

	{#if loading}
		<div class="loading">
			<div class="spinner"></div>
		</div>
	{:else if $connections.length === 0}
		<div class="empty-state">
			<h3>No Database Connections</h3>
			<p>Get started by adding your first database connection</p>
			<button class="btn" on:click={openModal}>Add Your First Connection</button>
		</div>
	{:else}
		<div class="grid grid-3">
			{#each $connections as connection}
				<div class="connection-card">
					<div class="connection-header">
						<div class="connection-type">{connection.type.toUpperCase()}</div>
						<div class="connection-status status-{connection.status}">
							{connection.status}
						</div>
					</div>
					
					<div class="connection-body">
						<h3>{connection.name}</h3>
						<div class="connection-details">
							<div class="detail-item">
								<span class="detail-label">Host:</span>
								<span class="detail-value">{connection.host}:{connection.port}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Database:</span>
								<span class="detail-value">{connection.database}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Username:</span>
								<span class="detail-value">{connection.username || 'N/A'}</span>
							</div>
						</div>
					</div>

					<div class="connection-actions">
						<button class="btn btn-primary" on:click={() => manageConnection(connection)}>
							Manage
						</button>
						<button class="btn btn-danger" on:click={() => deleteConnection(connection.id)}>
							Delete
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Add Connection Modal -->
{#if showModal}
	<div 
		class="modal-content" 
		on:click|stopPropagation
		bind:this={modalElement}
		class:dragging={isDragging}
	>
		<div 
			class="modal-header"
			on:mousedown={startDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Add Database Connection
			</h2>
			<button class="modal-close" on:click={closeModal}>&times;</button>
		</div>

			{#if error}
				<div class="alert alert-error">
					{error}
				</div>
			{/if}

			{#if success}
				<div class="alert alert-success">
					{success}
				</div>
			{/if}

			<form on:submit|preventDefault>
				<div class="form-group">
					<label for="name" class="form-label">Connection Name</label>
					<input
						type="text"
						id="name"
						bind:value={newConnection.name}
						class="form-input"
						required
					/>
				</div>

				<div class="form-group">
					<label for="type" class="form-label">Database Type</label>
					<select
						id="type"
						bind:value={newConnection.type}
						class="form-select"
						on:change={updatePort}
						required
					>
						<option value="mysql">MySQL</option>
						<option value="postgres">PostgreSQL</option>
						<option value="mongodb">MongoDB</option>
						<option value="mariadb">MariaDB</option>
						<option value="sqlite">SQLite</option>
						<option value="redis">Redis</option>
						<option value="oracle">Oracle Database</option>
						<option value="sqlserver">SQL Server</option>
						<option value="cassandra">Apache Cassandra</option>
						<option value="elasticsearch">Elasticsearch</option>
						<option value="influxdb">InfluxDB</option>
						<option value="cockroachdb">CockroachDB</option>
					</select>
				</div>

				<div class="form-row">
					<div class="form-group">
						<label for="host" class="form-label">Host</label>
						<input
							type="text"
							id="host"
							bind:value={newConnection.host}
							class="form-input"
							placeholder="localhost"
							required
						/>
					</div>

					<div class="form-group">
						<label for="port" class="form-label">Port</label>
						<input
							type="number"
							id="port"
							bind:value={newConnection.port}
							class="form-input"
							required
						/>
					</div>
				</div>

				{#if databaseConfigs[newConnection.type]?.requiresDatabase}
					<div class="form-group">
						<label for="database" class="form-label">Database Name</label>
						<input
							type="text"
							id="database"
							bind:value={newConnection.database}
							class="form-input"
							required
						/>
					</div>
				{/if}

				<div class="form-row">
					<div class="form-group">
						<label for="username" class="form-label">Username</label>
						<input
							type="text"
							id="username"
							bind:value={newConnection.username}
							class="form-input"
						/>
					</div>

					<div class="form-group">
						<label for="password" class="form-label">Password</label>
						<input
							type="password"
							id="password"
							bind:value={newConnection.password}
							class="form-input"
						/>
					</div>
				</div>

				<div class="form-actions">
					<button
						type="button"
						class="btn btn-secondary"
						on:click={testConnection}
						disabled={testingConnection}
					>
						{#if testingConnection}
							Testing...
						{:else}
							Test Connection
						{/if}
					</button>

					<button
						type="button"
						class="btn"
						on:click={saveConnection}
						disabled={loading || !success}
					>
						{#if loading}
							Saving...
						{:else}
							Save Connection
						{/if}
					</button>
				</div>
			</form>
	</div>
{/if}

<!-- Database Info Modal -->
{#if selectedConnection && databaseInfo}
	<div 
		class="modal-content large" 
		on:click|stopPropagation
		bind:this={manageModalElement}
		class:dragging={isDraggingManage}
	>
		<div 
			class="modal-header"
			on:mousedown={startManageDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Manage {selectedConnection.name}
			</h2>
			<button class="modal-close" on:click={() => { selectedConnection = null; databaseInfo = null; resetManageModalPosition(); }}>&times;</button>
		</div>

			<div class="database-info">
				<h3>Database: {databaseInfo.name}</h3>
				
				{#if databaseInfo.collections}
					<h4>Collections ({databaseInfo.collections.length})</h4>
					<div class="collection-list">
						{#each databaseInfo.collections as collection}
							<div class="collection-item">
								<span class="collection-name">{collection}</span>
								<div class="collection-actions">
									<a href="/api-management?database={selectedConnection.id}&collection={collection}" class="btn btn-sm">
										Generate API
									</a>
								</div>
							</div>
						{/each}
					</div>
				{/if}

				{#if databaseInfo.tables}
					<h4>Tables ({databaseInfo.tables.length})</h4>
					<div class="collection-list">
						{#each databaseInfo.tables as table}
							<div class="collection-item">
								<span class="collection-name">{table}</span>
								<div class="collection-actions">
									<a href="/api-management?database={selectedConnection.id}&collection={table}" class="btn btn-sm">
										Generate API
									</a>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
	</div>
{/if}

<style>
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 32px;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: #333;
	}

	.connection-card {
		background: white;
		border-radius: 12px;
		padding: 24px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.connection-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
	}

	.connection-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.connection-type {
		background: #667eea;
		color: white;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 600;
	}

	.connection-body h3 {
		color: #333;
		margin-bottom: 16px;
		font-size: 1.2rem;
	}

	.connection-details {
		margin-bottom: 20px;
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		margin-bottom: 8px;
	}

	.detail-label {
		color: #666;
		font-weight: 500;
	}

	.detail-value {
		color: #333;
		font-family: monospace;
	}

	.connection-actions {
		display: flex;
		gap: 8px;
	}

	.connection-actions .btn {
		flex: 1;
		text-align: center;
		padding: 8px 16px;
		font-size: 0.9rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 16px;
	}

	.form-actions {
		display: flex;
		gap: 12px;
		margin-top: 24px;
	}

	.form-actions .btn {
		flex: 1;
	}

	@keyframes modalSlideIn {
		from {
			opacity: 0;
			transform: translate(-50%, -50%) scale(0.95);
		}
		to {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
	}

	/* Modern Modal Styling - No Background Overlay */
	.modal-content {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: white;
		border-radius: 16px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
		width: 500px;
		max-width: 90vw;
		max-height: 90vh;
		overflow-y: auto;
		animation: modalSlideIn 0.3s ease-out;
		cursor: default;
		transition: box-shadow 0.2s ease;
		z-index: 1000;
		border: 2px solid rgba(102, 126, 234, 0.2);
	}

	.modal-content.dragging {
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
		transform: none !important;
		animation: none;
		z-index: 1001;
		border-color: rgba(102, 126, 234, 0.5);
	}

	.modal-content.large {
		width: 700px;
		max-width: 95vw;
	}

	@keyframes modalSlideIn {
		from {
			opacity: 0;
			transform: translateY(-30px) scale(0.95);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	.modal-content.large {
		max-width: 700px;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24px 24px 16px 24px;
		border-bottom: 1px solid #eee;
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border-radius: 16px 16px 0 0;
		cursor: move;
		user-select: none;
		position: relative;
	}

	.modal-header:hover {
		background: linear-gradient(135deg, #5a6fd8, #6a42a0);
	}

	.modal-header:active {
		cursor: grabbing;
	}

	.modal-title {
		font-size: 1.5rem;
		font-weight: 600;
		color: white;
		margin: 0;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.drag-icon {
		font-size: 1.2rem;
		opacity: 0.7;
		cursor: move;
		transform: rotate(90deg);
		transition: opacity 0.2s;
	}

	.modal-header:hover .drag-icon {
		opacity: 1;
	}

	.modal-close {
		background: rgba(255, 255, 255, 0.2);
		border: none;
		color: white;
		font-size: 24px;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.2s;
	}

	.modal-close:hover {
		background: rgba(255, 255, 255, 0.3);
	}

	.modal form {
		padding: 24px;
	}

	.modal-content.large {
		max-width: 600px;
	}

	.database-info h3 {
		color: #333;
		margin-bottom: 20px;
		padding-bottom: 10px;
		border-bottom: 1px solid #eee;
	}

	.database-info h4 {
		color: #555;
		margin: 20px 0 12px 0;
		font-size: 1.1rem;
	}

	.collection-list {
		max-height: 300px;
		overflow-y: auto;
		border: 1px solid #eee;
		border-radius: 6px;
	}

	.collection-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #eee;
	}

	.collection-item:last-child {
		border-bottom: none;
	}

	.collection-item:hover {
		background-color: #f8f9fa;
	}

	.collection-name {
		font-family: monospace;
		font-weight: 500;
		color: #333;
	}

	.btn-sm {
		padding: 6px 12px;
		font-size: 0.8rem;
	}

	.empty-state {
		text-align: center;
		padding: 80px 20px;
		color: #666;
	}

	.empty-state h3 {
		color: #333;
		margin-bottom: 12px;
	}

	.empty-state p {
		margin-bottom: 24px;
		font-size: 1.1rem;
	}
</style>
