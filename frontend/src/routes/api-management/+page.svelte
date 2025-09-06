<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { apiClient } from '$lib/api';
	import { isAuthenticated, connections, apiKeys, endpoints, logs, logout, user } from '$lib/stores';
	import Navbar from '$lib/components/Navbar.svelte';

	let activeTab = 'keys';
	let loading = false;
	let error = '';
	let success = '';
	
	// API Keys
	let showKeyModal = false;
	let newAPIKey = { database_id: '', name: '' };
	
	// Endpoints
	let showEndpointModal = false;
	let newEndpoint = { database_id: '', collection: '', method: 'GET' };

	// Code Example Modal
	let showCodeModal = false;
	let selectedEndpoint = null;
	let codeLanguage = 'javascript';

	// Drag functionality for API Key modal
	let isDraggingKey = false;
	let keyModalElement;
	let keyDragStartX = 0;
	let keyDragStartY = 0;
	let keyCurrentX = 0;
	let keyCurrentY = 0;

	// Drag functionality for Endpoint modal
	let isDraggingEndpoint = false;
	let endpointModalElement;
	let endpointDragStartX = 0;
	let endpointDragStartY = 0;
	let endpointCurrentX = 0;
	let endpointCurrentY = 0;

	// Drag functionality for Code Example modal
	let isDraggingCode = false;
	let codeModalElement;
	let codeDragStartX = 0;
	let codeDragStartY = 0;
	let codeCurrentX = 0;
	let codeCurrentY = 0;
	
	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		// Get tab from URL params
		const urlParams = new URLSearchParams(window.location.search);
		activeTab = urlParams.get('tab') || 'endpoints';
		
		// Check if we need to auto-generate endpoints for a collection
		const databaseId = urlParams.get('database');
		const collectionName = urlParams.get('collection');

		await loadData();

		// Auto-generate endpoints if parameters are provided
		if (databaseId && collectionName) {
			await generateCollectionEndpoints(databaseId, collectionName);
		}
	});

	async function loadData() {
		loading = true;
		try {
			const [connectionsData, apiKeysData, endpointsData, logsData] = await Promise.all([
				apiClient.getConnections(),
				apiClient.getAPIKeys(),
				apiClient.getEndpoints(),
				apiClient.getLogs()
			]);

			connections.set(connectionsData);
			apiKeys.set(apiKeysData);
			endpoints.set(endpointsData);
			logs.set(logsData);
		} catch (err) {
			error = 'Failed to load data';
		} finally {
			loading = false;
		}
	}

	function switchTab(tab) {
		activeTab = tab;
		const url = new URL(window.location);
		url.searchParams.set('tab', tab);
		window.history.pushState({}, '', url);
	}

	// API Keys Management
	function openAPIKeyModal() {
		showKeyModal = true;
		newAPIKey = { database_id: '', name: '' };
		error = '';
	}

	async function createAPIKey() {
		if (!newAPIKey.database_id || !newAPIKey.name) {
			error = 'Please fill all fields';
			return;
		}

		loading = true;
		try {
			await apiClient.createAPIKey(newAPIKey);
			await loadData();
			showKeyModal = false;
			success = 'API key created successfully';
		} catch (err) {
			error = err.response?.data?.error || 'Failed to create API key';
		} finally {
			loading = false;
		}
	}

	async function toggleAPIKey(id) {
		try {
			await apiClient.toggleAPIKey(id);
			await loadData();
			success = 'API key status updated';
		} catch (err) {
			error = 'Failed to update API key';
		}
	}

	// Endpoints Management
	function openEndpointModal() {
		showEndpointModal = true;
		newEndpoint = { database_id: '', collection: '', method: 'GET' };
		error = '';
	}

	async function createEndpoint() {
		if (!newEndpoint.database_id || !newEndpoint.collection || !newEndpoint.method) {
			error = 'Please fill all fields';
			return;
		}

		loading = true;
		try {
			await apiClient.createEndpoint(newEndpoint);
			await loadData();
			showEndpointModal = false;
			success = 'Endpoint created successfully';
		} catch (err) {
			error = err.response?.data?.error || 'Failed to create endpoint';
		} finally {
			loading = false;
		}
	}

	async function toggleEndpoint(id) {
		try {
			await apiClient.toggleEndpoint(id);
			await loadData();
			success = 'Endpoint status updated';
		} catch (err) {
			error = 'Failed to update endpoint';
		}
	}

	async function deleteEndpoint(id) {
		if (!confirm('Are you sure you want to delete this endpoint? This action cannot be undone.')) {
			return;
		}

		try {
			await apiClient.deleteEndpoint(id);
			await loadData();
			success = 'Endpoint deleted successfully';
		} catch (err) {
			error = `Failed to delete endpoint: ${err.response?.data?.error || err.message}`;
		}
	}

	async function generateCollectionEndpoints(databaseId, collectionName) {
		const methods = ['GET', 'POST', 'PUT', 'DELETE'];
		let generatedCount = 0;
		
		try {
			for (const method of methods) {
				// Check if endpoint already exists
				const existingEndpoint = $endpoints.find(ep => 
					ep.database_id === databaseId && 
					ep.collection === collectionName && 
					ep.method === method
				);

				if (!existingEndpoint) {
					await apiClient.createEndpoint({
						database_id: databaseId,
						collection: collectionName,
						method: method
					});
					generatedCount++;
				}
			}

			if (generatedCount > 0) {
				await loadData();
				success = `Generated ${generatedCount} new endpoints for collection "${collectionName}"`;
			} else {
				success = `All endpoints for collection "${collectionName}" already exist`;
			}
		} catch (err) {
			error = `Failed to generate endpoints: ${err.response?.data?.error || err.message}`;
		}
	}

	function copyToClipboard(text) {
		navigator.clipboard.writeText(text);
		success = 'Copied to clipboard';
	}

	// Code Example Functions
	function showCodeExample(endpoint) {
		selectedEndpoint = endpoint;
		showCodeModal = true;
	}

	function closeCodeModal() {
		showCodeModal = false;
		selectedEndpoint = null;
		resetCodeModalPosition();
	}

	function generateCodeExample(endpoint, language = 'javascript') {
		if (!endpoint) return '';

		const baseUrl = 'http://localhost:8080';
		const url = `${baseUrl}${endpoint.path}`;
		const collection = endpoint.collection;
		
		// Get API Key for headers
		const apiKey = $apiKeys.length > 0 ? $apiKeys[0].key : 'your-api-key-here';
		
		// Sample data based on collection name
		const getSampleData = (collection) => {
			if (collection === 'mahasiswas' || collection === 'mahasiswa') {
				return {
					nama: "John Doe",
					nim: "123456789",
					jurusan: "Informatika",
					email: "john@example.com",
					angkatan: "2023",
					alamat: "Jakarta"
				};
			} else if (collection === 'alumnis' || collection === 'alumni') {
				return {
					nama: "Jane Smith",
					nim: "987654321",
					jurusan: "Sistem Informasi",
					email: "jane@example.com",
					tahun_lulus: "2022",
					pekerjaan: "Software Engineer"
				};
			} else if (collection === 'news' || collection === 'berita') {
				return {
					title: "Breaking News",
					content: "This is the news content...",
					author: "Admin",
					published: true,
					created_at: new Date().toISOString()
				};
			} else if (collection === 'users' || collection === 'user') {
				return {
					username: "johndoe",
					email: "john@example.com",
					full_name: "John Doe",
					role: "user",
					active: true
				};
			} else if (collection === 'products' || collection === 'produk') {
				return {
					name: "Sample Product",
					description: "Product description",
					price: 100000,
					category: "Electronics",
					stock: 50
				};
			} else {
				return {
					name: "Sample Name",
					description: "Sample description",
					status: "active",
					created_at: new Date().toISOString()
				};
			}
		};

		const sampleData = getSampleData(collection);

		if (language === 'javascript') {
			switch (endpoint.method) {
				case 'GET':
					return `// Get all ${collection}
fetch('${url}', {
  headers: {
    'X-API-Key': '${apiKey}',
    'Content-Type': 'application/json'
  }
})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));

// Get single ${collection.slice(0, -1)} by ID
fetch('${url}/1', {
  headers: {
    'X-API-Key': '${apiKey}',
    'Content-Type': 'application/json'
  }
})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));`;

				case 'POST':
					return `// Create new ${collection.slice(0, -1)}
fetch('${url}', {
  method: 'POST',
  headers: {
    'X-API-Key': '${apiKey}',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(${JSON.stringify(sampleData, null, 2)})
})
  .then(response => response.json())
  .then(data => console.log('Created:', data))
  .catch(error => console.error('Error:', error));`;

				case 'PUT':
					return `// Update ${collection.slice(0, -1)} by ID
fetch('${url}/1', {
  method: 'PUT',
  headers: {
    'X-API-Key': '${apiKey}',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(${JSON.stringify(sampleData, null, 2)})
})
  .then(response => response.json())
  .then(data => console.log('Updated:', data))
  .catch(error => console.error('Error:', error));`;

				case 'DELETE':
					return `// Delete ${collection.slice(0, -1)} by ID
fetch('${url}/1', {
  method: 'DELETE',
  headers: {
    'X-API-Key': '${apiKey}',
    'Content-Type': 'application/json'
  }
})
  .then(response => {
    if (response.ok) {
      console.log('${collection.slice(0, -1)} deleted successfully');
    }
  })
  .catch(error => console.error('Error:', error));`;

				default:
					return `// ${endpoint.method} request to ${url}
fetch('${url}', {
  method: '${endpoint.method}',
  headers: {
    'X-API-Key': '${apiKey}',
    'Content-Type': 'application/json'
  }
})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));`;
			}
		} else if (language === 'curl') {
			switch (endpoint.method) {
				case 'GET':
					return `# Get all ${collection}
curl -X GET "${url}" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json"

# Get single ${collection.slice(0, -1)} by ID
curl -X GET "${url}/1" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json"`;

				case 'POST':
					return `# Create new ${collection.slice(0, -1)}
curl -X POST "${url}" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify(sampleData, null, 2)}'`;

				case 'PUT':
					return `# Update ${collection.slice(0, -1)} by ID
curl -X PUT "${url}/1" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify(sampleData, null, 2)}'`;

				case 'DELETE':
					return `# Delete ${collection.slice(0, -1)} by ID
curl -X DELETE "${url}/1" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json"`;

				default:
					return `# ${endpoint.method} request to ${url}
curl -X ${endpoint.method} "${url}" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json"`;
			}
		}

		return '';
	}

	function formatDateTime(dateString) {
		return new Date(dateString).toLocaleString();
	}

	function getStatusCodeClass(code) {
		if (code >= 200 && code < 300) return 'status-success';
		if (code >= 400 && code < 500) return 'status-warning';
		if (code >= 500) return 'status-error';
		return '';
	}

	async function deleteAPIKeyHandler(apiKeyId) {
		if (!confirm('Are you sure you want to delete this API key? This action cannot be undone.')) {
			return;
		}

		try {
			await apiClient.deleteAPIKey(apiKeyId);
			await loadData();
			success = 'API key deleted successfully';
		} catch (err) {
			error = `Failed to delete API key: ${err.response?.data?.error || err.message}`;
		}
	}

	async function deleteEndpointHandler(endpointId) {
		if (!confirm('Are you sure you want to delete this endpoint? This action cannot be undone.')) {
			return;
		}

		try {
			await apiClient.deleteEndpoint(endpointId);
			await loadData();
			success = 'Endpoint deleted successfully';
		} catch (err) {
			error = `Failed to delete endpoint: ${err.response?.data?.error || err.message}`;
		}
	}

	async function clearLogsHandler() {
		if (!confirm('Are you sure you want to clear all logs? This action cannot be undone.')) {
			return;
		}

		try {
			await apiClient.clearLogs();
			await loadData();
			success = 'All logs cleared successfully';
		} catch (err) {
			error = `Failed to clear logs: ${err.response?.data?.error || err.message}`;
		}
	}

	function handleLogout() {
		logout();
		goto('/login');
	}

	// Drag functionality for API Key modal
	function startKeyDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingKey = true;
			
			const rect = keyModalElement.getBoundingClientRect();
			keyDragStartX = e.clientX - rect.left;
			keyDragStartY = e.clientY - rect.top;
			
			keyCurrentX = rect.left;
			keyCurrentY = rect.top;
			
			keyModalElement.style.left = keyCurrentX + 'px';
			keyModalElement.style.top = keyCurrentY + 'px';
			keyModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragKey);
			document.addEventListener('mouseup', stopKeyDrag);
			
			e.preventDefault();
			keyModalElement.style.userSelect = 'none';
		}
	}

	function dragKey(e) {
		if (!isDraggingKey) return;
		
		keyCurrentX = e.clientX - keyDragStartX;
		keyCurrentY = e.clientY - keyDragStartY;
		
		const modalRect = keyModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		const margin = 50;
		keyCurrentX = Math.max(-modalRect.width + margin, Math.min(keyCurrentX, viewportWidth - margin));
		keyCurrentY = Math.max(0, Math.min(keyCurrentY, viewportHeight - modalRect.height));
		
		keyModalElement.style.left = keyCurrentX + 'px';
		keyModalElement.style.top = keyCurrentY + 'px';
	}

	function stopKeyDrag() {
		isDraggingKey = false;
		keyModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragKey);
		document.removeEventListener('mouseup', stopKeyDrag);
	}

	function resetKeyModalPosition() {
		if (keyModalElement) {
			keyModalElement.style.left = '';
			keyModalElement.style.top = '';
			keyModalElement.style.transform = '';
			keyCurrentX = 0;
			keyCurrentY = 0;
		}
	}

	// Drag functionality for Endpoint modal
	function startEndpointDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingEndpoint = true;
			
			const rect = endpointModalElement.getBoundingClientRect();
			endpointDragStartX = e.clientX - rect.left;
			endpointDragStartY = e.clientY - rect.top;
			
			endpointCurrentX = rect.left;
			endpointCurrentY = rect.top;
			
			endpointModalElement.style.left = endpointCurrentX + 'px';
			endpointModalElement.style.top = endpointCurrentY + 'px';
			endpointModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragEndpoint);
			document.addEventListener('mouseup', stopEndpointDrag);
			
			e.preventDefault();
			endpointModalElement.style.userSelect = 'none';
		}
	}

	function dragEndpoint(e) {
		if (!isDraggingEndpoint) return;
		
		endpointCurrentX = e.clientX - endpointDragStartX;
		endpointCurrentY = e.clientY - endpointDragStartY;
		
		const modalRect = endpointModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		const margin = 50;
		endpointCurrentX = Math.max(-modalRect.width + margin, Math.min(endpointCurrentX, viewportWidth - margin));
		endpointCurrentY = Math.max(0, Math.min(endpointCurrentY, viewportHeight - modalRect.height));
		
		endpointModalElement.style.left = endpointCurrentX + 'px';
		endpointModalElement.style.top = endpointCurrentY + 'px';
	}

	function stopEndpointDrag() {
		isDraggingEndpoint = false;
		endpointModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragEndpoint);
		document.removeEventListener('mouseup', stopEndpointDrag);
	}

	function resetEndpointModalPosition() {
		if (endpointModalElement) {
			endpointModalElement.style.left = '';
			endpointModalElement.style.top = '';
			endpointModalElement.style.transform = '';
			endpointCurrentX = 0;
			endpointCurrentY = 0;
		}
	}

	// Drag functionality for Code modal
	function startCodeDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingCode = true;
			
			const rect = codeModalElement.getBoundingClientRect();
			codeDragStartX = e.clientX - rect.left;
			codeDragStartY = e.clientY - rect.top;
			
			codeCurrentX = rect.left;
			codeCurrentY = rect.top;
			
			codeModalElement.style.left = codeCurrentX + 'px';
			codeModalElement.style.top = codeCurrentY + 'px';
			codeModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragCode);
			document.addEventListener('mouseup', stopCodeDrag);
			
			e.preventDefault();
			codeModalElement.style.userSelect = 'none';
		}
	}

	function dragCode(e) {
		if (!isDraggingCode) return;
		
		codeCurrentX = e.clientX - codeDragStartX;
		codeCurrentY = e.clientY - codeDragStartY;
		
		const modalRect = codeModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		const margin = 50;
		codeCurrentX = Math.max(-modalRect.width + margin, Math.min(codeCurrentX, viewportWidth - margin));
		codeCurrentY = Math.max(0, Math.min(codeCurrentY, viewportHeight - modalRect.height));
		
		codeModalElement.style.left = codeCurrentX + 'px';
		codeModalElement.style.top = codeCurrentY + 'px';
	}

	function stopCodeDrag() {
		isDraggingCode = false;
		codeModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragCode);
		document.removeEventListener('mouseup', stopCodeDrag);
	}

	function resetCodeModalPosition() {
		if (codeModalElement) {
			codeModalElement.style.left = '';
			codeModalElement.style.top = '';
			codeModalElement.style.transform = '';
			codeCurrentX = 0;
			codeCurrentY = 0;
		}
	}

	// Update close functions
	function closeKeyModal() {
		showKeyModal = false;
		resetKeyModalPosition();
	}

	function closeEndpointModal() {
		showEndpointModal = false;
		resetEndpointModalPosition();
	}
</script>

<svelte:head>
	<title>API Management - Database Manager</title>
</svelte:head>

<Navbar {user} {handleLogout} />

<div class="container">
	<div class="page-header">
		<h1>API Management</h1>
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

	<!-- Tabs -->
	<div class="tabs">
		<button 
			class="tab {activeTab === 'keys' ? 'active' : ''}"
			on:click={() => switchTab('keys')}
		>
			API Keys
		</button>
		<button 
			class="tab {activeTab === 'endpoints' ? 'active' : ''}"
			on:click={() => switchTab('endpoints')}
		>
			Endpoints
		</button>
		<button 
			class="tab {activeTab === 'logs' ? 'active' : ''}"
			on:click={() => switchTab('logs')}
		>
			API Logs
		</button>
	</div>

	{#if loading}
		<div class="loading">
			<div class="spinner"></div>
		</div>
	{:else}
		<!-- API Keys Tab -->
		{#if activeTab === 'keys'}
			<div class="tab-content">
				<div class="tab-header">
					<h2>API Keys</h2>
					<button class="btn" on:click={openAPIKeyModal}>Create API Key</button>
				</div>

				{#if $apiKeys.length === 0}
					<div class="empty-state">
						<p>No API keys created yet</p>
						<button class="btn" on:click={openAPIKeyModal}>Create First API Key</button>
					</div>
				{:else}
					<div class="api-keys-grid">
						{#each $apiKeys as apiKey}
							<div class="api-key-card">
								<div class="api-key-header">
									<div class="api-key-name">{apiKey.name}</div>
									<div class="api-key-status">
										<button 
											class="toggle-btn {apiKey.is_active ? 'active' : 'inactive'}"
											on:click={() => toggleAPIKey(apiKey.id)}
										>
											{apiKey.is_active ? 'Active' : 'Inactive'}
										</button>
										<button 
											class="btn btn-sm btn-danger"
											on:click={() => deleteAPIKeyHandler(apiKey.id)}
											title="Delete API Key"
										>
											Delete
										</button>
									</div>
								</div>
								
								<div class="api-key-info">
									<div class="info-item">
										<span class="info-label">Database:</span>
										<span class="info-value">{apiKey.database?.name || 'Unknown'}</span>
									</div>
									<div class="info-item">
										<span class="info-label">Created:</span>
										<span class="info-value">{formatDateTime(apiKey.created_at)}</span>
									</div>
								</div>

								<div class="api-key-value">
									<input type="text" value={apiKey.key} readonly class="key-input" />
									<button class="btn-copy" on:click={() => copyToClipboard(apiKey.key)}>
										Copy
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- Endpoints Tab -->
		{#if activeTab === 'endpoints'}
			<div class="tab-content">
				<div class="tab-header">
					<h2>API Endpoints</h2>
					<button class="btn" on:click={openEndpointModal}>Create Endpoint</button>
				</div>

				{#if $endpoints.length === 0}
					<div class="empty-state">
						<p>No endpoints created yet</p>
						<button class="btn" on:click={openEndpointModal}>Create First Endpoint</button>
					</div>
				{:else}
					<div class="endpoints-table">
						<table class="table">
							<thead>
								<tr>
									<th>Method</th>
									<th>Path</th>
									<th>Collection</th>
									<th>Database</th>
									<th>Status</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each $endpoints as endpoint}
									<tr>
										<td>
											<span class="method-badge method-{endpoint.method.toLowerCase()}">
												{endpoint.method}
											</span>
										</td>
										<td class="path-cell">
											<code>{endpoint.path}</code>
											<button class="btn-copy" on:click={() => copyToClipboard(`http://localhost:8080${endpoint.path}`)}>
												Copy URL
											</button>
										</td>
										<td>{endpoint.collection}</td>
										<td>{endpoint.database?.name || 'Unknown'}</td>
										<td>
											<span class="status-badge {endpoint.is_active ? 'status-active' : 'status-inactive'}">
												{endpoint.is_active ? 'Active' : 'Inactive'}
											</span>
										</td>
										<td>
											<div class="action-buttons">
												<button 
													class="btn btn-sm btn-info"
													on:click={() => showCodeExample(endpoint)}
												>
													View Code
												</button>
												<button 
													class="btn btn-sm"
													on:click={() => toggleEndpoint(endpoint.id)}
												>
													{endpoint.is_active ? 'Disable' : 'Enable'}
												</button>
												<button 
													class="btn btn-sm btn-danger"
													on:click={() => deleteEndpointHandler(endpoint.id)}
												>
													Delete
												</button>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Logs Tab -->
		{#if activeTab === 'logs'}
			<div class="tab-content">
				<div class="tab-header">
					<h2>API Request Logs</h2>
					{#if $logs.length > 0}
						<button class="btn btn-danger" on:click={clearLogsHandler}>
							Clear All Logs
						</button>
					{/if}
				</div>

				{#if $logs.length === 0}
					<div class="empty-state">
						<p>No API requests logged yet</p>
					</div>
				{:else}
					<div class="logs-table">
						<table class="table">
							<thead>
								<tr>
									<th>Timestamp</th>
									<th>Method</th>
									<th>Path</th>
									<th>Status</th>
									<th>Response Time</th>
									<th>IP Address</th>
								</tr>
							</thead>
							<tbody>
								{#each $logs as log}
									<tr>
										<td>{formatDateTime(log.created_at)}</td>
										<td>
											<span class="method-badge method-{log.method.toLowerCase()}">
												{log.method}
											</span>
										</td>
										<td><code>{log.path}</code></td>
										<td>
											<span class="status-code {getStatusCodeClass(log.status_code)}">
												{log.status_code}
											</span>
										</td>
										<td>{log.response_time}ms</td>
										<td>{log.ip_address}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<!-- Create API Key Modal -->
{#if showKeyModal}
	<div 
		class="modal-content" 
		on:click|stopPropagation
		bind:this={keyModalElement}
		class:dragging={isDraggingKey}
	>
		<div 
			class="modal-header"
			on:mousedown={startKeyDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Create API Key
			</h2>
			<button class="modal-close" on:click={closeKeyModal}>&times;</button>
		</div>

			<form on:submit|preventDefault={createAPIKey}>
				<div class="form-group">
					<label for="keyName" class="form-label">API Key Name</label>
					<input
						type="text"
						id="keyName"
						bind:value={newAPIKey.name}
						class="form-input"
						placeholder="e.g., Production API Key"
						required
					/>
				</div>

				<div class="form-group">
					<label for="keyDatabase" class="form-label">Database Connection</label>
					<select
						id="keyDatabase"
						bind:value={newAPIKey.database_id}
						class="form-select"
						required
					>
						<option value="">Select a database</option>
						{#each $connections as connection}
							<option value={connection.id}>{connection.name} ({connection.type})</option>
						{/each}
					</select>
				</div>

				<div class="form-actions">
					<button type="button" class="btn btn-secondary" on:click={closeKeyModal}>
						Cancel
					</button>
					<button type="submit" class="btn" disabled={loading}>
						{loading ? 'Creating...' : 'Create API Key'}
					</button>
				</div>
			</form>
	</div>
{/if}

<!-- Create Endpoint Modal -->
{#if showEndpointModal}
	<div 
		class="modal-content" 
		on:click|stopPropagation
		bind:this={endpointModalElement}
		class:dragging={isDraggingEndpoint}
	>
		<div 
			class="modal-header"
			on:mousedown={startEndpointDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Create API Endpoint
			</h2>
			<button class="modal-close" on:click={closeEndpointModal}>&times;</button>
		</div>

			<form on:submit|preventDefault={createEndpoint}>
				<div class="form-group">
					<label for="endpointDatabase" class="form-label">Database Connection</label>
					<select
						id="endpointDatabase"
						bind:value={newEndpoint.database_id}
						class="form-select"
						required
					>
						<option value="">Select a database</option>
						{#each $connections as connection}
							<option value={connection.id}>{connection.name} ({connection.type})</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label for="endpointCollection" class="form-label">Collection/Table Name</label>
					<input
						type="text"
						id="endpointCollection"
						bind:value={newEndpoint.collection}
						class="form-input"
						placeholder="e.g., users, products"
						required
					/>
				</div>

				<div class="form-group">
					<label for="endpointMethod" class="form-label">HTTP Method</label>
					<select
						id="endpointMethod"
						bind:value={newEndpoint.method}
						class="form-select"
						required
					>
						<option value="GET">GET</option>
						<option value="POST">POST</option>
						<option value="PUT">PUT</option>
						<option value="DELETE">DELETE</option>
					</select>
				</div>

				<div class="form-actions">
					<button type="button" class="btn btn-secondary" on:click={closeEndpointModal}>
						Cancel
					</button>
					<button type="submit" class="btn" disabled={loading}>
						{loading ? 'Creating...' : 'Create Endpoint'}
					</button>
				</div>
			</form>
	</div>
{/if}

<!-- Code Example Modal -->
{#if showCodeModal && selectedEndpoint}
	<div 
		class="modal-content code-modal" 
		on:click|stopPropagation
		bind:this={codeModalElement}
		class:dragging={isDraggingCode}
	>
		<div 
			class="modal-header"
			on:mousedown={startCodeDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Code Examples - {selectedEndpoint.method} {selectedEndpoint.path}
			</h2>
			<button class="modal-close" on:click={closeCodeModal}>&times;</button>
		</div>

		<div class="modal-body">
			<div class="code-tabs">
				<button 
					class="tab-button"
					class:active={codeLanguage === 'javascript'}
					on:click={() => codeLanguage = 'javascript'}
				>
					JavaScript (Fetch)
				</button>
				<button 
					class="tab-button"
					class:active={codeLanguage === 'curl'}
					on:click={() => codeLanguage = 'curl'}
				>
					cURL
				</button>
			</div>

			<div class="code-container">
				<pre><code>{generateCodeExample(selectedEndpoint, codeLanguage)}</code></pre>
				<button 
					class="btn-copy-code"
					on:click={() => copyToClipboard(generateCodeExample(selectedEndpoint, codeLanguage))}
				>
					Copy Code
				</button>
			</div>

			<div class="code-info">
				<h4>API Information:</h4>
				<ul>
					<li><strong>Endpoint:</strong> {selectedEndpoint.path}</li>
					<li><strong>Method:</strong> {selectedEndpoint.method}</li>
					<li><strong>Collection:</strong> {selectedEndpoint.collection}</li>
					<li><strong>Database:</strong> {selectedEndpoint.database?.name || 'Unknown'}</li>
					<li><strong>Status:</strong> {selectedEndpoint.is_active ? 'Active' : 'Inactive'}</li>
				</ul>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Draggable Modal Styling */
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

	:global(.modal-content) {
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

	:global(.modal-content.dragging) {
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
		transform: none !important;
		animation: none;
		z-index: 1001;
		border-color: rgba(102, 126, 234, 0.5);
	}

	:global(.modal-header) {
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

	:global(.modal-header:hover) {
		background: linear-gradient(135deg, #5a6fd8, #6a42a0);
	}

	:global(.modal-header:active) {
		cursor: grabbing;
	}

	:global(.modal-title) {
		font-size: 1.5rem;
		font-weight: 600;
		color: white;
		margin: 0;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	:global(.drag-icon) {
		font-size: 1.2rem;
		opacity: 0.7;
		cursor: move;
		transform: rotate(90deg);
		transition: opacity 0.2s;
	}

	:global(.modal-header:hover .drag-icon) {
		opacity: 1;
	}

	:global(.modal-close) {
		background: rgba(255, 255, 255, 0.2);
		border: none;
		color: white;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		font-size: 1.5rem;
		transition: background 0.2s;
	}

	:global(.modal-close:hover) {
		background: rgba(255, 255, 255, 0.3);
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: #333;
		margin-bottom: 32px;
	}

	.tabs {
		display: flex;
		border-bottom: 2px solid #eee;
		margin-bottom: 32px;
	}

	.tab {
		background: none;
		border: none;
		padding: 16px 24px;
		font-size: 1rem;
		font-weight: 600;
		color: #666;
		cursor: pointer;
		border-bottom: 3px solid transparent;
		transition: all 0.2s;
	}

	.tab:hover {
		color: #333;
	}

	.tab.active {
		color: #667eea;
		border-bottom-color: #667eea;
	}

	.tab-content {
		min-height: 400px;
	}

	.tab-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 24px;
	}

	.tab-header h2 {
		color: #333;
		font-size: 1.5rem;
		margin: 0;
	}

	.api-keys-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 20px;
	}

	.api-key-card {
		background: white;
		border-radius: 12px;
		padding: 24px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		border: 1px solid #eee;
	}

	.api-key-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.api-key-name {
		font-size: 1.1rem;
		font-weight: 600;
		color: #333;
	}

	.toggle-btn {
		padding: 6px 12px;
		border: none;
		border-radius: 6px;
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.toggle-btn.active {
		background: #28a745;
		color: white;
	}

	.toggle-btn.inactive {
		background: #dc3545;
		color: white;
	}

	.api-key-info {
		margin-bottom: 16px;
	}

	.info-item {
		display: flex;
		justify-content: space-between;
		margin-bottom: 8px;
	}

	.info-label {
		color: #666;
		font-weight: 500;
	}

	.info-value {
		color: #333;
	}

	.api-key-value {
		display: flex;
		gap: 8px;
	}

	.key-input {
		flex: 1;
		padding: 8px 12px;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-family: monospace;
		font-size: 0.9rem;
		background: #f8f9fa;
	}

	.btn-copy {
		padding: 8px 16px;
		background: #667eea;
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 0.9rem;
		cursor: pointer;
		transition: background 0.2s;
	}

	.btn-copy:hover {
		background: #5a6fd8;
	}

	.btn-danger {
		background: #dc3545;
		color: white;
	}

	.btn-danger:hover {
		background: #c82333;
	}

	.action-buttons {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.api-key-status {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.endpoints-table, .logs-table {
		background: white;
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.method-badge {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.method-get { background: #28a745; color: white; }
	.method-post { background: #007bff; color: white; }
	.method-put { background: #ffc107; color: #333; }
	.method-delete { background: #dc3545; color: white; }

	.path-cell {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.path-cell code {
		background: #f8f9fa;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.9rem;
	}

	.status-badge {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 600;
	}

	.status-active {
		background: #d4edda;
		color: #155724;
	}

	.status-inactive {
		background: #f8d7da;
		color: #721c24;
	}

	.status-code {
		padding: 4px 8px;
		border-radius: 4px;
		font-weight: 600;
		font-size: 0.9rem;
	}

	.status-success {
		background: #d4edda;
		color: #155724;
	}

	.status-warning {
		background: #fff3cd;
		color: #856404;
	}

	.status-error {
		background: #f8d7da;
		color: #721c24;
	}

	.form-actions {
		display: flex;
		gap: 12px;
		margin-top: 24px;
	}

	.form-actions .btn {
		flex: 1;
	}

	.empty-state {
		text-align: center;
		padding: 80px 20px;
		color: #666;
	}

	.empty-state p {
		margin-bottom: 20px;
		font-size: 1.1rem;
	}

	.action-buttons {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.btn-danger {
		background: #dc3545;
		color: white;
		border: 1px solid #dc3545;
	}

	.btn-danger:hover {
		background: #c82333;
		border-color: #bd2130;
	}

	/* Code Modal Specific Styles */
	:global(.code-modal) {
		width: 800px;
		max-width: 95vw;
	}

	.code-tabs {
		display: flex;
		gap: 8px;
		margin-bottom: 16px;
		border-bottom: 1px solid #eee;
	}

	.tab-button {
		background: transparent;
		border: none;
		padding: 8px 16px;
		cursor: pointer;
		border-bottom: 2px solid transparent;
		font-weight: 500;
		color: #666;
		transition: all 0.2s;
	}

	.tab-button.active {
		color: #667eea;
		border-bottom-color: #667eea;
	}

	.tab-button:hover {
		color: #667eea;
		background: rgba(102, 126, 234, 0.1);
	}

	.code-container {
		position: relative;
		background: #1e1e1e;
		border-radius: 8px;
		margin-bottom: 24px;
		overflow: hidden;
	}

	.code-container pre {
		margin: 0;
		padding: 20px;
		background: #1e1e1e;
		color: #d4d4d4;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		font-size: 14px;
		line-height: 1.6;
		overflow-x: auto;
	}

	.code-container code {
		background: transparent;
		color: #d4d4d4;
		font-family: inherit;
		font-size: inherit;
	}

	.btn-copy-code {
		position: absolute;
		top: 12px;
		right: 12px;
		background: rgba(102, 126, 234, 0.8);
		color: white;
		border: none;
		padding: 6px 12px;
		border-radius: 4px;
		font-size: 12px;
		cursor: pointer;
		transition: background 0.2s;
	}

	.btn-copy-code:hover {
		background: rgba(102, 126, 234, 1);
	}

	.code-info {
		background: #f8f9fa;
		padding: 16px;
		border-radius: 8px;
		border: 1px solid #e9ecef;
	}

	.code-info h4 {
		margin: 0 0 12px 0;
		color: #333;
		font-size: 1.1rem;
	}

	.code-info ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.code-info li {
		padding: 4px 0;
		color: #555;
	}

	.code-info strong {
		color: #333;
		font-weight: 600;
	}

	.btn-info {
		background: #17a2b8;
		color: white;
		border: 1px solid #17a2b8;
	}

	.btn-info:hover {
		background: #138496;
		border-color: #117a8b;
	}

	/* Code Modal Styling */
	.code-modal {
		width: 800px;
		max-width: 95vw;
	}

	.code-tabs {
		display: flex;
		gap: 8px;
		margin-bottom: 16px;
		border-bottom: 1px solid #eee;
		padding-bottom: 8px;
	}

	.tab-button {
		padding: 8px 16px;
		border: 1px solid #ddd;
		background: #f8f9fa;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.2s;
		font-size: 0.9rem;
	}

	.tab-button:hover {
		background: #e9ecef;
		border-color: #adb5bd;
	}

	.tab-button.active {
		background: #667eea;
		color: white;
		border-color: #667eea;
	}

	.code-container {
		position: relative;
		margin-bottom: 24px;
	}

	.code-container pre {
		background: #2d3748;
		color: #e2e8f0;
		padding: 16px;
		border-radius: 8px;
		overflow-x: auto;
		font-size: 0.9rem;
		line-height: 1.5;
		margin: 0;
		font-family: 'Fira Code', 'Monaco', 'Courier New', monospace;
	}

	.code-container code {
		font-family: inherit;
	}

	.btn-copy-code {
		position: absolute;
		top: 8px;
		right: 8px;
		background: rgba(0, 0, 0, 0.7);
		color: white;
		border: none;
		padding: 6px 12px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.8rem;
		transition: background 0.2s;
	}

	.btn-copy-code:hover {
		background: rgba(0, 0, 0, 0.9);
	}

	.code-info {
		background: #f8f9fa;
		padding: 16px;
		border-radius: 8px;
		border-left: 4px solid #667eea;
	}

	.code-info h4 {
		margin: 0 0 12px 0;
		color: #333;
		font-size: 1.1rem;
	}
</style>
